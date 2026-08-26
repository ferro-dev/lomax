package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/ferro-dev/lomax/internal/pluginhost"
)

// newPluginCmd builds the `lomax plugin` command group: list, install, and
// remove subprocess plugins (see docs/music-cli-plan.md section 9).
func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage lomax plugins",
	}
	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginRemoveCmd())
	return cmd
}

func newPluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List discovered plugins and whether each one handshakes successfully",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginList(cmd)
		},
	}
}

func runPluginList(cmd *cobra.Command) error {
	pluginDir, err := pluginhost.DefaultPluginDir()
	if err != nil {
		return err
	}

	descriptors := pluginhost.Discover(pluginDir, pluginhost.ExtraDirsFromEnv())
	out := cmd.OutOrStdout()
	if len(descriptors) == 0 {
		_, err := fmt.Fprintf(out, "no plugins found in %s or on $PATH\n", pluginDir)
		return err
	}

	t := table.New().Headers("NAME", "PATH", "STATUS")
	for _, d := range descriptors {
		status := "ok"
		if h, err := pluginhost.Load(d); err != nil {
			status = "error: " + err.Error()
		} else {
			h.Close()
		}
		t.Row(d.Name, d.Path, status)
	}
	_, err = fmt.Fprintln(out, t.Render())
	return err
}

func newPluginInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <name>",
		Short: "Build a first-party plugin from source and install it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginInstall(cmd, args[0])
		},
	}
}

func runPluginInstall(cmd *cobra.Command, name string) error {
	// There's no plugin registry or release-binary distribution yet —
	// that lands with Milestone 6's release pipeline (see
	// docs/music-cli-plan.md section 12). Until then, "install" means
	// building a first-party plugin from this monorepo checkout, which is
	// only meaningful run from inside one.
	sourceDir := filepath.Join("plugins", name)
	if info, err := os.Stat(sourceDir); err != nil || !info.IsDir() {
		return fmt.Errorf("plugin install: no source found at %s — run this from a lomax source checkout (fetching prebuilt release binaries isn't implemented yet)", sourceDir)
	}

	pluginDir, err := pluginhost.DefaultPluginDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return fmt.Errorf("plugin install: create %s: %w", pluginDir, err)
	}

	destPath := filepath.Join(pluginDir, "lomax-plugin-"+name)
	build := exec.CommandContext(cmd.Context(), "go", "build", "-o", destPath, ".")
	build.Dir = sourceDir
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("plugin install: build %s: %w\n%s", name, err, output)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "installed %s to %s\n", name, destPath)
	return err
}

func newPluginRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginRemove(cmd, args[0])
		},
	}
}

func runPluginRemove(cmd *cobra.Command, name string) error {
	pluginDir, err := pluginhost.DefaultPluginDir()
	if err != nil {
		return err
	}

	path := filepath.Join(pluginDir, "lomax-plugin-"+name)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("plugin remove: %w", err)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", path)
	return err
}
