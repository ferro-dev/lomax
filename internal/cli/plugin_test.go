package cli

import (
	"strings"
	"testing"
)

// isolatePluginEnv points plugin discovery/install/remove at a throwaway
// XDG data dir and an empty $PATH, so these tests never see (or interfere
// with) plugins actually installed on the machine running them.
func isolatePluginEnv(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	t.Setenv("LOMAX_PLUGIN_PATH", "")
}

func TestPluginListWithNoneInstalled(t *testing.T) {
	isolatePluginEnv(t)

	got := runRoot(t, "plugin", "list")
	if !strings.Contains(got, "no plugins found") {
		t.Errorf("plugin list output = %q, want a no-plugins message", got)
	}
}

// TestPluginInstallRequiresSourceCheckout relies on `go test`'s working
// directory being internal/cli, which has no plugins/ subdirectory of its
// own — exactly the "not run from a lomax checkout" case install should
// reject.
func TestPluginInstallRequiresSourceCheckout(t *testing.T) {
	isolatePluginEnv(t)

	root := newRootCmd()
	root.SetArgs([]string{"plugin", "install", "discogs"})
	if err := root.Execute(); err == nil {
		t.Error("plugin install with no plugins/ source directory: got nil error, want an error")
	}
}

func TestPluginRemoveNonexistentReturnsError(t *testing.T) {
	isolatePluginEnv(t)

	root := newRootCmd()
	root.SetArgs([]string{"plugin", "remove", "never-installed"})
	if err := root.Execute(); err == nil {
		t.Error("plugin remove of a never-installed plugin: got nil error, want an error")
	}
}
