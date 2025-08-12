// cmd/project.go
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var outputFile string // Flag variable

// --- Data Structures for GitHub JSON Output ---
type Issue struct {
	Number   int       `json:"number"`
	Title    string    `json:"title"`
	Author   Author    `json:"author"`
	Body     string    `json:"body"`
	Comments []Comment `json:"comments"`
}

type Author struct {
	Login string `json:"login"`
}

type Comment struct {
	Author Author `json:"author"`
	Body   string `json:"body"`
}

// WriteTo conforms to the io.WriterTo interface, which is a Go standard.
func (i *Issue) WriteTo(w io.Writer) (n int64, err error) {
	// Use a temporary presenter to format output to the provided writer.
	presenter := ui.NewPresenter(w, io.Discard, nil)
	var totalBytes int

	// Helper to write and track byte count
	write := func(format string, a ...any) {
		if err != nil {
			return
		}
		var written int
		written, err = fmt.Fprintf(w, format, a...)
		totalBytes += written
	}

	presenter.Header(fmt.Sprintf("#%d %s", i.Number, i.Title))
	presenter.Detail("Author: %s", i.Author.Login)
	presenter.Newline()
	presenter.Step("Body:")
	write("%s\n\n", i.Body)

	if len(i.Comments) > 0 {
		presenter.Step("Comments (%d):", len(i.Comments))
		for _, comment := range i.Comments {
			presenter.Separator()
			presenter.Detail("Comment by %s:", comment.Author.Login)
			write("%s\n", comment.Body)
		}
	}
	presenter.Separator()

	return int64(totalBytes), err
}

// --- Cobra Command Definitions ---

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manages interactions with the Git hosting provider (e.g., GitHub).",
	Long:  `Provides commands to interact with project management features of the Git hosting provider, such as issues and pull requests.`,
}

var listIssuesCmd = &cobra.Command{
	Use:   "list-issues",
	Short: "Fetches and displays all open issues for the current repository.",
	Long:  `Fetches and displays all open issues, including their full body and all comments, from the current GitHub repository using the 'gh' CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Presenter for console messages (status, errors)
		consolePresenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context()

		// Determine the target for the main content output
		var outputTarget io.WriteCloser = os.Stdout

		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				consolePresenter.Error("Failed to create output file '%s': %v", outputFile, err)
				return err
			}
			outputTarget = file
			defer func() { _ = outputTarget.Close() }()
		}

		consolePresenter.Summary("Fetching GitHub issues for the current repository...")

		if !ExecClient.CommandExists("gh") {
			consolePresenter.Error("GitHub CLI ('gh') not found. This command is required.")
			consolePresenter.Advice(
				"Please install it from https://cli.github.com/ and authenticate with 'gh auth login'.",
			)
			return errors.New("gh cli not found")
		}

		consolePresenter.Step("Running 'gh issue list'...")
		jsonFields := "number,title,author,body,comments"
		stdout, stderr, err := ExecClient.CaptureOutput(
			ctx,
			".",
			"gh",
			"issue",
			"list",
			"--json",
			jsonFields,
		)
		if err != nil {
			consolePresenter.Error("Failed to fetch issues from GitHub CLI.")
			consolePresenter.Detail("Error: %v", err)
			if stderr != "" {
				consolePresenter.Detail("Stderr: %s", stderr)
			}
			consolePresenter.Advice(
				"Ensure you are in a GitHub repository and have run 'gh auth login'.",
			)
			return errors.New("gh issue list command failed")
		}

		var issues []Issue
		if err := json.Unmarshal([]byte(stdout), &issues); err != nil {
			consolePresenter.Error("Failed to parse JSON output from GitHub CLI: %v", err)
			return err
		}

		if len(issues) == 0 {
			consolePresenter.Info("No open issues found for this repository.")
			return nil
		}

		// Write the formatted issues to the designated target (console or file)
		for _, issue := range issues {
			_, _ = issue.WriteTo(outputTarget)
		}

		if outputFile != "" {
			consolePresenter.Success(
				"Successfully wrote %d issue(s) to %s",
				len(issues),
				outputFile,
			)
		}

		return nil
	},
}

func init() {
	listIssuesCmd.Flags().
		StringVarP(&outputFile, "output", "o", "", "Path to save the output file.")
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(listIssuesCmd)
}
