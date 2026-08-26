package pluginhost

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// buildTestPlugin compiles plugins/<name> (a real first-party plugin
// module in this monorepo) into a fixture binary named
// "lomax-plugin-<name>" under t.TempDir(), and returns its Descriptor.
// Building the actual plugin — rather than a synthetic stub — exercises
// the real go-plugin handshake end to end: process launch, protocol
// negotiation, gRPC dial, and Dispense.
func buildTestPlugin(t *testing.T, name string) Descriptor {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file location")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	pluginDir := filepath.Join(repoRoot, "plugins", name)

	binName := binaryPrefix + name
	if runtime.GOOS == "windows" {
		// Go's os/exec refuses to run a file with no recognized extension
		// on Windows even given a full path — a Windows-only quirk, since
		// lomax's actual plugin-naming convention (no extension) is a
		// Linux-only concern in production.
		binName += ".exe"
	}
	binPath := filepath.Join(t.TempDir(), binName)

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = pluginDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build plugins/%s: %v\n%s", name, err, out)
	}

	return Descriptor{Name: name, Path: binPath}
}

func TestLoadHandshakesWithARealPlugin(t *testing.T) {
	d := buildTestPlugin(t, "discogs")

	handle, err := Load(d)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer handle.Close()

	if handle.Name != "discogs" {
		t.Errorf("Name = %q, want %q", handle.Name, "discogs")
	}

	// No LOMAX_DISCOGS_TOKEN is set for this test process (and go-plugin
	// spawns the child inheriting this process's environment), so the
	// plugin's own "unconfigured" behavior applies: a clean no-match
	// rather than an error.
	match, err := handle.Resolver.ResolveMetadata(context.Background(), "Some Artist", "Some Album", "Some Title")
	if err != nil {
		t.Fatalf("ResolveMetadata: %v", err)
	}
	if match != nil {
		t.Errorf("ResolveMetadata with no token configured = %+v, want nil", match)
	}
}

func TestLoadRejectsAMissingBinary(t *testing.T) {
	_, err := Load(Descriptor{Name: "nonexistent", Path: filepath.Join(t.TempDir(), "does-not-exist")})
	if err == nil {
		t.Error("Load with a missing binary: got nil error, want an error")
	}
}
