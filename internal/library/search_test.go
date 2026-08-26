package library

import (
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/query"
)

func seedSearchFixtures(t *testing.T, db *DB) {
	t.Helper()
	tracks := []audio.Track{
		{Path: "/m/bowie-1972-a.flac", Title: "Starman", Artist: "David Bowie", AlbumArtist: "David Bowie", Album: "Ziggy Stardust", Year: 1972, TrackNum: 4},
		{Path: "/m/bowie-1972-b.flac", Title: "Five Years", Artist: "David Bowie", AlbumArtist: "David Bowie", Album: "Ziggy Stardust", Year: 1972, TrackNum: 1},
		{Path: "/m/bowie-1977.flac", Title: "Sound and Vision", Artist: "David Bowie", AlbumArtist: "David Bowie", Album: "Low", Year: 1977, TrackNum: 3},
		{Path: "/m/other.flac", Title: "Some Song", Artist: "Someone Else", Album: "Whatever", Year: 1972, TrackNum: 1},
	}
	for _, tr := range tracks {
		if err := db.Upsert(tr); err != nil {
			t.Fatalf("seed Upsert(%s): %v", tr.Path, err)
		}
	}
}

func TestSearchByArtistAndYear(t *testing.T) {
	db := openTestDB(t)
	seedSearchFixtures(t, db)

	q, err := query.Parse([]string{"artist:David Bowie", "year:1972"})
	if err != nil {
		t.Fatalf("query.Parse: %v", err)
	}
	got, err := db.Search(q)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Search() returned %d rows, want 2", len(got))
	}
	for _, r := range got {
		if r.Artist != "David Bowie" || r.Year != 1972 {
			t.Errorf("Search() row = %+v, want David Bowie / 1972", r)
		}
	}
}

func TestSearchByAlbum(t *testing.T) {
	db := openTestDB(t)
	seedSearchFixtures(t, db)

	q, err := query.Parse([]string{"album:Low"})
	if err != nil {
		t.Fatalf("query.Parse: %v", err)
	}
	got, err := db.Search(q)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 || got[0].Title != "Sound and Vision" {
		t.Errorf("Search(album:Low) = %+v, want just Sound and Vision", got)
	}
}

func TestSearchWithNoClausesMatchingReturnsEmpty(t *testing.T) {
	db := openTestDB(t)
	seedSearchFixtures(t, db)

	q, err := query.Parse([]string{"artist:Nobody"})
	if err != nil {
		t.Fatalf("query.Parse: %v", err)
	}
	got, err := db.Search(q)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Search(artist:Nobody) returned %d rows, want 0", len(got))
	}
}

func TestSearchRejectsNonNumericYear(t *testing.T) {
	db := openTestDB(t)
	q := &query.Query{Clauses: []query.Clause{{Field: "year", Value: "not-a-year"}}}
	if _, err := db.Search(q); err == nil {
		t.Error("Search with a non-numeric year clause: got nil error, want an error")
	}
}
