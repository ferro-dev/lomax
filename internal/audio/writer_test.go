package audio

import (
	"os"
	"path/filepath"
	"testing"

	flacvorbis "github.com/go-flac/flacvorbis/v2"
	goflac "github.com/go-flac/go-flac/v2"

	"github.com/ferro-dev/lomax/internal/testsupport"
)

func TestWriteTagsMP3RoundTrips(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "track.mp3", map[string]string{
		"TIT2": "Old Title",
		"TPE1": "Old Artist",
	})

	err := WriteTags(path, WritableTags{
		Title:       "New Title",
		Artist:      "New Artist",
		Album:       "New Album",
		AlbumArtist: "New Album Artist",
		Year:        2020,
	})
	if err != nil {
		t.Fatalf("WriteTags: %v", err)
	}

	track, err := ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack after write: %v", err)
	}
	if track.Title != "New Title" || track.Artist != "New Artist" || track.Album != "New Album" ||
		track.AlbumArtist != "New Album Artist" || track.Year != 2020 {
		t.Errorf("track after WriteTags = %+v, unexpected", track)
	}
}

func TestWriteTagsMP3OnlyUpdatesGivenFields(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "track.mp3", map[string]string{
		"TIT2": "Keep Me",
		"TPE1": "Original Artist",
	})

	if err := WriteTags(path, WritableTags{Artist: "Updated Artist"}); err != nil {
		t.Fatalf("WriteTags: %v", err)
	}

	track, err := ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack: %v", err)
	}
	if track.Title != "Keep Me" {
		t.Errorf("Title = %q, want unchanged %q", track.Title, "Keep Me")
	}
	if track.Artist != "Updated Artist" {
		t.Errorf("Artist = %q, want %q", track.Artist, "Updated Artist")
	}
}

func TestWriteTagsFLACRoundTrips(t *testing.T) {
	path := testsupport.WriteMinimalFLACFixture(t, "track.flac")

	err := WriteTags(path, WritableTags{
		Title:       "Sea of Love",
		Artist:      "Cat Power",
		Album:       "Jukebox",
		AlbumArtist: "Cat Power",
		Year:        2008,
	})
	if err != nil {
		t.Fatalf("WriteTags: %v", err)
	}

	track, err := ReadTrack(path)
	if err != nil {
		t.Fatalf("ReadTrack after write: %v", err)
	}
	if track.Title != "Sea of Love" || track.Artist != "Cat Power" || track.Album != "Jukebox" ||
		track.AlbumArtist != "Cat Power" || track.Year != 2008 {
		t.Errorf("track after WriteTags = %+v, unexpected", track)
	}
}

func TestWriteTagsFLACOverwriteDoesNotDuplicateFields(t *testing.T) {
	path := testsupport.WriteMinimalFLACFixture(t, "track.flac")

	if err := WriteTags(path, WritableTags{Title: "First"}); err != nil {
		t.Fatalf("first WriteTags: %v", err)
	}
	if err := WriteTags(path, WritableTags{Title: "Second"}); err != nil {
		t.Fatalf("second WriteTags: %v", err)
	}

	f, err := goflac.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	defer func() { _ = f.Close() }()

	var comment *flacvorbis.MetaDataBlockVorbisComment
	for _, block := range f.Meta {
		if block.Type == goflac.VorbisComment {
			comment, err = flacvorbis.ParseFromMetaDataBlock(*block)
			if err != nil {
				t.Fatalf("parse vorbis comment: %v", err)
			}
		}
	}
	if comment == nil {
		t.Fatal("no VorbisComment block found after two writes")
	}

	titles, err := comment.Get(flacvorbis.FIELD_TITLE)
	if err != nil {
		t.Fatalf("Get(TITLE): %v", err)
	}
	if len(titles) != 1 || titles[0] != "Second" {
		t.Errorf("TITLE comments = %v, want exactly [\"Second\"]", titles)
	}
}

func TestWriteTagsRejectsUnsupportedFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.ogg")
	if err := os.WriteFile(path, []byte("not really ogg"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	err := WriteTags(path, WritableTags{Title: "Anything"})
	if err == nil {
		t.Error("WriteTags on an .ogg file: got nil error, want an error")
	}
}

func TestWriteTagsZeroFieldsIsNoOp(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "track.mp3", map[string]string{"TIT2": "Unchanged"})

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	if err := WriteTags(path, WritableTags{}); err != nil {
		t.Fatalf("WriteTags with no fields set: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after no-op WriteTags: %v", err)
	}
	if string(before) != string(after) {
		t.Error("WriteTags with all-zero fields modified the file")
	}
}
