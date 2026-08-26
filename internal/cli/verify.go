package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVerifyCmd builds `lomax verify <path>`: reconcile the library database
// against the actual files under path — repointing rows for files that
// were moved, and reporting files that vanished (or, with --prune,
// forgetting them) and files on disk the database doesn't know about yet.
func newVerifyCmd() *cobra.Command {
	var dbPath string
	var prune bool

	cmd := &cobra.Command{
		Use:   "verify <path>",
		Short: "Check the library database against the filesystem: detect moved and deleted files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, args[0], dbPath, prune)
		},
	}
	addLibraryDBFlag(cmd, &dbPath)
	cmd.Flags().BoolVar(&prune, "prune", false, "remove database rows for files that no longer exist and couldn't be matched to a move")
	return cmd
}

func runVerify(cmd *cobra.Command, path, dbPath string, prune bool) error {
	db, err := openLibrary(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	report, err := db.Reconcile(path)
	if err != nil {
		return err
	}

	out, errOut := cmd.OutOrStdout(), cmd.ErrOrStderr()
	for _, m := range report.Moved {
		if err := db.UpdatePath(m.From, m.To); err != nil {
			_, _ = fmt.Fprintf(errOut, "warning: %v\n", err)
			continue
		}
		if _, err := fmt.Fprintf(out, "moved: %s -> %s\n", m.From, m.To); err != nil {
			return err
		}
	}
	for _, r := range report.Missing {
		if !prune {
			if _, err := fmt.Fprintf(out, "missing: %s\n", r.Path); err != nil {
				return err
			}
			continue
		}
		if err := db.Delete(r.Path); err != nil {
			_, _ = fmt.Fprintf(errOut, "warning: %v\n", err)
			continue
		}
		if _, err := fmt.Fprintf(out, "pruned: %s (no longer exists)\n", r.Path); err != nil {
			return err
		}
	}
	for _, u := range report.Untracked {
		if _, err := fmt.Fprintf(out, "untracked: %s\n", u); err != nil {
			return err
		}
	}

	if len(report.Moved) == 0 && len(report.Missing) == 0 && len(report.Untracked) == 0 {
		_, err := fmt.Fprintln(out, "library is consistent with the filesystem")
		return err
	}
	return nil
}
