// FILE: cmd/format.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Applies code formatting and auto-fixes linter issues.",
	Long: `Detects project type (Go, Python, Terraform) and applies standard formatting
and auto-fixable linter suggestions, modifying files in place. This is the primary
command for remediating code quality issues.

- Go: Runs 'golangci-lint run --fix', which applies all configured formatters and linters.
- Python: Runs 'isort .' and 'black .'.
- Terraform: Runs 'terraform fmt -recursive .'.`,
	Example: `  contextvibes format  # Apply formatting and fixes to the project`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		logger := AppLogger
		ctx := cmd.Context()

		presenter.Summary("Applying code formatting and auto-fixes.")

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

		var formatErrors []error

		switch projType {
		case project.Terraform:
			presenter.Header("Terraform Formatting")
			if err := runCommand(ctx, presenter, cwd, "terraform", []string{"fmt", "-recursive", "."}); err != nil {
				formatErrors = append(formatErrors, err)
			}
		case project.Python:
			presenter.Header("Python Formatting")
			if err := runCommand(ctx, presenter, cwd, "isort", []string{"."}); err != nil {
				formatErrors = append(formatErrors, err)
			}
			if err := runCommand(ctx, presenter, cwd, "black", []string{"."}); err != nil {
				// Black exits non-zero on reformat, which is not a critical error for this command.
				logger.InfoContext(
					ctx,
					"black completed with non-zero exit (likely reformatted files)",
					"error",
					err,
				)
			}
		case project.Go:
			presenter.Header("Go Formatting & Lint Fixes")
			err := runCommand(ctx, presenter, cwd, "golangci-lint", []string{"run", "--fix"})
			if err != nil {
				// A non-zero exit from the linter is informational for the format command.
				// It means it fixed what it could but unfixable issues remain.
				presenter.Warning(
					"'golangci-lint --fix' completed but reported issues that could not be auto-fixed. See output above.",
				)
				logger.InfoContext(
					ctx,
					"'golangci-lint --fix' completed with non-zero exit",
					"error",
					err,
				)
			} else {
				presenter.Success("✓ golangci-lint completed.")
			}
		}

		presenter.Newline()
		presenter.Header("Formatting Summary")
		if len(formatErrors) > 0 {
			errMsg := fmt.Sprintf(
				"%d formatting tool(s) reported critical errors.",
				len(formatErrors),
			)
			presenter.Error(errMsg)
			for _, failure := range formatErrors {
				presenter.Detail("- %v", failure)
			}
			return errors.New(errMsg)
		}

		presenter.Success("All formatting and auto-fixing tools completed.")
		return nil
	},
}

// runCommand is a helper to execute a single formatting tool.
// It returns an error for any non-zero exit code, which the caller must interpret.
func runCommand(
	ctx context.Context,
	presenter *ui.Presenter,
	cwd, command string,
	args []string,
) error {
	presenter.Step("Running %s...", command)
	if !ExecClient.CommandExists(command) {
		presenter.Warning("'%s' command not found, skipping.", command)
		return nil // Not a critical error, just skip
	}

	err := ExecClient.Execute(ctx, cwd, command, args...)
	if err != nil {
		// Do not print a success message if the command fails.
		// The Execute function already pipes the tool's output to the user.
		return fmt.Errorf("'%s' failed", command)
	}

	presenter.Success("✓ %s completed.", command)
	return nil
}

func init() {
	rootCmd.AddCommand(formatCmd)
}
