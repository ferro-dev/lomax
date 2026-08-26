package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/library"
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

// TestImportAppliedRegistersInLibrary exercises a real (non-dry-run) import
// end to end. The fixture has a title/album but no artist, so
// resolveFile's "needs both artist and title" MusicBrainz gate skips it —
// fully offline — while still giving the naming template enough to work
// with for a normal-looking destination path.
func TestImportAppliedRegistersInLibrary(t *testing.T) {
	src := testsupport.WriteID3v2Fixture(t, "track.mp3", map[string]string{
		"TIT2": "Test Title", "TALB": "Test Album",
	})
	dest := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "library.db")

	got := runRoot(t, "import", src, "--dest", dest, "--dry-run=false", "--library-db", dbPath)
	if !strings.Contains(got, "imported to") {
		t.Errorf("import output = %q, want it to report the import", got)
	}

	db, err := library.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen library: %v", err)
	}
	defer func() { _ = db.Close() }()
	all, err := db.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("database after import = %+v, want exactly one registered track", all)
	}
	if !strings.HasPrefix(all[0].Path, dest) {
		t.Errorf("registered path %q is not under the destination %q", all[0].Path, dest)
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
