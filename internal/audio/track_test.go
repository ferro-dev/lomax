package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ferro-dev/lomax/internal/testsupport"
)

func TestReadTrackParsesID3v2Tags(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "fixture.mp3", map[string]string{
		"TIT2": "Sea of Love",
		"TPE1": "Cat Power",
		"TALB": "Jukebox",
		"TPE2": "Cat Power",
		"TYER": "2008",
		"TRCK": "3/12",
		"TPOS": "1/1",
		"TCON": "Rock",
	})

	track, err := ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack: %v", err)
	}

	want := Track{
		Path:        path,
		Title:       "Sea of Love",
		Artist:      "Cat Power",
		Album:       "Jukebox",
		AlbumArtist: "Cat Power",
		Genre:       "Rock",
		Year:        2008,
		TrackNum:    3,
		TrackTotal:  12,
		DiscNum:     1,
		DiscTotal:   1,
	}

	if track.Title != want.Title || track.Artist != want.Artist || track.Album != want.Album {
		t.Errorf("core fields = %+v, want title/artist/album %+v", track, want)
	}
	if track.AlbumArtist != want.AlbumArtist {
		t.Errorf("AlbumArtist = %q, want %q", track.AlbumArtist, want.AlbumArtist)
	}
	if track.Genre != want.Genre {
		t.Errorf("Genre = %q, want %q", track.Genre, want.Genre)
	}
	if track.Year != want.Year {
		t.Errorf("Year = %d, want %d", track.Year, want.Year)
	}
	if track.TrackNum != want.TrackNum || track.TrackTotal != want.TrackTotal {
		t.Errorf("Track = %d/%d, want %d/%d", track.TrackNum, track.TrackTotal, want.TrackNum, want.TrackTotal)
	}
	if track.DiscNum != want.DiscNum || track.DiscTotal != want.DiscTotal {
		t.Errorf("Disc = %d/%d, want %d/%d", track.DiscNum, track.DiscTotal, want.DiscNum, want.DiscTotal)
	}
	if track.Size == 0 {
		t.Error("Size = 0, want the fixture file's actual size")
	}
}

func TestReadTrackErrorsOnNonAudioFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("just some text, no tags here"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if _, err := ReadTrack(path); err == nil {
		t.Error("ReadTrack on a non-audio file: got nil error, want an error")
	}
}

func TestReadTrackErrorsOnMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.mp3")
	if _, err := ReadTrack(path); err == nil {
		t.Error("ReadTrack on a missing file: got nil error, want an error")
	}
}
