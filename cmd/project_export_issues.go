// cmd/project_export_issues.go
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var exportOutputFile string

// temp struct for unmarshalling JSON output from 'gh issue view'
type issueExportData struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Comments []struct {
		Body string `json:"body"`
	} `json:"comments"`
}

var projectExportIssuesCmd = &cobra.Command{
	Use:     "export-issues",
	Aliases: []string{"export"},
	Short:   "Exports all issues with content and comments to a single file.",
	Long: `Fetches all issues from the current repository, including the full body
content and all comments for each issue. It then compiles all of this
information into a single, comprehensive Markdown file.

This is extremely useful for providing a complete project context dump to an
AI assistant for analysis, planning, or summarization.`,
	Example: `  # Export all issues to the default file 'project_issues_export.md'
  contextvibes project export-issues

  # Export all issues to a custom file
  contextvibes project export-issues -o full_backlog.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if !ExecClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found.")
			return errors.New("gh cli not found")
		}

		presenter.Summary("Exporting full issue context to %s...", exportOutputFile)

		// 1. Get all issue numbers
		presenter.Step("Fetching all issue numbers...")
		listArgs := []string{"issue", "list", "--json", "number", "--jq", ".[].number"}
		stdout, stderr, err := ExecClient.CaptureOutput(ctx, ".", "gh", listArgs...)
		if err != nil {
			presenter.Error("Failed to list issue numbers: %v", err)
			presenter.Detail("Stderr: %s", stderr)
			return err
		}

		issueNumbers := strings.Fields(stdout)
		if len(issueNumbers) == 0 {
			presenter.Info("No issues found in the repository.")
			return nil
		}
		presenter.Success("✓ Found %d issues.", len(issueNumbers))

		// 2. Loop and export each issue
		var buffer bytes.Buffer
		totalIssues := len(issueNumbers)
		for i, numStr := range issueNumbers {
			presenter.Step("Exporting issue %d of %d: #%s...", i+1, totalIssues, numStr)

			// MODIFIED: Use --json flag for reliable machine-readable output
			viewArgs := []string{"issue", "view", numStr, "--json", "title,body,comments"}
			issueJSON, viewStderr, viewErr := ExecClient.CaptureOutput(ctx, ".", "gh", viewArgs...)
			if viewErr != nil {
				presenter.Warning(
					"Could not export issue #%s: %v. Stderr: %s",
					numStr,
					viewErr,
					viewStderr,
				)
				continue
			}

			var data issueExportData
			if err := json.Unmarshal([]byte(issueJSON), &data); err != nil {
				presenter.Warning("Could not parse JSON for issue #%s: %v", numStr, err)
				continue
			}

			// Format the output
			buffer.WriteString(fmt.Sprintf("\n\n---\n\n## Issue #%s: %s\n\n", numStr, data.Title))
			buffer.WriteString("### Body\n\n")
			buffer.WriteString(data.Body)

			if len(data.Comments) > 0 {
				buffer.WriteString("\n\n### Comments\n")
				for _, comment := range data.Comments {
					buffer.WriteString("\n---\n\n")
					buffer.WriteString(comment.Body)
				}
			}
		}

		// 3. Write to file
		presenter.Step("Writing aggregated content to %s...", exportOutputFile)
		err = os.WriteFile(exportOutputFile, buffer.Bytes(), 0o644)
		if err != nil {
			presenter.Error("Failed to write to output file %s: %v", exportOutputFile, err)
			return err
		}

		presenter.Newline()
		presenter.Success("Successfully exported %d issues to %s.", totalIssues, exportOutputFile)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectExportIssuesCmd)

	projectExportIssuesCmd.Flags().
		StringVarP(&exportOutputFile, "output", "o", "project_issues_export.md", "Path for the output markdown file")
}
