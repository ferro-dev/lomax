// Package pluginhost is the host side of lomax's plugin system: discovering
// plugin binaries and launching them over HashiCorp go-plugin, using the
// wire contract in internal/pluginapi. See docs/music-cli-plan.md section 9.
package pluginhost

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ferro-dev/lomax/internal/xdg"
)

// binaryPrefix is the naming convention every lomax plugin binary follows.
const binaryPrefix = "lomax-plugin-"

// PluginPathEnv lists extra directories to search for plugins, colon-
// separated (semicolon on Windows) like $PATH — the pre-config-system
// stand-in for section 9's `plugin_path` config key, same pattern as
// LOMAX_ACOUSTID_API_KEY predating a real config file.
const PluginPathEnv = "LOMAX_PLUGIN_PATH"

// Descriptor is one discovered plugin: its short name (the binary's
// filename with the "lomax-plugin-" prefix stripped) and its path.
type Descriptor struct {
	Name string
	Path string
}

// DefaultPluginDir returns $XDG_DATA_HOME/lomax/plugins — where
// `lomax plugin install` places binaries (see section 9).
func DefaultPluginDir() (string, error) {
	dataDir, err := xdg.DataDir()
	if err != nil {
		return "", fmt.Errorf("pluginhost: determine default plugin directory: %w", err)
	}
	return filepath.Join(dataDir, "lomax", "plugins"), nil
}

// ExtraDirsFromEnv splits PluginPathEnv into directories, or returns nil if
// it's unset.
func ExtraDirsFromEnv() []string {
	v := os.Getenv(PluginPathEnv)
	if v == "" {
		return nil
	}
	return filepath.SplitList(v)
}

// Discover finds every "lomax-plugin-*" file in pluginDir, then extraDirs,
// then $PATH, in that order — matching section 9's documented discovery
// priority. A name found in an earlier directory shadows the same name
// found later, so a user's plugin dir can override one incidentally also
// on $PATH. Missing or unreadable directories are skipped, not an error —
// only $PATH is guaranteed to exist.
func Discover(pluginDir string, extraDirs []string) []Descriptor {
	dirs := append([]string{pluginDir}, extraDirs...)
	dirs = append(dirs, filepath.SplitList(os.Getenv("PATH"))...)

	seen := make(map[string]bool)
	var found []Descriptor
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, binaryPrefix) {
				continue
			}
			short := strings.TrimPrefix(name, binaryPrefix)
			if seen[short] {
				continue
			}
			seen[short] = true
			found = append(found, Descriptor{Name: short, Path: filepath.Join(dir, name)})
		}
	}

	sort.Slice(found, func(i, j int) bool { return found[i].Name < found[j].Name })
	return found
}
