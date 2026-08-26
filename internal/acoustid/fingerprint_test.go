package acoustid

import (
	"errors"
	"path/filepath"
	"testing"
)

// TestComputeUnavailableWhenFpcalcMissing forces PATH to an empty directory
// so the test is deterministic regardless of whether the host running it
// has chromaprint installed.
func TestComputeUnavailableWhenFpcalcMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := Compute(filepath.Join(t.TempDir(), "whatever.mp3"))
	if !errors.Is(err, ErrFpcalcUnavailable) {
		t.Errorf("Compute with no fpcalc on PATH: err = %v, want ErrFpcalcUnavailable", err)
	}
}
