// cmd/product/quality/quality.go
package quality

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed quality.md.tpl
var qualityLongDescription string

// QualityCmd represents the quality command.
var QualityCmd = &cobra.Command{
	Use:           "quality",
	Example:       `  contextvibes product quality`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
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
			failures := executeEnhancedGoQualityChecks(ctx, presenter, globals.ExecClient)
			if len(failures) > 0 {
				criticalErrors = append(criticalErrors, failures...)
			}
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

type qualityCheck struct {
	Name       string
	Command    string
	Args       []string
	SuccessMsg string
	FailAdvice string
}

func executeEnhancedGoQualityChecks(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
) []string {
	var failures []string

	goQualityChecks := []qualityCheck{
		{
			Name:       "Verifying Go module dependencies are tidy",
			Command:    "go",
			Args:       []string{"mod", "tidy"},
			SuccessMsg: "Dependencies are tidy.",
			FailAdvice: "Run 'go mod tidy' or 'contextvibes product format' and commit the changes.",
		},
		{
			Name:       "Checking for suspicious constructs with go vet",
			Command:    "go",
			Args:       []string{"vet", "./..."},
			SuccessMsg: "Code passes go vet.",
			FailAdvice: "Run 'go vet ./...' to see and fix the reported issues.",
		},
		{
			Name:       "Running static analysis with golangci-lint",
			Command:    "golangci-lint",
			Args:       []string{"run"},
			SuccessMsg: "Linter passed (includes formatting checks).",
			FailAdvice: "Review the linter output above to fix issues, or run 'contextvibes product format' to apply auto-fixes.",
		},
		{
			Name:       "Scanning for known vulnerabilities",
			Command:    "govulncheck",
			Args:       []string{"./..."},
			SuccessMsg: "No known vulnerabilities found.",
			FailAdvice: "Review the vulnerability report above and update dependencies as needed.",
		},
	}

	for _, check := range goQualityChecks {
		presenter.Step("Running check: %s...", check.Name)

		if !execClient.CommandExists(check.Command) {
			errMsg := fmt.Sprintf("Required tool '%s' not found in PATH.", check.Command)
			presenter.Error(errMsg)
			failures = append(failures, errMsg)

			continue
		}

		err := execClient.Execute(ctx, ".", check.Command, check.Args...)
		if err != nil {
			presenter.Error("! Check failed: %s", check.Name)
			presenter.Advice(check.FailAdvice)
			failures = append(failures, fmt.Sprintf("'%s' failed", check.Name))
		} else {
			presenter.Success("✓ %s", check.SuccessMsg)
		}

		presenter.Newline()
	}

	return failures
}

func init() {
	desc, err := cmddocs.ParseAndExecute(qualityLongDescription, nil)
	if err != nil {
		panic(err)
	}

	QualityCmd.Short = desc.Short
	QualityCmd.Long = desc.Long
}
