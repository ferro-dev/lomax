// Package xdg resolves the XDG Base Directory paths lomax uses throughout
// (see docs/music-cli-plan.md section 7). It only computes the base
// directories; callers append their own "lomax/..." subpath and handle
// their own env-var overrides (e.g. LOMAX_STATE_DIR), since those vary by
// caller.
package xdg

import (
	"fmt"
	"os"
	"path/filepath"
)

// StateDir returns $XDG_STATE_HOME, or $HOME/.local/state if unset.
func StateDir() (string, error) {
	return dir("XDG_STATE_HOME", ".local", "state")
}

// DataDir returns $XDG_DATA_HOME, or $HOME/.local/share if unset.
func DataDir() (string, error) {
	return dir("XDG_DATA_HOME", ".local", "share")
}

func dir(xdgVar string, fallback ...string) (string, error) {
	if v := os.Getenv(xdgVar); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("xdg: determine home directory: %w", err)
	}
	return filepath.Join(append([]string{home}, fallback...)...), nil
}
