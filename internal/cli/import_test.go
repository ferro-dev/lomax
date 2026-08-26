package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/resolve"
	"github.com/ferro-dev/lomax/internal/testsupport"
)

func TestBitrateClass(t *testing.T) {
	cases := map[string]string{"FLAC": "lossless", "DSF": "lossless", "ALAC": "lossless", "MP3": "lossy", "M4A": "lossy", "OGG": "lossy"}
	for fileType, want := range cases {
		if got := bitrateClass(fileType); got != want {
			t.Errorf("bitrateClass(%q) = %q, want %q", fileType, got, want)
		}
	}
}

func TestFieldOrCurrent(t *testing.T) {
	if got := fieldOrCurrent("proposed", "current"); got != "proposed" {
		t.Errorf("fieldOrCurrent with a proposed value = %q, want %q", got, "proposed")
	}
	if got := fieldOrCurrent("", "current"); got != "current" {
		t.Errorf("fieldOrCurrent with no proposed value = %q, want %q", got, "current")
	}
}

func TestDestinationFieldsFallsBackToCurrentTagsWithNoProposal(t *testing.T) {
	track := audio.Track{Artist: "Artist", Album: "Album", Title: "Title", Year: 1999, FileType: "FLAC"}
	fields := destinationFields(track, nil, "/library/track.flac")

	if fields.Artist != "Artist" || fields.Album != "Album" || fields.Title != "Title" || fields.Year != 1999 {
		t.Errorf("destinationFields with no proposal = %+v, want the track's own tags", fields)
	}
	if fields.Format != "FLAC" || fields.Ext != "flac" || fields.BitrateClass != "lossless" {
		t.Errorf("destinationFields format/ext/bitrate = %+v, unexpected", fields)
	}
}

func TestDestinationFieldsPrefersProposalOverCurrentTags(t *testing.T) {
	track := audio.Track{Artist: "Old Artist", Year: 1990}
	proposal := &resolve.Proposal{Artist: "New Artist", Year: 2020}

	fields := destinationFields(track, proposal, "/library/track.mp3")
	if fields.Artist != "New Artist" || fields.Year != 2020 {
		t.Errorf("destinationFields = %+v, want the proposal's values to win", fields)
	}
}

// TestImportWithNoTagsStillOrganizesByFallback exercises the fully offline
// path: a file with no tags at all has nothing for MusicBrainz to search
// with and no AcoustID key is configured, so import falls back entirely to
// the file's own (empty) tags for naming. No network access occurs.
func TestImportWithNoTagsStillOrganizesByFallback(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "untagged.mp3", map[string]string{})
	dest := t.TempDir()

	got := runRoot(t, "import", path, "--dest", dest)
	if !strings.Contains(got, "no metadata match found") {
		t.Errorf("import output missing no-match message:\n%s", got)
	}
	if !strings.Contains(got, dest) {
		t.Errorf("import output missing a destination path under %s:\n%s", dest, got)
	}
}

func TestImportRequiresDest(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"import", t.TempDir()})
	if err := root.Execute(); err == nil {
		t.Error("import with no --dest: got nil error, want an error")
	}
}

func TestImportMissingSourceReturnsError(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"import", filepath.Join(t.TempDir(), "does-not-exist"), "--dest", t.TempDir()})
	if err := root.Execute(); err == nil {
		t.Error("import on a missing source: got nil error, want an error")
	}
}

func TestImportRejectsBadNamingTemplate(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"import", t.TempDir(), "--dest", t.TempDir(), "--naming-template", "{nonsense}"})
	if err := root.Execute(); err == nil {
		t.Error("import with an unknown template field: got nil error, want an error")
	}
}
