// cmd/improve.go
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	issueType  string
	issueTitle string
	issueBody  string
)

var improveCmd = &cobra.Command{
	Use:     "improve",
	Aliases: []string{"issue", "suggest"},
	Short:   "Create a new feature, bug, or chore suggestion as a GitHub Issue.",
	Long:    `Creates a new GitHub Issue, either interactively or non-interactively via flags.\nThis command uses the 'gh' CLI in the background to create the issue with the provided details and appropriate labels.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if !ExecClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found. This command is required to create issues.")
			presenter.Advice("Please install it from https://cli.github.com/ and authenticate with 'gh auth login'.")
			return errors.New("gh cli not found")
		}

		// If title is provided via flag, run non-interactively.
		if issueTitle != "" {
			presenter.Info("Running in non-interactive mode.")
			if issueType == "" {
				issueType = "feature" // Default to feature
				presenter.Info("No --type specified, defaulting to 'feature'.")
			}
		} else {
			// Interactive flow using huh
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("What kind of improvement is this?").
						Options(
							huh.NewOption("Feature Request", "feature"),
							huh.NewOption("Bug Report", "bug"),
							huh.NewOption("Chore / Refactor", "chore"),
							huh.NewOption("Documentation", "documentation"),
							huh.NewOption("Epic (Large Feature)", "epic"),
						).
						Value(&issueType),

					huh.NewInput().
						Title("What is the title for this issue?").
						Value(&issueTitle).
						Validate(func(s string) error {
							if len(s) == 0 {
								return fmt.Errorf("title cannot be empty")
							}
							return nil
						}),

					huh.NewText().
						Title("Please describe the improvement.").
						Value(&issueBody),
				),
			)

			if err := form.Run(); err != nil {
				presenter.Error("Improvement submission aborted: %v", err)
				return err
			}
		}

		// Get labels based on type
		labels := []string{issueType}
		if issueType == "feature" || issueType == "epic" {
			labels = append(labels, "enhancement")
		}

		// Confirmation
		var confirmed bool = false
		if assumeYes {
			confirmed = true
		} else {
			presenter.Newline()
			presenter.Header("--- Issue Preview ---")
			presenter.Detail("Title: %s", issueTitle)
			presenter.Detail("Labels: %s", strings.Join(labels, ", "))
			presenter.Step("Body:")
			fmt.Fprintln(presenter.Out(), issueBody)
			presenter.Newline()

			var err error
			confirmed, err = presenter.PromptForConfirmation("Create this issue on GitHub?")
			if err != nil {
				return err
			}
		}

		if !confirmed {
			presenter.Info("Issue creation aborted by user.")
			return nil
		}

		// Execution
		presenter.Step("Creating issue via 'gh' CLI...")
		ghArgs := []string{
			"issue", "create",
			"--title", issueTitle,
			"--body", issueBody,
		}
		for _, label := range labels {
			ghArgs = append(ghArgs, "--label", label)
		}

		if err := ExecClient.Execute(ctx, ".", "gh", ghArgs...); err != nil {
			presenter.Error("Failed to create GitHub issue: %v", err)
			return err
		}

		presenter.Success("Successfully created issue.")
		return nil
	},
}

func init() {
	projectCmd.AddCommand(improveCmd)
	improveCmd.Flags().StringVarP(&issueType, "type", "t", "", "Type of the issue (feature, bug, chore, documentation, epic)")
	improveCmd.Flags().StringVarP(&issueTitle, "title", "T", "", "Title of the issue")
	improveCmd.Flags().StringVarP(&issueBody, "body", "b", "", "Body of the issue")
}
