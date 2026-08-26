// Package testsupport holds fixture builders shared by tests across lomax's
// internal packages. It is not part of the public API surface.
package testsupport

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// BuildID3v2Tag hand-assembles a minimal ID3v2.3 tag containing the given
// text frames, followed by padBytes of zero-filled "audio" payload. This
// produces a real, spec-correct ID3v2.3 stream without needing a checked-in
// binary fixture or a full MP3 encoder — github.com/dhowden/tag detects and
// reads MP3 tags via the "ID3" magic + header alone and never validates the
// MPEG frames that follow.
//
// frames maps a 4-character frame ID (e.g. "TIT2") to its text content.
func BuildID3v2Tag(frames map[string]string, padBytes int) []byte {
	var body bytes.Buffer
	for id, value := range frames {
		var frame bytes.Buffer
		frame.WriteByte(0x00) // text encoding: ISO-8859-1
		frame.WriteString(value)

		body.WriteString(id)
		size := make([]byte, 4)
		binary.BigEndian.PutUint32(size, uint32(frame.Len()))
		body.Write(size)
		body.Write([]byte{0x00, 0x00}) // flags
		body.Write(frame.Bytes())
	}

	var out bytes.Buffer
	out.WriteString("ID3")
	out.Write([]byte{0x03, 0x00}) // version 2.3.0
	out.WriteByte(0x00)           // flags
	out.Write(syncsafe(uint32(body.Len())))
	out.Write(body.Bytes())
	out.Write(make([]byte, padBytes))
	return out.Bytes()
}

// syncsafe encodes n as an ID3v2 syncsafe 4-byte integer: 7 significant bits
// per byte, high bit always zero.
func syncsafe(n uint32) []byte {
	return []byte{
		byte((n >> 21) & 0x7F),
		byte((n >> 14) & 0x7F),
		byte((n >> 7) & 0x7F),
		byte(n & 0x7F),
	}
}

// WriteID3v2Fixture builds an ID3v2.3 tag via BuildID3v2Tag and writes it to
// a file named name inside t.TempDir(), returning the file's path.
func WriteID3v2Fixture(t *testing.T, name string, frames map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, BuildID3v2Tag(frames, 128), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return path
}
