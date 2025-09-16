// cmd/factory/init/init.go
package init_cmd

import (
	_ "embed"
	"errors"
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

// InitCmd represents the init command
var InitCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return errors.New("not a git repository")
		}
		projectRoot := strings.TrimSpace(stdout)
		configPath := filepath.Join(projectRoot, config.DefaultConfigFileName)

		if _, err := os.Stat(configPath); err == nil {
			presenter.Info("Configuration file already exists: %s", presenter.Highlight(configPath))
			return nil
		}

		defaultConfig := config.GetDefaultConfig()
		if err := config.UpdateAndSaveConfig(defaultConfig, configPath); err != nil {
			return err
		}

		presenter.Success("Successfully created .contextvibes.yaml.")
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(initLongDescription, nil)
	if err != nil {
		panic(err)
	}
	InitCmd.Short = desc.Short
	InitCmd.Long = desc.Long
}
