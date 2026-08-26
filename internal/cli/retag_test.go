package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/testsupport"
)

// TestRetagWithNoTagsWritesNothing exercises the fully offline path: a file
// with no tags has nothing for MusicBrainz to search with and no AcoustID
// key is configured, so there's no proposal and nothing to write. No
// network access occurs.
func TestRetagWithNoTagsWritesNothing(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "untagged.mp3", map[string]string{})
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	got := runRoot(t, "retag", path, "--dry-run=false")
	if !strings.Contains(got, "no metadata match found") {
		t.Errorf("retag output missing no-match message:\n%s", got)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after retag: %v", err)
	}
	if string(before) != string(after) {
		t.Error("retag modified a file with no resolvable proposal")
	}
}

func TestRetagMissingPathReturnsError(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"retag", filepath.Join(t.TempDir(), "does-not-exist")})
	if err := root.Execute(); err == nil {
		t.Error("retag on a missing path: got nil error, want an error")
	}
}
