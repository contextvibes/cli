// cmd/improve.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	issueType         string
	issueTitle        string
	issueBody         string
	parentIssueNumber int
)

var improveCmd = &cobra.Command{
	Use:     "improve",
	Aliases: []string{"issue", "suggest"},
	Short:   "Create a new feature, bug, or chore suggestion as a GitHub Issue.",
	Long: `Creates a new GitHub Issue, either interactively or non-interactively via flags.
When using the --parent flag, it automatically adds the new issue to the parent's official GitHub Tasklist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if !ExecClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found.")
			return errors.New("gh cli not found")
		}

		if issueTitle == "" { // Interactive Mode
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("What kind of improvement is this?").
						Options(
							huh.NewOption("Epic", "epic"),
							huh.NewOption("User Story", "user-story"),
							huh.NewOption("Task / Chore", "chore"),
							huh.NewOption("Bug Report", "bug"),
							huh.NewOption("Documentation", "documentation"),
						).
						Value(&issueType),
					huh.NewInput().Title("What is the title for this issue?").Value(&issueTitle),
					huh.NewText().Title("Please describe the improvement.").Value(&issueBody),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}

		labels := []string{issueType}
		if issueType == "feature" || issueType == "epic" || issueType == "user-story" {
			labels = append(labels, "enhancement")
		}

		confirmed := false
		if assumeYes {
			confirmed = true
		} else {
			presenter.Newline()
			presenter.Header("--- Issue Preview ---")
			presenter.Detail("Title: %s", issueTitle)
			presenter.Detail("Labels: %s", strings.Join(labels, ", "))
			if parentIssueNumber > 0 {
				presenter.Detail("Parent Issue: #%d", parentIssueNumber)
			}
			presenter.Step("Body:")
			_, _ = fmt.Fprintln(presenter.Out(), issueBody)
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

		presenter.Step("Creating issue via 'gh' CLI...")
		ghArgs := []string{"issue", "create", "--title", issueTitle, "--body", issueBody}
		for _, label := range labels {
			ghArgs = append(ghArgs, "--label", label)
		}

		stdout, stderr, err := ExecClient.CaptureOutput(ctx, ".", "gh", ghArgs...)
		if err != nil {
			presenter.Error("Failed to create GitHub issue: %v", err)
			presenter.Detail("Stderr: %s", stderr)
			return err
		}
		newIssueURL := strings.TrimSpace(stdout)
		presenter.Success("Successfully created issue: %s", newIssueURL)

		if parentIssueNumber > 0 {
			if err := linkToParentAsTask(presenter, ctx, newIssueURL, parentIssueNumber); err != nil {
				// The helper function prints its own errors.
				return err
			}
		}

		return nil
	},
}

// linkToParentAsTask uses the modern `gh issue edit --add-task` command.
func linkToParentAsTask(presenter *ui.Presenter, ctx context.Context, newIssueURL string, parentNum int) error {
	presenter.Step("Adding new issue %s as a sub-task to parent #%d...", newIssueURL, parentNum)
	
	parentNumStr := fmt.Sprintf("%d", parentNum)

	// This is the new, simpler, and more correct command.
	updateArgs := []string{
		"issue", "edit", parentNumStr,
		"--add-task", newIssueURL,
	}

	// We use CaptureOutput to check for a specific warning from `gh`.
	_, stderr, err := ExecClient.CaptureOutput(ctx, ".", "gh", updateArgs...)
	if err != nil {
		presenter.Error("Failed to link issue to parent as a task: %v", err)
		presenter.Detail("Stderr: %s", stderr)
		return err
	}

	// The gh command might print a warning if the body didn't previously have a tasklist. This is okay.
	if strings.Contains(stderr, "warning:") {
		presenter.Warning("  gh CLI reported a warning (this is usually okay): %s", stderr)
	}

	presenter.Success("Successfully added as a sub-task to parent issue #%d.", parentNum)
	return nil
}

func init() {
	projectCmd.AddCommand(improveCmd)
	improveCmd.Flags().StringVarP(&issueType, "type", "t", "", "Type of the issue (epic, user-story, chore, bug, documentation)")
	improveCmd.Flags().StringVarP(&issueTitle, "title", "T", "", "Title of the issue")
	improveCmd.Flags().StringVarP(&issueBody, "body", "b", "", "Body of the issue")
	improveCmd.Flags().IntVarP(&parentIssueNumber, "parent", "p", 0, "The issue number of the parent epic or user story")
}
