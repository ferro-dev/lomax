package pluginhost

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o755); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
}

func TestDiscoverFindsPrefixedBinaries(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "lomax-plugin-discogs")
	touch(t, dir, "lomax-plugin-lastfm")
	touch(t, dir, "not-a-plugin")

	got := Discover(dir, nil)
	if len(got) != 2 {
		t.Fatalf("Discover() = %+v, want 2 entries", got)
	}
	if got[0].Name != "discogs" || got[1].Name != "lastfm" {
		t.Errorf("Discover() names = [%s, %s], want [discogs, lastfm] (sorted)", got[0].Name, got[1].Name)
	}
}

func TestDiscoverPreferPluginDirOverExtraDirs(t *testing.T) {
	pluginDir := t.TempDir()
	extraDir := t.TempDir()
	touch(t, pluginDir, "lomax-plugin-discogs")
	touch(t, extraDir, "lomax-plugin-discogs")

	got := Discover(pluginDir, []string{extraDir})
	if len(got) != 1 {
		t.Fatalf("Discover() = %+v, want exactly 1 (deduped) entry", got)
	}
	if got[0].Path != filepath.Join(pluginDir, "lomax-plugin-discogs") {
		t.Errorf("Discover() path = %s, want the plugin-dir copy to win over the extra-dir one", got[0].Path)
	}
}

func TestDiscoverSkipsMissingDirectories(t *testing.T) {
	got := Discover(filepath.Join(t.TempDir(), "does-not-exist"), nil)
	if len(got) != 0 {
		t.Errorf("Discover() on a missing directory = %+v, want empty", got)
	}
}

func TestDiscoverIgnoresDirectoriesNamedLikeAPlugin(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "lomax-plugin-notabinary"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got := Discover(dir, nil)
	if len(got) != 0 {
		t.Errorf("Discover() = %+v, want directories to be skipped", got)
	}
}
