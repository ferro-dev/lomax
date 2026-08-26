package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ferro-dev/lomax/internal/testsupport"
)

func TestInspectSingleFileShowsTags(t *testing.T) {
	path := testsupport.WriteID3v2Fixture(t, "track.mp3", map[string]string{
		"TIT2": "Sea of Love",
		"TPE1": "Cat Power",
		"TALB": "Jukebox",
	})

	got := runRoot(t, "inspect", path)

	for _, want := range []string{"Sea of Love", "Cat Power", "Jukebox"} {
		if !strings.Contains(got, want) {
			t.Errorf("inspect output missing %q:\n%s", want, got)
		}
	}
}

func TestInspectDirectoryListsEveryTrack(t *testing.T) {
	dir := t.TempDir()
	writeInto := func(name string, frames map[string]string) {
		data := testsupport.BuildID3v2Tag(frames, 128)
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	writeInto("one.mp3", map[string]string{"TIT2": "First Track", "TPE1": "Artist A"})
	writeInto("two.mp3", map[string]string{"TIT2": "Second Track", "TPE1": "Artist B"})
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("skip me"), 0o644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}

	got := runRoot(t, "inspect", dir)

	for _, want := range []string{"First Track", "Artist A", "Second Track", "Artist B"} {
		if !strings.Contains(got, want) {
			t.Errorf("inspect output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "skip me") {
		t.Errorf("inspect output should not reflect non-audio files:\n%s", got)
	}
}

func TestInspectEmptyDirectoryReportsNoFiles(t *testing.T) {
	got := runRoot(t, "inspect", t.TempDir())
	if !strings.Contains(got, "no audio files found") {
		t.Errorf("inspect on an empty directory: got %q, want a no-audio-files message", got)
	}
}

func TestInspectMissingPathReturnsError(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"inspect", filepath.Join(t.TempDir(), "does-not-exist")})
	if err := root.Execute(); err == nil {
		t.Error("inspect on a missing path: got nil error, want an error")
	}
}
