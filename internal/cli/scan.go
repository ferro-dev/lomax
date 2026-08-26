package cli

import (
	"fmt"
	"os"

	"github.com/ferro-dev/lomax/internal/audio"
)

// scanPathForAudio resolves path to the list of audio files it names: just
// path itself if it's a file, or every recognized audio file under it if
// it's a directory. op is the calling command's name, used to prefix
// errors (e.g. "inspect", "resolve").
func scanPathForAudio(op, path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", op, path, err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	files, err := audio.Scan(path)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", op, path, err)
	}
	return files, nil
}
