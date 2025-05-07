// cmd/format.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/contextvibes/cli/internal/project"
	// "github.com/contextvibes/cli/internal/tools" // No longer needed for exec functions
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
	// No direct import of internal/exec needed if using global ExecClient from cmd/root.go
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Applies code formatting (go fmt, terraform fmt, isort, black).",
	Long: `Detects project type (Go, Python, Terraform) and applies standard formatting
using available tools in PATH, modifying files in place.

- Go: Runs 'go fmt ./...'
- Python: Runs 'isort .' and 'black .'.
- Terraform: Runs 'terraform fmt -recursive .'.

This command focuses only on applying formatting, unlike 'quality' which checks
formatters, linters, and validators.`,
	Example:       `  contextvibes format  # Apply formatting to Go, Python, or Terraform files`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		// Use global ExecClient from cmd/root.go
		if ExecClient == nil {
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Applying code formatting.")

		cwd, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Format: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			logger.ErrorContext(ctx, "Format: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "format"), slog.String("type", string(projType)))

		hasTerraform := projType == project.Terraform
		hasPython := projType == project.Python
		hasGo := projType == project.Go

		if !hasTerraform && !hasPython && !hasGo {
			presenter.Info("No supported components (Terraform, Python, Go) found for formatting in this directory.")
			return nil
		}

		var formatErrors []error

		// --- Terraform Formatting ---
		if hasTerraform {
			presenter.Newline()
			presenter.Header("Terraform Formatting")
			tool := "terraform"
			if ExecClient.CommandExists(tool) { // Use ExecClient
				presenter.Step("Running terraform fmt...")
				logger.Info("Executing terraform fmt -recursive .", slog.String("source_command", "format"))
				// terraform fmt pipes its own output (files changed)
				errFmt := ExecClient.Execute(ctx, cwd, tool, "fmt", "-recursive", ".") // Use ExecClient
				if errFmt != nil {
					errMsg := fmt.Sprintf("`terraform fmt` failed or reported issues. Error: %v", errFmt)
					presenter.Error(errMsg)
					formatErrors = append(formatErrors, errors.New("terraform fmt failed"))
					logger.Error("Terraform fmt failed", slog.String("source_command", "format"), slog.String("error", errFmt.Error()))
				} else {
					presenter.Success("terraform fmt completed.")
					logger.Info("Terraform fmt successful", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Terraform formatting.", tool)
				presenter.Warning(msg)
				logger.Warn("Terraform format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", tool))
			}
		}

		// --- Python Formatting ---
		if hasPython {
			presenter.Newline()
			presenter.Header("Python Formatting")
			pythonDir := "."

			toolIsort := "isort"
			if ExecClient.CommandExists(toolIsort) { // Use ExecClient
				presenter.Step("Running %s...", toolIsort)
				logger.Info("Executing isort .", slog.String("source_command", "format"))
				errIsort := ExecClient.Execute(ctx, cwd, toolIsort, pythonDir) // Use ExecClient
				if errIsort != nil {
					errMsg := fmt.Sprintf("`%s` failed or reported issues. Error: %v", toolIsort, errIsort)
					presenter.Error(errMsg)
					formatErrors = append(formatErrors, errors.New("isort failed"))
					logger.Error("isort failed", slog.String("source_command", "format"), slog.String("error", errIsort.Error()))
				} else {
					presenter.Success("%s completed.", toolIsort)
					logger.Info("isort successful", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping import sorting.", toolIsort)
				presenter.Warning(msg)
				logger.Warn("isort format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolIsort))
			}

			toolBlack := "black"
			if ExecClient.CommandExists(toolBlack) { // Use ExecClient
				presenter.Step("Running %s...", toolBlack)
				logger.Info("Executing black .", slog.String("source_command", "format"))
				// Black exits 0 if no changes, 1 if reformatted, >1 on error.
				// ExecClient.Execute will return an error for non-zero exit.
				// We can interpret this: if no error, no changes. If error, could be reformat or actual fail.
				// For `format` command, successful reformatting is a success.
				// The `OSCommandExecutor` logs the exit code, so we can rely on its error message or check stderr.
				// For simplicity, we treat any non-zero exit from black as "files were formatted or error occurred".
				// The user sees black's direct output.
				errBlack := ExecClient.Execute(ctx, cwd, toolBlack, pythonDir) // Use ExecClient
				if errBlack != nil {
					// Check if it's just a reformatting (exit code 1 for black typically means files changed)
					// This requires more complex error inspection if we want to distinguish.
					// For now, if black exits non-zero, we log it as potentially having issues.
					// A more robust solution might use CaptureOutput and inspect exit code and stderr.
					errMsg := fmt.Sprintf("`%s` completed (may have reformatted files or encountered an issue). Error (if any): %v", toolBlack, errBlack)
					presenter.Info(errMsg) // Info, as reformatting is the goal. If actual error, black would show it.
					logger.Warn("black completed with non-zero exit", slog.String("source_command", "format"), slog.String("error", errBlack.Error()))
					// Don't add to formatErrors unless we are sure it's a critical failure, not just reformatting.
					// If it's a critical failure, black's output to stderr (piped by Execute) should indicate it.
				} else {
					presenter.Success("%s completed (no changes needed).", toolBlack)
					logger.Info("black successful (no changes)", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Python code formatting.", toolBlack)
				presenter.Warning(msg)
				logger.Warn("black format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolBlack))
			}
		}

		// --- Go Formatting ---
		if hasGo {
			presenter.Newline()
			presenter.Header("Go Formatting")
			goDir := "./..."

			toolGo := "go"
			if ExecClient.CommandExists(toolGo) { // Use ExecClient
				presenter.Step("Running go fmt...")
				logger.Info("Executing go fmt ./...", slog.String("source_command", "format"))
				// `go fmt` prints changed file paths to stdout.
				// We can use CaptureOutput to see if it did anything.
				stdout, stderr, errFmt := ExecClient.CaptureOutput(ctx, cwd, toolGo, "fmt", goDir) // Use ExecClient
				if errFmt != nil {
					errMsg := fmt.Sprintf("`go fmt` failed. Error: %v", errFmt)
					if stderr != "" {
						errMsg += fmt.Sprintf("\nStderr: %s", stderr)
					}
					presenter.Error(errMsg)
					formatErrors = append(formatErrors, errors.New("go fmt failed"))
					logger.Error("go fmt failed", slog.String("source_command", "format"), slog.String("error", errFmt.Error()), slog.String("stderr", stderr))
				} else {
					if stdout != "" {
						presenter.Success("go fmt completed and formatted the following files:")
						presenter.Detail(stdout) // Show which files were formatted
					} else {
						presenter.Success("go fmt completed (no files needed formatting).")
					}
					logger.Info("go fmt successful", slog.String("source_command", "format"), slog.String("stdout", stdout))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Go formatting.", toolGo)
				presenter.Warning(msg)
				logger.Warn("Go format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolGo))
			}
		}

		presenter.Newline()
		presenter.Header("Formatting Summary")
		if len(formatErrors) > 0 {
			errMsg := fmt.Sprintf("%d formatting tool(s) reported errors.", len(formatErrors))
			presenter.Error(errMsg)
			presenter.Advice("Review the errors above.")
			logger.Error("Format command failed due to errors", slog.String("source_command", "format"), slog.Int("error_count", len(formatErrors)))
			return formatErrors[0]
		}

		presenter.Success("All formatting tools completed successfully or applied changes.")
		logger.Info("Format command finished", slog.String("source_command", "format"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)
}
