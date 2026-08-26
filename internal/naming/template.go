// Package naming implements lomax's destination-path naming templates:
// user-configurable patterns like "{album_artist}/{year} - {album}/…" that
// the import workflow renders into a relative library path per track. See
// docs/music-cli-plan.md section 8.
package naming

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Default is the naming template used when the user hasn't configured one.
const Default = "{album_artist}/{year} - {album}/{disc:02}-{track:02} {title}.{ext}"

// Fields is the set of values a Template can substitute into a rendered
// path. A zero value for any field (empty string, or 0 for the numeric
// fields) means "unknown" and renders as an empty string — see Render.
type Fields struct {
	Artist       string
	AlbumArtist  string
	Album        string
	Title        string
	Format       string // container/codec type, e.g. "MP3", "FLAC" (Track.FileType)
	Ext          string // destination file extension, e.g. "mp3", "flac" (no leading dot)
	BitrateClass string // "lossless" or "lossy"

	Year  int
	Disc  int
	Track int
}

// intFields are the numeric Fields that support zero-padding via
// "{name:0N}"; every other known field is a plain string substitution.
var intFields = map[string]func(Fields) int{
	"year":  func(f Fields) int { return f.Year },
	"disc":  func(f Fields) int { return f.Disc },
	"track": func(f Fields) int { return f.Track },
}

// stringFields are the plain string-valued template variables.
var stringFields = map[string]func(Fields) string{
	"artist":        func(f Fields) string { return f.Artist },
	"album_artist":  func(f Fields) string { return f.AlbumArtist },
	"album":         func(f Fields) string { return f.Album },
	"title":         func(f Fields) string { return f.Title },
	"format":        func(f Fields) string { return f.Format },
	"ext":           func(f Fields) string { return f.Ext },
	"bitrate_class": func(f Fields) string { return f.BitrateClass },
}

// tokenPattern matches "{name}" or "{name:0N}" (N = zero-pad width).
var tokenPattern = regexp.MustCompile(`\{(\w+)(?::0(\d+))?\}`)

// Template is a parsed, ready-to-render naming template.
type Template struct {
	raw string
}

// Parse validates raw and returns a Template. Every "{field}" reference is
// checked against the known field set at parse time, so a typo in a
// configured template surfaces immediately rather than mid-import.
func Parse(raw string) (*Template, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("naming: template must not be empty")
	}
	for _, m := range tokenPattern.FindAllStringSubmatch(raw, -1) {
		name := m[1]
		if _, ok := stringFields[name]; ok {
			continue
		}
		if _, ok := intFields[name]; ok {
			continue
		}
		return nil, fmt.Errorf("naming: unknown template field %q", name)
	}
	return &Template{raw: raw}, nil
}

// Render substitutes f into the template and returns the resulting relative
// path. Path separators ("/") in the template itself are preserved as
// directory boundaries; a "/" that appears inside a field's *value* (e.g.
// an artist tagged "AC/DC") is replaced so it can't inject an unintended
// directory level — the only filesystem-reserved character on the
// case-sensitive Linux filesystems this project targets (see
// docs/music-cli-plan.md section 7).
//
// An unknown field (Fields' zero value) renders as an empty string rather
// than being omitted, so a template segment built entirely from unknown
// fields (e.g. "{year} - {album}" with both blank) can render as a run of
// literal separator characters with nothing between them. On Linux
// filesystems this is merely an odd-looking name, not an error. It's worth
// knowing if developing on a non-Linux filesystem, though — a leading or
// trailing space in a path component is silently normalized away by the
// Win32 file APIs, which can make an MkdirAll and a later file-open of the
// literal same Go string target subtly different paths. Fixing tags first
// (`lomax resolve`/`retag`) sidesteps this; teaching Render to skip
// fields entirely rather than blank them is a possible future refinement.
func (t *Template) Render(f Fields) string {
	return tokenPattern.ReplaceAllStringFunc(t.raw, func(tok string) string {
		m := tokenPattern.FindStringSubmatch(tok)
		name, width := m[1], m[2]

		var value string
		if getInt, ok := intFields[name]; ok {
			n := getInt(f)
			if n == 0 {
				value = ""
			} else if width != "" {
				w, _ := strconv.Atoi(width)
				value = fmt.Sprintf("%0*d", w, n)
			} else {
				value = strconv.Itoa(n)
			}
		} else {
			value = stringFields[name](f)
		}
		return sanitizeSegmentValue(value)
	})
}

// sanitizeSegmentValue strips characters from a substituted field value that
// would corrupt the path structure or embed a NUL byte.
func sanitizeSegmentValue(v string) string {
	v = strings.ReplaceAll(v, "/", "-")
	v = strings.ReplaceAll(v, "\x00", "")
	return v
}
