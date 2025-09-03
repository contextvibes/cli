// cmd/project_setup_labels.go
package cmd

import (
	"errors"
	"strings"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

// label defines the properties for a GitHub label we want to ensure exists.
type label struct {
	Name        string
	Color       string
	Description string
}

// requiredLabels is the canonical list of labels our workflow requires.
var requiredLabels = []label{
	{Name: "epic", Color: "3E4B8B", Description: "A large body of work that can be broken down into smaller stories."},
	{Name: "user-story", Color: "5319E7", Description: "A specific feature or requirement from a user's perspective."},
	{Name: "feature", Color: "a2eeef", Description: "New feature or request."},
	{Name: "bug", Color: "d73a4a", Description: "Something isn't working."},
	{Name: "documentation", Color: "0075ca", Description: "Improvements or additions to documentation."},
	{Name: "chore", Color: "cfd3d7", Description: "Miscellaneous tasks that don't add user-facing value."},
	{Name: "enhancement", Color: "a2eeef", Description: "New feature or request."},
}

var projectSetupLabelsCmd = &cobra.Command{
	Use:     "setup-labels",
	Aliases: []string{"labels"},
	Short:   "Creates a standard set of GitHub labels in the repository.",
	Long: `Ensures that the current repository has a standard set of labels required for the project workflow (e.g., 'epic', 'user-story').

This command is idempotent. If a label already exists, it will be skipped. It uses the 'gh' CLI in the background to create the labels.`,
	Example: `  contextvibes project setup-labels`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if !ExecClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found.")
			return errors.New("gh cli not found")
		}

		presenter.Summary("Setting up required GitHub labels...")
		presenter.Newline()

		createdCount := 0
		existsCount := 0
		errorCount := 0

		for _, lbl := range requiredLabels {
			presenter.Step("Ensuring label '%s' exists...", lbl.Name)
			ghArgs := []string{
				"label", "create", lbl.Name,
				"--color", lbl.Color,
				"--description", lbl.Description,
			}

			_, stderr, err := ExecClient.CaptureOutput(ctx, ".", "gh", ghArgs...)
			if err != nil {
				// Check if the error is because the label already exists, which is not a failure for us.
				if strings.Contains(stderr, "already exists") {
					presenter.Success("  ✓ '%s' already exists.", lbl.Name)
					existsCount++
				} else {
					presenter.Error("  ! Failed to create '%s': %v", lbl.Name, err)
					presenter.Detail("    Stderr: %s", stderr)
					errorCount++
				}
			} else {
				presenter.Success("  ✓ Created label '%s'.", lbl.Name)
				createdCount++
			}
		}

		presenter.Newline()
		presenter.Header("--- Label Setup Summary ---")
		presenter.Detail("Labels created: %d", createdCount)
		presenter.Detail("Labels already existed: %d", existsCount)
		presenter.Detail("Errors encountered: %d", errorCount)

		if errorCount > 0 {
			return errors.New("one or more labels could not be created")
		}
		presenter.Success("All required labels are configured for this repository.")
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectSetupLabelsCmd)
}
