package audio

import (
	"fmt"
	"os"

	"github.com/dhowden/tag"
)

// Track is lomax's own tag abstraction: a normalized view over the fields
// github.com/dhowden/tag exposes, plus filesystem facts. Keeping this as a
// project-owned struct (rather than passing tag.Metadata around directly)
// means the rest of the codebase depends on one interface even as tag
// reading and, later, tag writing are split across format-specific libraries
// (see docs/music-cli-plan.md section 6).
type Track struct {
	Path string
	Size int64

	FileType tag.FileType
	Format   tag.Format

	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Composer    string
	Genre       string
	Year        int

	TrackNum, TrackTotal int
	DiscNum, DiscTotal   int

	Comment    string
	HasPicture bool
}

// ReadTrack opens path and reads its tags through the project's tag
// abstraction. It returns an error if the file cannot be opened or no
// recognizable tag format is found.
func ReadTrack(path string) (Track, error) {
	f, err := os.Open(path)
	if err != nil {
		return Track{}, fmt.Errorf("read track %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return Track{}, fmt.Errorf("read track %s: %w", path, err)
	}

	m, err := tag.ReadFrom(f)
	if err != nil {
		return Track{}, fmt.Errorf("read track %s: %w", path, err)
	}

	trackNum, trackTotal := m.Track()
	discNum, discTotal := m.Disc()

	return Track{
		Path:        path,
		Size:        info.Size(),
		FileType:    m.FileType(),
		Format:      m.Format(),
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
		Composer:    m.Composer(),
		Genre:       m.Genre(),
		Year:        m.Year(),
		TrackNum:    trackNum,
		TrackTotal:  trackTotal,
		DiscNum:     discNum,
		DiscTotal:   discTotal,
		Comment:     m.Comment(),
		HasPicture:  m.Picture() != nil,
	}, nil
}
