package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/audio"
	"github.com/ferro-dev/lomax/internal/library"
)

// newRetagCmd builds `lomax retag <path>`: the manage-in-place workflow —
// resolve metadata for files already in the library, show the diff, and
// (unless --dry-run) write the resolved tags back in place. Unlike import,
// retag never moves or renames files.
func newRetagCmd() *cobra.Command {
	var acoustidKey string
	var dbPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "retag <path>",
		Short: "Resolve metadata and fix tags on files already in your library, in place",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if acoustidKey == "" {
				acoustidKey = os.Getenv(acoustIDAPIKeyEnv)
			}
			return runRetag(cmd, args[0], acoustidKey, dbPath, dryRun)
		},
	}
	cmd.Flags().StringVar(&acoustidKey, "acoustid-key", "", "AcoustID API key for fingerprint lookups (or set "+acoustIDAPIKeyEnv+")")
	addLibraryDBFlag(cmd, &dbPath)
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "preview proposed changes without writing them")
	return cmd
}

func runRetag(cmd *cobra.Command, path, acoustidKey, dbPath string, dryRun bool) error {
	files, err := scanPathForAudio("retag", path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "no audio files found under %s\n", path)
		return err
	}

	// dry-run touches nothing, including the library database — only open
	// it once there's actually something to record.
	var db *library.DB
	if !dryRun {
		db, err = openLibrary(dbPath)
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
	}

	resolver, closeResolver, err := newResolver(cmd, acoustidKey)
	if err != nil {
		return err
	}
	defer closeResolver()

	ctx := cmd.Context()
	out, errOut := cmd.OutOrStdout(), cmd.ErrOrStderr()
	for _, file := range files {
		track, proposal, ok := resolveFile(cmd, ctx, resolver, file)
		if !ok {
			continue
		}
		if err := printProposal(cmd, file, track, proposal); err != nil {
			return err
		}
		if dryRun {
			continue
		}

		final := track
		if shouldApply(proposal, dryRun) {
			if err := audio.WriteTags(file, writableTagsFromProposal(proposal)); err != nil {
				_, _ = fmt.Fprintf(errOut, "warning: %v\n", err)
				continue
			}
			// Re-read the file's own tags after writing, rather than
			// hand-merging track and proposal, so the database always
			// reflects exactly what's on disk.
			updated, err := audio.ReadTrack(file)
			if err != nil {
				_, _ = fmt.Fprintf(errOut, "warning: %v\n", err)
				continue
			}
			final = updated
			if _, err := fmt.Fprintf(out, "  wrote tags to %s\n\n", file); err != nil {
				return err
			}
		}

		if err := db.Upsert(final); err != nil {
			_, _ = fmt.Fprintf(errOut, "warning: %s: failed to update library database: %v\n", file, err)
		}
	}
	return nil
}
