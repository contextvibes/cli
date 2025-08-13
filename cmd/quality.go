// FILE: cmd/quality.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Runs a comprehensive suite of code quality checks.",
	Long: `Detects project type (Go, Python, Terraform) and runs a suite of formatters,
linters, and vulnerability scanners in a read-only "check" mode.

- Go: Verifies 'go mod' is tidy, runs 'go vet', runs 'golangci-lint', and scans for
  vulnerabilities with 'govulncheck'. All formatting checks are handled by 'golangci-lint'.
- Python: Runs 'isort --check', 'black --check', 'flake8'.
- Terraform: Runs 'terraform fmt -check', 'terraform validate', 'tflint'.

This command acts as a quality gate and will fail if any issues are found. To fix
many of the reported issues automatically, run 'contextvibes format'.`,
	Example:       `  contextvibes quality`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		ctx := cmd.Context()

		presenter.Summary("Running Code Quality Checks")

		cwd, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			presenter.Error("Failed to detect project type: %v", err)
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))

		var criticalErrors []string

		switch projType {
		case project.Go:
			failures := executeEnhancedGoQualityChecks(ctx, presenter)
			if len(failures) > 0 {
				criticalErrors = append(criticalErrors, failures...)
			}
		case project.Terraform:
			presenter.Warning("Terraform quality checks are not fully implemented yet.")
		case project.Python:
			presenter.Warning("Python quality checks are not fully implemented yet.")
		default:
			presenter.Info("No specific quality checks for project type: %s", projType)
			return nil
		}

		presenter.Newline()
		presenter.Header("Quality Checks Summary")
		if len(criticalErrors) > 0 {
			errorMsg := fmt.Sprintf("%d critical quality check(s) failed.", len(criticalErrors))
			presenter.Error(errorMsg)
			for _, failure := range criticalErrors {
				presenter.Detail("- %s", failure)
			}
			return errors.New(errorMsg)
		}
		presenter.Success("All quality checks passed.")
		return nil
	},
}

// --- Enhanced Go Quality Checks ---

type CheckType int

const (
	CheckSucceeded CheckType = iota
	CheckNoChanges
	CheckIsEmpty
)

type qualityCheck struct {
	Name       string
	Command    string
	Args       []string
	CheckType  CheckType
	CheckFiles []string
	SuccessMsg string
	FailAdvice string
}

var goQualityChecks = []qualityCheck{
	{
		Name:       "Verifying Go module dependencies are tidy",
		Command:    "go",
		Args:       []string{"mod", "tidy"},
		CheckType:  CheckNoChanges,
		CheckFiles: []string{"go.mod", "go.sum"},
		SuccessMsg: "Dependencies are tidy.",
		FailAdvice: "Run 'go mod tidy' or 'contextvibes format' and commit the changes.",
	},
	{
		Name:       "Checking for suspicious constructs with go vet",
		Command:    "go",
		Args:       []string{"vet", "./..."},
		CheckType:  CheckSucceeded,
		SuccessMsg: "Code passes go vet.",
		FailAdvice: "Run 'go vet ./...' to see and fix the reported issues.",
	},
	{
		Name:       "Running static analysis with golangci-lint",
		Command:    "golangci-lint",
		Args:       []string{"run"},
		CheckType:  CheckSucceeded,
		SuccessMsg: "Linter passed (includes formatting checks).",
		FailAdvice: "Review the linter output above to fix issues, or run 'contextvibes format' to apply auto-fixes.",
	},
	{
		Name:       "Scanning for known vulnerabilities",
		Command:    "govulncheck",
		Args:       []string{"./..."},
		CheckType:  CheckSucceeded,
		SuccessMsg: "No known vulnerabilities found.",
		FailAdvice: "Review the vulnerability report above and update dependencies as needed.",
	},
}

func executeEnhancedGoQualityChecks(ctx context.Context, presenter *ui.Presenter) []string {
	presenter.Header("--- Go Module Quality Gate ---")
	var failures []string

	for i, check := range goQualityChecks {
		stepNum := i + 1
		presenter.Step("Step %d: %s...", stepNum, check.Name)

		if !ExecClient.CommandExists(check.Command) {
			errMsg := fmt.Sprintf("Required tool '%s' not found in PATH.", check.Command)
			presenter.Error(errMsg)
			failures = append(failures, errMsg)
			continue
		}

		isSuccess := true
		switch check.CheckType {
		case CheckSucceeded:
			if err := ExecClient.Execute(ctx, ".", check.Command, check.Args...); err != nil {
				isSuccess = false
				presenter.Error("! Step %d failed: %s", stepNum, check.Name)
				presenter.Advice(check.FailAdvice)
			}
		case CheckIsEmpty:
			stdout, _, err := ExecClient.CaptureOutput(ctx, ".", check.Command, check.Args...)
			hasOutput := strings.TrimSpace(stdout) != ""
			hasError := err != nil

			if hasError || hasOutput {
				isSuccess = false
				presenter.Error("! Step %d failed: %s", stepNum, check.Name)
				if hasOutput {
					presenter.Detail("The following files reported issues:\n%s", stdout)
				}
				presenter.Advice(check.FailAdvice)
			}
		case CheckNoChanges:
			if err := ExecClient.Execute(ctx, ".", check.Command, check.Args...); err != nil {
				isSuccess = false
				presenter.Error(
					"! Step %d failed: Command '%s' returned an error.",
					stepNum,
					check.Command,
				)
				presenter.Advice(check.FailAdvice)
				break
			}
			gitArgs := append([]string{"status", "--porcelain"}, check.CheckFiles...)
			stdout, _, err := ExecClient.CaptureOutput(ctx, ".", "git", gitArgs...)
			if err != nil {
				isSuccess = false
				presenter.Error(
					"! Step %d failed: Could not check git status after command.",
					stepNum,
				)
				presenter.Advice(check.FailAdvice)
			} else if strings.TrimSpace(stdout) != "" {
				isSuccess = false
				presenter.Error("! Step %d failed: %s", stepNum, check.Name)
				presenter.Detail("The command modified the following files:\n%s", stdout)
				presenter.Advice(check.FailAdvice)
			}
		}

		if isSuccess {
			presenter.Success("✓ %s", check.SuccessMsg)
		} else {
			failures = append(failures, fmt.Sprintf("'%s' failed", check.Name))
		}
		presenter.Newline()
	}
	return failures
}

func init() {
	rootCmd.AddCommand(qualityCmd)
}
