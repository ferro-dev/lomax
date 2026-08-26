package audio

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis/v2"
	goflac "github.com/go-flac/go-flac/v2"
)

// albumArtistVorbisField is the de facto (non-standardized) Vorbis Comment
// key for album artist, used by beets, Picard, and every other tagger.
const albumArtistVorbisField = "ALBUMARTIST"

// WritableTags is the set of fields WriteTags can update. An empty string
// (or 0 for Year) means "leave this field unchanged" — mirroring
// resolve.Proposal's semantics, since a Proposal is WriteTags' primary
// caller.
//
// Track/disc numbers, genre, and comment are deliberately not writable yet:
// the metadata resolver doesn't propose them either (see
// docs/music-cli-plan.md Milestone 2 notes) and adding write support ahead
// of a source for those values would be dead code.
type WritableTags struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Year        int
}

// IsZero reports whether fields has nothing to write.
func (w WritableTags) IsZero() bool {
	return w.Title == "" && w.Artist == "" && w.Album == "" && w.AlbumArtist == "" && w.Year == 0
}

// WriteTags updates path's tags in place with the non-empty fields in
// fields, dispatching to a format-specific writer by file extension. Only
// MP3 (ID3v2) and FLAC (Vorbis Comments) are supported — MP4/M4A and OGG
// writing are not yet implemented (see docs/music-cli-plan.md section 6:
// the Go tag-writing ecosystem is fragmented and MP4 write support in
// particular needs real design work, deferred past Milestone 3).
func WriteTags(path string, fields WritableTags) error {
	if fields.IsZero() {
		return nil
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		return writeMP3Tags(path, fields)
	case ".flac":
		return writeFLACTags(path, fields)
	default:
		return fmt.Errorf("audio: writing tags to %s files is not yet supported", filepath.Ext(path))
	}
}

func writeMP3Tags(path string, fields WritableTags) error {
	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("audio: open %s for tag writing: %w", path, err)
	}
	defer func() { _ = t.Close() }()

	if fields.Title != "" {
		t.SetTitle(fields.Title)
	}
	if fields.Artist != "" {
		t.SetArtist(fields.Artist)
	}
	if fields.Album != "" {
		t.SetAlbum(fields.Album)
	}
	if fields.AlbumArtist != "" {
		t.AddTextFrame(t.CommonID("Band/Orchestra/Accompaniment"), t.DefaultEncoding(), fields.AlbumArtist)
	}
	if fields.Year != 0 {
		t.SetYear(strconv.Itoa(fields.Year))
	}

	if err := t.Save(); err != nil {
		return fmt.Errorf("audio: save %s: %w", path, err)
	}
	return nil
}

func writeFLACTags(path string, fields WritableTags) error {
	f, err := goflac.ParseFile(path)
	if err != nil {
		return fmt.Errorf("audio: open %s for tag writing: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var comment *flacvorbis.MetaDataBlockVorbisComment
	commentIdx := -1
	for i, block := range f.Meta {
		if block.Type == goflac.VorbisComment {
			comment, err = flacvorbis.ParseFromMetaDataBlock(*block)
			if err != nil {
				return fmt.Errorf("audio: parse existing vorbis comment in %s: %w", path, err)
			}
			commentIdx = i
			break
		}
	}
	if comment == nil {
		comment = flacvorbis.New()
	}

	set := func(key, val string) error {
		if val == "" {
			return nil
		}
		return setVorbisField(comment, key, val)
	}
	if err := set(flacvorbis.FIELD_TITLE, fields.Title); err != nil {
		return fmt.Errorf("audio: set title on %s: %w", path, err)
	}
	if err := set(flacvorbis.FIELD_ARTIST, fields.Artist); err != nil {
		return fmt.Errorf("audio: set artist on %s: %w", path, err)
	}
	if err := set(flacvorbis.FIELD_ALBUM, fields.Album); err != nil {
		return fmt.Errorf("audio: set album on %s: %w", path, err)
	}
	if err := set(albumArtistVorbisField, fields.AlbumArtist); err != nil {
		return fmt.Errorf("audio: set album artist on %s: %w", path, err)
	}
	if fields.Year != 0 {
		if err := set(flacvorbis.FIELD_DATE, strconv.Itoa(fields.Year)); err != nil {
			return fmt.Errorf("audio: set year on %s: %w", path, err)
		}
	}

	marshaled := comment.Marshal()
	if commentIdx >= 0 {
		f.Meta[commentIdx] = &marshaled
	} else {
		f.Meta = append(f.Meta, &marshaled)
	}

	if err := f.Save(path); err != nil {
		return fmt.Errorf("audio: save %s: %w", path, err)
	}
	return nil
}

// setVorbisField replaces any existing value(s) for key with val. Vorbis
// Comment blocks store comments as a flat "KEY=VALUE" string slice with no
// built-in update operation, so an unconditional Add would leave a stale
// duplicate entry behind on every re-tag.
func setVorbisField(c *flacvorbis.MetaDataBlockVorbisComment, key, val string) error {
	filtered := make([]string, 0, len(c.Comments))
	for _, cmt := range c.Comments {
		k, _, found := strings.Cut(cmt, "=")
		if found && strings.EqualFold(k, key) {
			continue
		}
		filtered = append(filtered, cmt)
	}
	c.Comments = filtered
	return c.Add(key, val)
}
