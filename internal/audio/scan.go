// Package audio reads audio file tags and stream properties for lomax's
// import, inspect, and organize workflows.
package audio

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// supportedExt are the file extensions Scan considers audio files. This
// mirrors the formats github.com/dhowden/tag can parse (MP3, MP4/M4A, FLAC,
// OGG, DSF); WAV/AIFF are not tag-bearing formats in the same sense and are
// deliberately excluded here.
var supportedExt = map[string]bool{
	".mp3":  true,
	".m4a":  true,
	".m4b":  true,
	".m4p":  true,
	".mp4":  true,
	".flac": true,
	".ogg":  true,
	".oga":  true,
	".dsf":  true,
}

// IsAudioFile reports whether path has an extension lomax recognizes as an
// audio file, by extension only (case-insensitive). It does not open the
// file or validate its contents.
func IsAudioFile(path string) bool {
	return supportedExt[strings.ToLower(filepath.Ext(path))]
}

// Scan walks root recursively and returns the paths of all recognized audio
// files, sorted for deterministic output. root may be a single file, in
// which case Scan returns just that path if it looks like an audio file.
func Scan(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if IsAudioFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
