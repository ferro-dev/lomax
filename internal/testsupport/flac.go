package testsupport

import (
	"os"
	"path/filepath"
	"testing"

	goflac "github.com/go-flac/go-flac/v2"
)

// BuildMinimalFLAC assembles the smallest byte stream go-flac's ParseFile
// will accept: the "fLaC" magic, a STREAMINFO block (mandatory, but
// go-flac treats its 34-byte payload as opaque — it never validates the
// bit-packed sample-rate/channel fields against real audio), and a
// synthetic "frame" section starting with a valid FLAC frame sync code
// (0xFF 0xF8), which is all checkFLACStream requires since go-flac never
// decodes frame contents. This produces a real, parseable FLAC stream
// without a full encoder or a checked-in binary fixture.
func BuildMinimalFLAC(padBytes int) []byte {
	streamInfo := &goflac.MetaDataBlock{Type: goflac.StreamInfo, Data: make([]byte, 34)}

	out := []byte("fLaC")
	out = append(out, streamInfo.Marshal(true)...)
	out = append(out, 0xFF, 0xF8) // FLAC frame sync code
	out = append(out, make([]byte, padBytes)...)
	return out
}

// WriteMinimalFLACFixture writes BuildMinimalFLAC's output to a file named
// name inside t.TempDir(), returning the file's path.
func WriteMinimalFLACFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, BuildMinimalFLAC(128), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return path
}
