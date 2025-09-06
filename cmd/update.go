// cmd/update.go
package cmd

import (
	"os"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Installs or updates the CLI to the latest version.",
	Long: `Installs or updates the CLI to the latest released version from the official GitHub repository.

This command runs 'go install github.com/contextvibes/cli/cmd/cv@latest' in the background.
The 'go install' command will compile the binary and place it in your Go binary path
(usually '$HOME/go/bin').

For the command to be available in your shell, you must have your Go binary path
included in your system's PATH environment variable.`,
	Example: `  # Install or update to the latest version:
  contextvibes update`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		ctx := cmd.Context()
		logger := AppLogger

		presenter.Summary("Updating the CLI to the latest version...")

		installPath := "github.com/contextvibes/cli/cmd/cv@latest"
		presenter.Step("Running 'go install %s'...", installPath)

		err := ExecClient.Execute(ctx, ".", "go", "install", installPath)
		if err != nil {
			presenter.Error("Update failed. See the 'go install' output above for details.")
			logger.ErrorContext(ctx, "go install failed during update", "error", err)
			return err
		}

		presenter.Newline()
		presenter.Success("CLI has been successfully installed/updated.")
		presenter.Advice("The new binary 'cv' is now available in your Go binary path.")
		presenter.Detail("  You may need to restart your shell for the changes to take effect.")
		presenter.Advice("Ensure your Go binary path (e.g., '$HOME/go/bin') is in your system's PATH.")
		presenter.Detail("  You can check with: 'which cv'")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
