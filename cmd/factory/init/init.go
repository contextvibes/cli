// Package initcmd provides the command to initialize the project configuration.
package initcmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed init.md.tpl
var initLongDescription string

// InitCmd represents the init command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var InitCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Initializing ContextVibes configuration...")

		stdout, stderr, err := globals.ExecClient.CaptureOutput(
			ctx,
			".",
			"git",
			"rev-parse",
			"--show-toplevel",
		)
		if err != nil {
			presenter.Error("Failed to determine project root. Are you inside a Git repository?")
			presenter.Detail("Stderr: %s", stderr)

			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("not a git repository")
		}
		projectRoot := strings.TrimSpace(stdout)
		configPath := filepath.Join(projectRoot, config.DefaultConfigFileName)

		//nolint:noinlineerr // Inline check is standard for os.Stat.
		if _, err := os.Stat(configPath); err == nil {
			presenter.Info("Configuration file already exists: %s", presenter.Highlight(configPath))

			return nil
		}

		defaultConfig := config.GetDefaultConfig()
		//nolint:noinlineerr // Inline check is standard for config save.
		if err := config.UpdateAndSaveConfig(defaultConfig, configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		presenter.Success("Successfully created .contextvibes.yaml.")

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(initLongDescription, nil)
	if err != nil {
		panic(err)
	}

	InitCmd.Short = desc.Short
	InitCmd.Long = desc.Long
}
