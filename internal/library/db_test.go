package library

import (
	"path/filepath"
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
)

// openTestDB opens a fresh library database under t.TempDir(), running
// migrations against a real (in-process, pure-Go) SQLite file.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestOpenRunsMigrations(t *testing.T) {
	db := openTestDB(t)

	for _, table := range []string{"artists", "albums", "tracks"} {
		var name string
		err := db.sql.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", table, err)
		}
	}
}

func TestUpsertInsertsThenUpdatesInPlace(t *testing.T) {
	db := openTestDB(t)

	track := audio.Track{
		Path: "/music/artist/album/01 title.flac", Title: "Old Title", Artist: "Artist",
		Album: "Album", AlbumArtist: "Artist", Year: 1999, TrackNum: 1, TrackTotal: 10,
		DiscNum: 1, DiscTotal: 1, Size: 12345,
	}
	if err := db.Upsert(track); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	all, err := db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("All() returned %d rows, want 1", len(all))
	}
	if all[0].Title != "Old Title" || all[0].Artist != "Artist" || all[0].Year != 1999 {
		t.Errorf("row after insert = %+v, unexpected", all[0])
	}

	track.Title = "New Title"
	track.Year = 2020
	if err := db.Upsert(track); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	all, err = db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("All() after update returned %d rows, want 1 (same path should update, not duplicate)", len(all))
	}
	if all[0].Title != "New Title" || all[0].Year != 2020 {
		t.Errorf("row after update = %+v, want Title=New Title Year=2020", all[0])
	}
}

func TestUpsertHandlesEmptyArtistAndAlbum(t *testing.T) {
	db := openTestDB(t)

	if err := db.Upsert(audio.Track{Path: "/music/untagged.mp3"}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	all, err := db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("All() returned %d rows, want 1", len(all))
	}
	if all[0].Artist != "" || all[0].Album != "" || all[0].AlbumArtist != "" {
		t.Errorf("row for an untagged track = %+v, want empty artist/album/album artist", all[0])
	}
}

func TestDeleteRemovesTrack(t *testing.T) {
	db := openTestDB(t)
	track := audio.Track{Path: "/music/track.mp3", Title: "T"}
	if err := db.Upsert(track); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	if err := db.Delete(track.Path); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	all, err := db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("All() after Delete returned %d rows, want 0", len(all))
	}
}

func TestDeleteOfUntrackedPathIsNotAnError(t *testing.T) {
	db := openTestDB(t)
	if err := db.Delete("/music/never-tracked.mp3"); err != nil {
		t.Errorf("Delete on an untracked path: %v, want nil", err)
	}
}

func TestUpdatePathMovesRow(t *testing.T) {
	db := openTestDB(t)
	track := audio.Track{Path: "/old/track.mp3", Title: "T"}
	if err := db.Upsert(track); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	if err := db.UpdatePath("/old/track.mp3", "/new/track.mp3"); err != nil {
		t.Fatalf("UpdatePath: %v", err)
	}

	all, err := db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 || all[0].Path != "/new/track.mp3" {
		t.Errorf("All() after UpdatePath = %+v, want a single row at /new/track.mp3", all)
	}
}

func TestUpdatePathOnMissingRowReturnsError(t *testing.T) {
	db := openTestDB(t)
	if err := db.UpdatePath("/nope.mp3", "/also-nope.mp3"); err == nil {
		t.Error("UpdatePath on a path with no track: got nil error, want an error")
	}
}
