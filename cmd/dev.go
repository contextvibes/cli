// cmd/dev.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

// devCmd represents the base command for developer-focused utilities.
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Developer-focused utilities for working on the CLI itself.",
	Long:  `Provides commands useful for the development and maintenance of the contextvibes CLI.`,
}

// devAliasCmd suggests and prints a shell alias for running the dev version.
var devAliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Prints a suggested shell alias for running the dev version.",
	Long: `Prints a shell 'alias' command to standard output.

This alias ('cvd' for Context Vibes Dev) simplifies running the CLI from source
during development. The output of this command is designed to be used with 'eval'.`,
	Example: `  # To use the alias in your current session:
  eval "$(contextvibes dev alias)"

  # To make the alias permanent, add the following to your shell profile (~/.bashrc, ~/.zshrc):
  eval "$(contextvibes dev alias)"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// We write instructions to stderr so that stdout contains only the alias command for `eval`.
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		aliasName := "cvd"
		aliasCommand := "go run cmd/cv/main.go" // CORRECTED PATH
		aliasString := fmt.Sprintf("alias %s='%s'", aliasName, aliasCommand)

		// Print the alias command to STDOUT. This is the part that `eval` will execute.
		fmt.Fprintln(presenter.Out(), aliasString)

		// Print user instructions to STDERR so they don't interfere with `eval`.
		shell := filepath.Base(os.Getenv("SHELL"))
		profileFile := ""
		switch shell {
		case "bash":
			profileFile = "~/.bashrc"
		case "zsh":
			profileFile = "~/.zshrc"
		default:
			profileFile = "your shell profile file (e.g., ~/.bash_profile, ~/.zprofile)"
		}

		presenter.Newline()
		presenter.Advice("Alias '%s' is ready for your current shell session.", presenter.Highlight(aliasName))
		presenter.Advice("To make this alias permanent, add the following line to %s:", presenter.Highlight(profileFile))
		presenter.Detail("  eval \"$(%s dev alias)\"", cmd.Root().Use) // Note: This will now use 'cv' once installed

		return nil
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
	devCmd.AddCommand(devAliasCmd)
}
