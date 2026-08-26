// Package cli defines the lomax command tree.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the lomax version. Overridable at build time via
// -ldflags "-X github.com/ferro-dev/lomax/internal/cli.Version=x.y.z".
var Version = "0.0.0-dev"

// Attribution is the mandated Alan Lomax credit shown by `--version` and the
// `about` command. It must appear on every primary surface — see
// docs/attribution.md and docs/music-cli-plan.md section 13.
const Attribution = "Named after Alan Lomax (1915–2002). " +
	"Independent project, not affiliated with the Lomax estate or ACE."

// newRootCmd builds the root command tree. Kept as a constructor (rather than a
// package-level singleton) so tests can run it in isolation with captured I/O.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "lomax",
		Short:         "A Linux-native CLI music library manager",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("lomax {{.Version}}\n" + Attribution + "\n")
	root.AddCommand(newAboutCmd())
	root.AddCommand(newInspectCmd())
	root.AddCommand(newResolveCmd())
	root.AddCommand(newRetagCmd())
	root.AddCommand(newImportCmd())
	return root
}

// newAboutCmd prints the version and the mandatory attribution line.
func newAboutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "about",
		Short: "Show version and attribution",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "lomax %s\n%s\n", Version, Attribution)
			return err
		},
	}
}

// Execute runs the root command and returns any error for main to handle.
func Execute() error {
	return newRootCmd().Execute()
}
