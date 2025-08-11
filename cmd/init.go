// cmd/init.go
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a project with a default .contextvibes.yaml configuration.",
	Long: `Checks for an existing .contextvibes.yaml file in the project root.
If one does not exist, it creates a new file populated with the default
configuration values, including Git settings and validation patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context()

		presenter.Summary("Initializing ContextVibes configuration...")

		// Find the git repository root
		stdout, stderr, err := ExecClient.CaptureOutput(ctx, ".", "git", "rev-parse", "--show-toplevel")
		if err != nil {
			presenter.Error("Failed to determine project root. Are you inside a Git repository?")
			presenter.Detail("Error: %v", err)
			presenter.Detail("Stderr: %s", stderr)
			return errors.New("not a git repository")
		}
		projectRoot := strings.TrimSpace(stdout)
		configPath := filepath.Join(projectRoot, config.DefaultConfigFileName)

		// Check if the file already exists
		if _, err := os.Stat(configPath); err == nil {
			presenter.Info("Configuration file already exists: %s", presenter.Highlight(configPath))
			presenter.Advice("No action taken.")
			return nil
		} else if !os.IsNotExist(err) {
			presenter.Error("Could not check for existing config file: %v", err)
			return err
		}

		// Create the default configuration
		presenter.Step("Creating default configuration file at %s...", configPath)
		defaultConfig := config.GetDefaultConfig()
		if err := config.UpdateAndSaveConfig(defaultConfig, configPath); err != nil {
			presenter.Error("Failed to write configuration file: %v", err)
			return err
		}

		presenter.Success("Successfully created .contextvibes.yaml.")
		presenter.Advice("Review the file and customize it for your project's needs.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
