package cli

import (
	"bytes"
	"strings"
	"testing"
)

// runRoot executes the root command with the given args, capturing combined
// stdout/stderr.
func runRoot(t *testing.T, args ...string) string {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return out.String()
}

func TestAboutShowsVersionAndAttribution(t *testing.T) {
	got := runRoot(t, "about")
	if !strings.Contains(got, "Alan Lomax") {
		t.Errorf("about output missing Alan Lomax attribution: %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Errorf("about output missing version %q: %q", Version, got)
	}
}

func TestVersionFlagIncludesAttribution(t *testing.T) {
	got := runRoot(t, "--version")
	if !strings.Contains(got, "Alan Lomax") {
		t.Errorf("--version output missing attribution: %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Errorf("--version output missing version %q: %q", Version, got)
	}
}

// TestAttributionMentionsNonAffiliation guards the disambiguation requirement:
// the credit must state the project is independent / not affiliated.
func TestAttributionMentionsNonAffiliation(t *testing.T) {
	if !strings.Contains(Attribution, "not affiliated") {
		t.Errorf("attribution must state non-affiliation: %q", Attribution)
	}
}
