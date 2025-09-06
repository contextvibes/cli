// cmd/project_link.go
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var parentLinkNumber int

var projectLinkCmd = &cobra.Command{
	Use:   "link <issue-number>",
	Short: "Links an issue to a parent epic or user story.",
	Long: `Links an existing issue to a parent issue by appending a task list item to the parent's body.

This command first reads the parent issue's current body, appends the new task,
and then updates the issue with the new combined body.`,
	Example: `  # Link issue #47 to its parent epic, issue #24
  contextvibes project link 47 --to 24`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		childIssueNumberStr := args[0]
		_, err := strconv.Atoi(childIssueNumberStr)
		if err != nil {
			presenter.Error("Invalid issue number provided: %s", childIssueNumberStr)
			return fmt.Errorf("invalid issue number: %w", err)
		}

		if parentLinkNumber <= 0 {
			presenter.Error("A parent issue number must be provided using the --to flag.")
			return errors.New("--to flag is required")
		}

		if !ExecClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found.")
			return errors.New("gh cli not found")
		}

		presenter.Summary(
			"Linking issue #%s to parent #%d...",
			childIssueNumberStr,
			parentLinkNumber,
		)

		// Step 1: Fetch the parent issue's current body
		presenter.Step("  Fetching current body of parent issue #%d...", parentLinkNumber)
		parentStr := fmt.Sprintf("%d", parentLinkNumber)
		fetchArgs := []string{"issue", "view", parentStr, "--json", "body"}
		parentJSON, stderr, fetchErr := ExecClient.CaptureOutput(ctx, ".", "gh", fetchArgs...)
		if fetchErr != nil {
			presenter.Error("Failed to fetch parent issue body: %v", fetchErr)
			presenter.Detail("Stderr: %s", stderr)
			return fetchErr
		}

		var parentData struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal([]byte(parentJSON), &parentData); err != nil {
			presenter.Error("Failed to parse parent issue data: %v", err)
			return err
		}

		// Step 2: Append the new task to the existing body
		newBody := parentData.Body
		taskToAdd := fmt.Sprintf("- [ ] #%s", childIssueNumberStr)
		// Ensure there's a newline before our addition if the body isn't empty
		if newBody != "" {
			newBody += "\n"
		}
		newBody += taskToAdd

		// Step 3: Update the issue with the new, combined body
		presenter.Step("  Appending task to parent issue's body...")
		updateArgs := []string{
			"issue", "edit", parentStr,
			"--body", newBody,
		}

		if err := ExecClient.Execute(ctx, ".", "gh", updateArgs...); err != nil {
			presenter.Error("Failed to link issue to parent by editing its body: %v", err)
			return err
		}

		presenter.Success(
			"Successfully linked issue #%s to parent #%d.",
			childIssueNumberStr,
			parentLinkNumber,
		)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectLinkCmd)

	projectLinkCmd.Flags().
		IntVar(&parentLinkNumber, "to", 0, "The issue number of the parent epic to link to (required)")
	projectLinkCmd.MarkFlagRequired("to")
}
