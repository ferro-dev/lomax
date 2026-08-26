package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/testsupport"
)

// TestResolveWithNoTagsReportsNoMatch exercises the fully offline path: a
// file with no artist/title tags skips MusicBrainz (nothing to search
// with), and no AcoustID key is configured so that source is never
// constructed either. No network access occurs.
func TestResolveWithNoTagsReportsNoMatch(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "untagged.mp3", map[string]string{})

	got := runRoot(t, "resolve", path)
	if !strings.Contains(got, "no metadata match found") {
		t.Errorf("resolve output = %q, want a no-match message", got)
	}
}

func TestResolveMissingPathReturnsError(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"resolve", filepath.Join(t.TempDir(), "does-not-exist")})
	if err := root.Execute(); err == nil {
		t.Error("resolve on a missing path: got nil error, want an error")
	}
}

func TestResolveDryRunFalseIsRejected(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"resolve", "--dry-run=false", t.TempDir()})
	err := root.Execute()
	if err == nil {
		t.Fatal("resolve --dry-run=false: got nil error, want an error")
	}
	if !strings.Contains(err.Error(), "Milestone 3") {
		t.Errorf("resolve --dry-run=false error = %q, want it to explain writing isn't supported yet", err.Error())
	}
}
