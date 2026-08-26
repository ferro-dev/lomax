package audio

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestProbeUnavailableWhenFFprobeMissing forces PATH to an empty directory
// so the test is deterministic regardless of whether the host running it
// has ffmpeg installed.
func TestProbeUnavailableWhenFFprobeMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := Probe(filepath.Join(t.TempDir(), "whatever.mp3"))
	if !errors.Is(err, ErrProbeUnavailable) {
		t.Errorf("Probe with no ffprobe on PATH: err = %v, want ErrProbeUnavailable", err)
	}
}
