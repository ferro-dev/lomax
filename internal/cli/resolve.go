package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newResolveCmd builds `lomax resolve <path>`: look up metadata for one
// file or every audio file under a directory from MusicBrainz (by tags)
// and, if configured, AcoustID (by fingerprint), and preview the proposed
// changes as a diff. This command is preview-only, even with tag writing
// now implemented (see import/retag) — resolve is the no-side-effects way
// to check what would happen. --dry-run=false is rejected outright.
func newResolveCmd() *cobra.Command {
	var acoustidKey string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "resolve <path>",
		Short: "Resolve metadata from MusicBrainz/AcoustID and preview proposed tag changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !dryRun {
				return fmt.Errorf("resolve is always dry-run; use `lomax retag` or `lomax import` to write changes")
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

	resolver, closeResolver, err := newResolver(cmd, acoustidKey)
	if err != nil {
		return err
	}
	defer closeResolver()

	ctx := cmd.Context()
	for _, file := range files {
		track, proposal, ok := resolveFile(cmd, ctx, resolver, file)
		if !ok {
			continue
		}
		if err := printProposal(cmd, file, track, proposal); err != nil {
			return err
		}
	}
	return nil
}
