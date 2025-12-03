// Package quality provides the command to run code quality checks.
package quality

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/pipeline"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed quality.md.tpl
var qualityLongDescription string

const contextFile = "_contextvibes.md"

// QualityCmd represents the quality command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var QualityCmd = &cobra.Command{
	Use:           "quality",
	Example:       `  contextvibes product quality`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Running Code Quality Pipeline")

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		presenter.Newline()

		// Initialize Pipeline Runner
		runner := pipeline.NewRunner(presenter, globals.ExecClient)
		var checks []pipeline.Check

		// Assemble Pipeline based on Project Type
		switch projType {
		case project.Go:
			checks = []pipeline.Check{
				&pipeline.GoModTidyCheck{},
				&pipeline.GoVetCheck{},
				&pipeline.GolangCILintCheck{},
				&pipeline.GoVulnCheck{},
				&pipeline.GitleaksCheck{},
				&pipeline.DeadcodeCheck{},
			}
		case project.Terraform, project.Pulumi, project.Python, project.Unknown:
			fallthrough
		default:
			presenter.Info("No specific quality checks configured for %s.", projType)

			return nil
		}

		// Execute Pipeline
		results, err := runner.Run(ctx, checks)

		// Logic: If issues found, generate report. If success, clean up stale report.
		if err != nil || hasWarnings(results) {
			if genErr := generateContextFile(results); genErr != nil {
				presenter.Error("Failed to generate context file: %v", genErr)
			} else {
				presenter.Newline()
				presenter.Info("Generated AI Context: %s", contextFile)
				presenter.Advice("Pass this file to your AI to fix the issues.")
			}
		} else {
			// Cleanup: If all passed, remove the stale context file if it exists.
			// Uber Style: Inline check
			if _, err := os.Stat(contextFile); err == nil {
				if removeErr := os.Remove(contextFile); removeErr == nil {
					presenter.Info("Removed stale AI Context file: %s (all checks passed)", contextFile)
				}
			}
		}

		if err != nil {
			presenter.Error("Pipeline failed.")

			//nolint:wrapcheck // Wrapping is handled by caller.
			return err
		}

		presenter.Success("All quality checks passed.")

		return nil
	},
}

func hasWarnings(results []pipeline.Result) bool {
	for _, result := range results {
		switch result.Status {
		case pipeline.StatusWarn, pipeline.StatusFail:
			return true
		case pipeline.StatusPass:
			continue
		}
	}

	return false
}

func generateContextFile(results []pipeline.Result) error {
	var buf bytes.Buffer

	// 1. The Prompt
	buf.WriteString("# AI Task: Fix Quality Issues\n\n")
	buf.WriteString("You are a senior software engineer. Analyze the quality report below.\n")
	buf.WriteString("Your goal is to fix the **Linter Errors** and address the **Dead Code** warnings.\n\n")
	buf.WriteString("## Instructions\n")
	buf.WriteString("1.  **Analyze**: Look at the specific error messages and file paths.\n")
	buf.WriteString("2.  **Plan**: Create a plan to resolve each issue.\n")
	buf.WriteString("3.  **Execute**: Provide the code changes (using `cat` scripts or `sed`) to fix the codebase.\n")
	buf.WriteString("4.  **Verify**: Remind me to run `contextvibes product quality` again.\n\n")

	buf.WriteString("---\n\n")

	// 2. The Report
	buf.WriteString(fmt.Sprintf("# Quality Report (%s)\n\n", time.Now().Format(time.RFC3339)))

	for _, result := range results {
		// Use standard ContextVibes markers instead of icons
		marker := "+"

		// Use switch for status check (fixes staticcheck QF1003)
		switch result.Status {
		case pipeline.StatusFail:
			marker = "!"
		case pipeline.StatusWarn:
			marker = "~"
		case pipeline.StatusPass:
			marker = "+" // Explicitly handle Pass to satisfy exhaustive
		}

		buf.WriteString(fmt.Sprintf("## %s %s\n", marker, result.Name))
		buf.WriteString(fmt.Sprintf("**Status:** %s\n", statusToString(result.Status)))

		if result.Message != "" {
			buf.WriteString(fmt.Sprintf("**Message:** %s\n", result.Message))
		}

		if result.Details != "" {
			buf.WriteString("\n**Details:**\n")
			buf.WriteString("text\n")
			buf.WriteString(strings.TrimSpace(result.Details))
			buf.WriteString("\n\n")
		}

		if result.Error != nil {
			buf.WriteString(fmt.Sprintf("\n**System Error:** %v\n", result.Error))
		}

		buf.WriteString("\n")
	}

	//nolint:mnd // 0600 is standard secure file permission.
	if err := os.WriteFile(contextFile, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

func statusToString(s pipeline.Status) string {
	switch s {
	case pipeline.StatusPass:
		return "Pass"
	case pipeline.StatusFail:
		return "Fail"
	case pipeline.StatusWarn:
		return "Warning"
	default:
		return "Unknown"
	}
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(qualityLongDescription, nil)
	if err != nil {
		panic(err)
	}

	QualityCmd.Short = desc.Short
	QualityCmd.Long = desc.Long
}
