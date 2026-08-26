package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/library"
)

func seedLibrary(t *testing.T, tracks ...audio.Track) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "library.db")
	db, err := library.Open(dbPath)
	if err != nil {
		t.Fatalf("library.Open: %v", err)
	}
	defer func() { _ = db.Close() }()
	for _, tr := range tracks {
		if err := db.Upsert(tr); err != nil {
			t.Fatalf("Upsert(%s): %v", tr.Path, err)
		}
	}
	return dbPath
}

func TestQueryReturnsMatchingTracks(t *testing.T) {
	dbPath := seedLibrary(t, audio.Track{
		Path: "/music/bowie/ziggy/04 starman.flac", Title: "Starman",
		Artist: "David Bowie", AlbumArtist: "David Bowie", Album: "Ziggy Stardust", Year: 1972, TrackNum: 4,
	})

	got := runRoot(t, "query", "artist:David Bowie", "year:1972", "--library-db", dbPath)
	for _, want := range []string{"Starman", "David Bowie", "Ziggy Stardust", "1972"} {
		if !strings.Contains(got, want) {
			t.Errorf("query output missing %q:\n%s", want, got)
		}
	}
}

func TestQueryNoMatchesReportsMessage(t *testing.T) {
	dbPath := seedLibrary(t, audio.Track{Path: "/m/t.mp3", Artist: "Someone"})

	got := runRoot(t, "query", "artist:Nobody", "--library-db", dbPath)
	if !strings.Contains(got, "no tracks matched") {
		t.Errorf("query output = %q, want a no-match message", got)
	}
}

func TestQueryRejectsMalformedClause(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"query", "not-a-clause"})
	if err := root.Execute(); err == nil {
		t.Error("query with a malformed clause: got nil error, want an error")
	}
}

func TestQueryRequiresAtLeastOneClause(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"query"})
	if err := root.Execute(); err == nil {
		t.Error("query with no clauses: got nil error, want an error")
	}
}
