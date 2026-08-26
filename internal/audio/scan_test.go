package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsAudioFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"track.mp3", true},
		{"track.MP3", true},
		{"track.flac", true},
		{"track.m4a", true},
		{"track.ogg", true},
		{"track.oga", true},
		{"track.dsf", true},
		{"cover.jpg", false},
		{"readme.txt", false},
		{"noextension", false},
	}
	for _, c := range cases {
		if got := IsAudioFile(c.path); got != c.want {
			t.Errorf("IsAudioFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestScanFindsAudioFilesRecursively(t *testing.T) {
	root := t.TempDir()
	makeFile := func(rel string) {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	makeFile("Artist/Album/01 Title.mp3")
	makeFile("Artist/Album/02 Title.flac")
	makeFile("Artist/Album/cover.jpg")
	makeFile("Artist/Album/notes.txt")
	makeFile("Other/track.m4a")

	got, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	want := []string{
		filepath.Join(root, "Artist/Album/01 Title.mp3"),
		filepath.Join(root, "Artist/Album/02 Title.flac"),
		filepath.Join(root, "Other/track.m4a"),
	}
	if len(got) != len(want) {
		t.Fatalf("Scan returned %d files, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Scan()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScanOnSingleFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "track.mp3")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 1 || got[0] != path {
		t.Errorf("Scan(%q) = %v, want [%q]", path, got, path)
	}
}
