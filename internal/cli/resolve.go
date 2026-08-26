package cli

import (
	"fmt"
	"os"

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

// newResolveCmd builds `lomax resolve <path>`: look up metadata for one
// file or every audio file under a directory from MusicBrainz (by tags)
// and, if configured, AcoustID (by fingerprint), and preview the proposed
// changes as a diff. Tag writing lands in Milestone 3, so this command is
// preview-only — --dry-run exists for forward compatibility with the
// --apply flag M3 will add.
func newResolveCmd() *cobra.Command {
	var acoustidKey string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "resolve <path>",
		Short: "Resolve metadata from MusicBrainz/AcoustID and preview proposed tag changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !dryRun {
				return fmt.Errorf("--dry-run=false is not supported yet; tag writing lands in Milestone 3")
			}
			if acoustidKey == "" {
				acoustidKey = os.Getenv(acoustIDAPIKeyEnv)
			}
			return runResolve(cmd, args[0], acoustidKey)
		},
	}
	cmd.Flags().StringVar(&acoustidKey, "acoustid-key", "", "AcoustID API key for fingerprint lookups (or set "+acoustIDAPIKeyEnv+")")
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "preview proposed changes without writing them")
	return cmd
}

func runResolve(cmd *cobra.Command, path, acoustidKey string) error {
	files, err := scanPathForAudio("resolve", path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "no audio files found under %s\n", path)
		return err
	}

	mb := musicbrainz.NewClient(fmt.Sprintf("lomax/%s ( https://github.com/ferro-dev/lomax )", Version))
	var ai *acoustid.Client
	if acoustidKey != "" {
		ai = acoustid.NewClient(acoustidKey)
	}
	resolver := resolve.NewResolver(mb, ai)

	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	for _, file := range files {
		track, err := audio.ReadTrack(file)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
			continue
		}

		proposal, warnings, err := resolver.Resolve(ctx, track)
		for _, w := range warnings {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", file, w)
		}
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %v\n", file, err)
			continue
		}
		if proposal == nil {
			if _, err := fmt.Fprintf(out, "%s: no metadata match found\n\n", file); err != nil {
				return err
			}
			continue
		}

		t := table.New().Headers("FIELD", "CURRENT", "PROPOSED").Rows(
			diffRow("Title", track.Title, proposal.Title),
			diffRow("Artist", track.Artist, proposal.Artist),
			diffRow("Album", track.Album, proposal.Album),
			diffRow("Album Artist", track.AlbumArtist, proposal.AlbumArtist),
			diffRow("Year", yearString(track.Year), yearString(proposal.Year)),
		)
		if _, err := fmt.Fprintf(out, "%s\n  source: %s\n%s\n\n", file, proposal.Source, t.Render()); err != nil {
			return err
		}
	}
	return nil
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
