// cmd/project_list.go
package cmd

import (
	"errors"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	projectOwner string
)

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"list-projects", "boards"},
	Short:   "Lists the GitHub Projects for the repository's owner.",
	Long: `Fetches and displays a list of GitHub Projects (boards) associated with the repository's owner.

This command uses the 'gh' CLI in the background ('gh project list') to provide a
quick overview of available project boards without leaving the terminal.`,
	Example: `  # List projects for the current repository's owner
  contextvibes project list

  # List projects for a different owner (e.g., an organization)
  contextvibes project list --owner "my-org"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if !ExecClient.CommandExists("gh") {
			presenter.Error(
				"GitHub CLI ('gh') not found. This command is required to list projects.",
			)
			presenter.Advice(
				"Please install it from https://cli.github.com/ and authenticate with 'gh auth login'.",
			)
			return errors.New("gh cli not found")
		}

		presenter.Summary("Fetching GitHub Projects...")

		// Build the 'gh project list' command with flags
		ghArgs := []string{"project", "list"}
		if projectOwner != "" {
			ghArgs = append(ghArgs, "--owner", projectOwner)
		}

		// Execute the command
		if err := ExecClient.Execute(ctx, ".", "gh", ghArgs...); err != nil {
			presenter.Error("Failed to fetch GitHub projects: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)

	projectListCmd.Flags().StringVarP(&projectOwner, "owner", "o", "", "List projects for the specified owner (user or organization)")
}
