package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/audio"
)

// newRetagCmd builds `lomax retag <path>`: the manage-in-place workflow —
// resolve metadata for files already in the library, show the diff, and
// (unless --dry-run) write the resolved tags back in place. Unlike import,
// retag never moves or renames files.
func newRetagCmd() *cobra.Command {
	var acoustidKey string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "retag <path>",
		Short: "Resolve metadata and fix tags on files already in your library, in place",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if acoustidKey == "" {
				acoustidKey = os.Getenv(acoustIDAPIKeyEnv)
			}
			return runRetag(cmd, args[0], acoustidKey, dryRun)
		},
	}
	cmd.Flags().StringVar(&acoustidKey, "acoustid-key", "", "AcoustID API key for fingerprint lookups (or set "+acoustIDAPIKeyEnv+")")
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "preview proposed changes without writing them")
	return cmd
}

func runRetag(cmd *cobra.Command, path, acoustidKey string, dryRun bool) error {
	files, err := scanPathForAudio("retag", path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "no audio files found under %s\n", path)
		return err
	}

	resolver := newResolver(acoustidKey)
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	for _, file := range files {
		track, proposal, ok := resolveFile(cmd, ctx, resolver, file)
		if !ok {
			continue
		}
		if err := printProposal(cmd, file, track, proposal); err != nil {
			return err
		}
		if !shouldApply(proposal, dryRun) {
			continue
		}

		if err := audio.WriteTags(file, writableTagsFromProposal(proposal)); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
			continue
		}
		if _, err := fmt.Fprintf(out, "  wrote tags to %s\n\n", file); err != nil {
			return err
		}
	}
	return nil
}
