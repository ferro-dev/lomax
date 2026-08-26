package cli

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/acoustid"
	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/musicbrainz"
	"github.com/ferro-dev/lomax/internal/resolve"
)

// acoustIDAPIKeyEnv is the environment variable fallback for --acoustid-key.
// AcoustID lookups are skipped entirely when neither is set — the source
// is optional, not a hard dependency (see docs/music-cli-plan.md section 6).
const acoustIDAPIKeyEnv = "LOMAX_ACOUSTID_API_KEY"

// changedFieldStyle highlights a proposed field value that differs from the
// track's current value.
var changedFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green

// newResolver builds a metadata Resolver backed by real MusicBrainz and (if
// acoustidKey is set) AcoustID clients. Shared by resolve, import, and
// retag, so all three commands resolve metadata identically.
func newResolver(acoustidKey string) *resolve.Resolver {
	mb := musicbrainz.NewClient(fmt.Sprintf("lomax/%s ( https://github.com/ferro-dev/lomax )", Version))
	var ai *acoustid.Client
	if acoustidKey != "" {
		ai = acoustid.NewClient(acoustidKey)
	}
	return resolve.NewResolver(mb, ai)
}

// resolveFile reads file's tags and resolves a metadata proposal for it,
// printing any non-fatal warnings to cmd's stderr. ok is false if the
// file's tags couldn't even be read — callers should skip the file in that
// case, since a suitable warning has already been printed.
func resolveFile(cmd *cobra.Command, ctx context.Context, resolver *resolve.Resolver, file string) (audio.Track, *resolve.Proposal, bool) {
	track, err := audio.ReadTrack(file)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
		return audio.Track{}, nil, false
	}

	proposal, warnings, err := resolver.Resolve(ctx, track)
	for _, w := range warnings {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", file, w)
	}
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %v\n", file, err)
		return track, nil, true
	}
	return track, proposal, true
}

// printProposal renders file's proposed changes (or a no-match line) to
// cmd's stdout.
func printProposal(cmd *cobra.Command, file string, track audio.Track, proposal *resolve.Proposal) error {
	out := cmd.OutOrStdout()
	if proposal == nil {
		_, err := fmt.Fprintf(out, "%s: no metadata match found\n\n", file)
		return err
	}

	t := table.New().Headers("FIELD", "CURRENT", "PROPOSED").Rows(
		diffRow("Title", track.Title, proposal.Title),
		diffRow("Artist", track.Artist, proposal.Artist),
		diffRow("Album", track.Album, proposal.Album),
		diffRow("Album Artist", track.AlbumArtist, proposal.AlbumArtist),
		diffRow("Year", yearString(track.Year), yearString(proposal.Year)),
	)
	_, err := fmt.Fprintf(out, "%s\n  source: %s\n%s\n\n", file, proposal.Source, t.Render())
	return err
}

// diffRow builds one FIELD/CURRENT/PROPOSED table row. An empty proposed
// value means the source didn't offer this field, shown as "(no change)"
// rather than a blank cell so it isn't mistaken for "propose clearing this
// field". A non-empty proposed value that differs from current is
// highlighted — the whole point of a diff view over beets' opaque matching
// (see docs/music-cli-plan.md section 14).
func diffRow(field, current, proposed string) []string {
	display := "(no change)"
	if proposed != "" {
		display = proposed
		if proposed != current {
			display = changedFieldStyle.Render(proposed)
		}
	}
	return []string{field, current, display}
}

func yearString(year int) string {
	if year == 0 {
		return ""
	}
	return fmt.Sprintf("%d", year)
}

// shouldApply reports whether a resolved proposal should actually be
// written: only when a source offered one and the caller isn't previewing.
// Factored out so the gating logic itself — not just WriteTags' mechanics —
// has a unit test that doesn't require a real resolver or network access.
func shouldApply(proposal *resolve.Proposal, dryRun bool) bool {
	return proposal != nil && !dryRun
}

// writableTagsFromProposal converts a resolved Proposal into the fields
// WriteTags understands. A nil proposal writes nothing.
func writableTagsFromProposal(proposal *resolve.Proposal) audio.WritableTags {
	if proposal == nil {
		return audio.WritableTags{}
	}
	return audio.WritableTags{
		Title:       proposal.Title,
		Artist:      proposal.Artist,
		Album:       proposal.Album,
		AlbumArtist: proposal.AlbumArtist,
		Year:        proposal.Year,
	}
}
