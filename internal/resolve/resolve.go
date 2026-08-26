// Package resolve implements lomax's metadata resolution pipeline: given a
// track's current tags, propose corrected/completed metadata from external
// sources. Resolution order is MusicBrainz (tag-based search) then AcoustID
// (fingerprint-based, for untagged or unmatched files) — see
// docs/music-cli-plan.md section 6. Discogs and Last.fm are deferred to
// Milestone 5's plugin ports.
package resolve

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ferro-dev/lomax/internal/acoustid"
	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/musicbrainz"
)

// Proposal is a set of proposed metadata field values, attributed to the
// source that produced them. An empty field means "no proposed change to
// this field" — callers should preserve the track's current value.
//
// Track and disc numbers are intentionally not part of Proposal for M2: see
// musicbrainz.Recording's doc comment for why.
type Proposal struct {
	Title       string
	Artist      string
	AlbumArtist string
	Album       string
	Year        int

	// Source is a human-readable attribution shown alongside the proposal,
	// e.g. "MusicBrainz recording <mbid> (score 100)". This is a deliberate
	// differentiator from beets' opaque matching (see
	// docs/music-cli-plan.md section 14): every proposal states exactly
	// where it came from.
	Source string
}

// Resolver resolves metadata proposals for tracks. Its lookup functions are
// injected (rather than holding concrete *musicbrainz.Client /
// *acoustid.Client fields directly) so tests can exercise the fallback
// logic without any network access or fpcalc subprocess; NewResolver wires
// up the real clients for production use.
type Resolver struct {
	// SearchRecording looks up a recording by tags. Nil disables the
	// MusicBrainz step entirely.
	SearchRecording func(ctx context.Context, artist, album, title string) (*musicbrainz.Recording, error)

	// Fingerprint and LookupFingerprint together implement the AcoustID
	// step. Both must be non-nil to enable it.
	Fingerprint       func(path string) (acoustid.Fingerprint, error)
	LookupFingerprint func(ctx context.Context, fp acoustid.Fingerprint) (*acoustid.Match, error)
}

// NewResolver builds a Resolver backed by real MusicBrainz and AcoustID
// clients. Either may be nil to disable that source entirely (e.g. no
// AcoustID API key configured).
func NewResolver(mb *musicbrainz.Client, ai *acoustid.Client) *Resolver {
	r := &Resolver{}
	if mb != nil {
		r.SearchRecording = mb.SearchRecording
	}
	if ai != nil {
		r.Fingerprint = acoustid.Compute
		r.LookupFingerprint = ai.Lookup
	}
	return r
}

// Resolve proposes metadata for track. It returns (nil, warnings, nil) when
// every enabled source was tried and none produced a confident match — that
// is a normal, non-error outcome. warnings collects non-fatal problems
// (a source erroring, or being skipped because it's unconfigured) that
// callers should surface to the user without aborting the resolution.
func (r *Resolver) Resolve(ctx context.Context, track audio.Track) (*Proposal, []string, error) {
	var warnings []string

	if r.SearchRecording != nil && strings.TrimSpace(track.Artist) != "" && strings.TrimSpace(track.Title) != "" {
		rec, err := r.SearchRecording(ctx, track.Artist, track.Album, track.Title)
		switch {
		case err == nil:
			return &Proposal{
				Title:       rec.Title,
				Artist:      rec.Artist,
				AlbumArtist: rec.Artist,
				Album:       rec.Album,
				Year:        rec.Year,
				Source:      fmt.Sprintf("MusicBrainz recording %s (score %d)", rec.MBID, rec.Score),
			}, warnings, nil
		case errors.Is(err, musicbrainz.ErrNoMatch):
			// Fall through to AcoustID.
		default:
			warnings = append(warnings, fmt.Sprintf("musicbrainz lookup failed: %v", err))
		}
	}

	if r.Fingerprint != nil && r.LookupFingerprint != nil {
		fp, err := r.Fingerprint(track.Path)
		switch {
		case err == nil:
			match, lookupErr := r.LookupFingerprint(ctx, fp)
			switch {
			case lookupErr == nil:
				return &Proposal{
					Title:       match.Title,
					Artist:      match.Artist,
					AlbumArtist: match.Artist,
					Album:       match.Album,
					Source:      fmt.Sprintf("AcoustID fingerprint match (recording %s)", match.RecordingMBID),
				}, warnings, nil
			case errors.Is(lookupErr, acoustid.ErrNoMatch):
				// No proposal from any source.
			case errors.Is(lookupErr, acoustid.ErrNoAPIKey):
				warnings = append(warnings, "acoustid: no API key configured, skipping fingerprint lookup")
			default:
				warnings = append(warnings, fmt.Sprintf("acoustid lookup failed: %v", lookupErr))
			}
		case errors.Is(err, acoustid.ErrFpcalcUnavailable):
			warnings = append(warnings, "acoustid: fpcalc not found on PATH, skipping fingerprint lookup")
		default:
			warnings = append(warnings, fmt.Sprintf("acoustid fingerprinting failed: %v", err))
		}
	}

	return nil, warnings, nil
}
