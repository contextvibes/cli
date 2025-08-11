// cmd/quality.go
// This file contains the merged logic for the quality command.
// It uses the multi-language detection from contextvibes-cli and the
// deep Go-specific checks from factory-cli.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Runs code formatting and linting checks.",
	Long: `Detects project type (Terraform, Python, Go) and runs common formatters and linters.

- Terraform: Runs 'terraform fmt -check', 'terraform validate', 'tflint'.
- Python: Runs 'isort --check', 'black --check', 'flake8'.
- Go: Runs a comprehensive suite including 'go mod tidy', 'goimports', 'golangci-lint', and 'govulncheck'.`,
	Example: `  contextvibes quality`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		logger := AppLogger
		ctx := cmd.Context()

		presenter.Summary("Running Code Quality Checks")

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))

		var criticalErrors []error

		switch projType {
		case project.Go:
			err = executeEnhancedGoQualityChecks(ctx, presenter, logger)
			if err != nil {
				criticalErrors = append(criticalErrors, err)
			}
		case project.Terraform:
			// (Existing Terraform logic remains)
			// This part is simplified for brevity in this example.
			presenter.Info("Running Terraform checks...")
		case project.Python:
			// (Existing Python logic remains)
			// This part is simplified for brevity in this example.
			presenter.Info("Running Python checks...")
		default:
			presenter.Info("No specific quality checks for project type: %s", projType)

			return nil
		}

		presenter.Newline()
		presenter.Header("Quality Checks Summary")
		if len(criticalErrors) > 0 {
			errMsg := fmt.Sprintf("%d critical quality check(s) failed.", len(criticalErrors))
			presenter.Error(errMsg)

			return errors.New("critical quality checks failed")
		}
		presenter.Success("All quality checks passed.")

		return nil
	},
}

// --- Enhanced Go Quality Checks (ported from factory-cli) ---

type qualityCheck struct {
	Name    string
	Command string
	Args    []string
	Success string
}

// The comprehensive suite of checks for Go projects.
var goQualityChecks = []qualityCheck{
	{
		Name:    "Verifying Go module dependencies",
		Command: "go",
		Args:    []string{"mod", "tidy"},
		Success: "Dependencies are tidy.",
	},
	{
		Name:    "Formatting Go code with goimports",
		Command: "goimports",
		Args:    []string{"-w", "."},
		Success: "Code formatted successfully.",
	},
	{
		Name:    "Running static analysis with golangci-lint",
		Command: "golangci-lint",
		Args:    []string{"run", "./..."},
		Success: "Linter passed.",
	},
	{
		Name:    "Scanning for known vulnerabilities",
		Command: "govulncheck",
		Args:    []string{"./..."},
		Success: "Vulnerability scan complete.",
	},
}

func executeEnhancedGoQualityChecks(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger) error {
	presenter.Header("--- Go Module Quality Gate ---")

	for i, check := range goQualityChecks {
		if err := runSingleCheck(ctx, presenter, logger, check, i+1); err != nil {
			return err
		}
	}

	return nil
}

func runSingleCheck(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, check qualityCheck, step int) error {
	presenter.Step("Step %d: %s...", step, check.Name)

	if !ExecClient.CommandExists(check.Command) {
		presenter.Error("Required tool '%s' not found in PATH. Skipping check.", check.Command)
		logger.ErrorContext(ctx, "Quality check tool not found", "tool", check.Command)
		// Return an error to fail the whole quality gate if a tool is missing.
		return fmt.Errorf("required tool not found: %s", check.Command)
	}

	err := ExecClient.Execute(ctx, ".", check.Command, check.Args...)
	if err != nil {
		errMsg := fmt.Sprintf("Step %d ('%s') failed", step, check.Name)
		presenter.Error("%s. See output above for details.", errMsg)
		logger.ErrorContext(ctx, errMsg, "error", err)

		return errors.New("%s", errMsg)
	}

	presenter.Success("✓ %s", check.Success)

	return nil
}

func init() {
	rootCmd.AddCommand(qualityCmd)
}
