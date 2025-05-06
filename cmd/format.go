// cmd/format.go

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/tools" // For CommandExists and ExecuteCommand
	"github.com/contextvibes/cli/internal/ui"    // Use Presenter
	"github.com/spf13/cobra"
)

// Assume AppLogger is initialized in rootCmd for AI file logging
// var AppLogger *slog.Logger // Defined in root.go

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
	Example: `  contextvibes format  # Apply formatting to Go, Python, or Terraform files`,
	Args:    cobra.NoArgs,
	// Add Silence flags as we handle output/errors via Presenter
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background() // Context currently not used by tools.ExecuteCommand

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

		// Store formatting errors
		var formatErrors []error

		// --- Terraform Formatting ---
		if hasTerraform {
			presenter.Newline()
			presenter.Header("Terraform Formatting")
			tool := "terraform"
			if tools.CommandExists(tool) {
				presenter.Step("Running terraform fmt...")
				logger.Info("Executing terraform fmt -recursive .", slog.String("source_command", "format"))
				errFmt := tools.ExecuteCommand(cwd, tool, "fmt", "-recursive", ".")
				if errFmt != nil {
					errMsg := "`terraform fmt` failed"
					presenter.Error(errMsg + ": " + errFmt.Error()) // Show error details
					formatErrors = append(formatErrors, errors.New("terraform fmt failed"))
					logger.Error("Terraform fmt failed", slog.String("source_command", "format"), slog.String("error", errFmt.Error()))
				} else {
					presenter.Success("terraform fmt completed.")
					logger.Info("Terraform fmt successful", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Terraform formatting.", tool)
				presenter.Warning(msg) // Warning as it can't perform the action
				logger.Warn("Terraform format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", tool))
			}
		}

		// --- Python Formatting ---
		if hasPython {
			presenter.Newline()
			presenter.Header("Python Formatting")
			pythonDir := "." // Assuming checks run from root

			// --- isort ---
			toolIsort := "isort"
			if tools.CommandExists(toolIsort) {
				presenter.Step("Running %s...", toolIsort)
				logger.Info("Executing isort .", slog.String("source_command", "format"))
				errIsort := tools.ExecuteCommand(cwd, toolIsort, pythonDir)
				if errIsort != nil {
					errMsg := fmt.Sprintf("`%s` failed", toolIsort)
					presenter.Error(errMsg + ": " + errIsort.Error())
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

			// --- black ---
			// Run black even if isort failed, they format different things
			toolBlack := "black"
			if tools.CommandExists(toolBlack) {
				presenter.Step("Running %s...", toolBlack)
				logger.Info("Executing black .", slog.String("source_command", "format"))
				errBlack := tools.ExecuteCommand(cwd, toolBlack, pythonDir)
				if errBlack != nil {
					// Black exits non-zero if files are changed *or* if an error occurs.
					// We only consider it a critical error if the execution truly failed beyond just reformatting.
					// Since ExecuteCommand returns a generic error, we'll treat any error as critical for now.
					// A more nuanced check could capture stderr.
					errMsg := fmt.Sprintf("`%s` failed or reformatted files", toolBlack)
					presenter.Error(errMsg + ": " + errBlack.Error()) // Report as error for consistency
					formatErrors = append(formatErrors, errors.New("black failed or reformatted"))
					logger.Error("black failed or reformatted", slog.String("source_command", "format"), slog.String("error", errBlack.Error()))
				} else {
					presenter.Success("%s completed (no changes needed).", toolBlack)
					logger.Info("black successful (no changes)", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Python formatting.", toolBlack)
				presenter.Warning(msg)
				logger.Warn("black format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolBlack))
			}
		}

		// --- Go Formatting ---
		if hasGo {
			presenter.Newline()
			presenter.Header("Go Formatting")
			goDir := "./..." // Target all subdirectories

			toolGo := "go"
			if tools.CommandExists(toolGo) {
				// --- go fmt ---
				presenter.Step("Running go fmt...")
				logger.Info("Executing go fmt ./...", slog.String("source_command", "format"))
				errFmt := tools.ExecuteCommand(cwd, toolGo, "fmt", goDir)
				if errFmt != nil {
					errMsg := "`go fmt` failed"
					presenter.Error(errMsg + ": " + errFmt.Error())
					formatErrors = append(formatErrors, errors.New("go fmt failed"))
					logger.Error("go fmt failed", slog.String("source_command", "format"), slog.String("error", errFmt.Error()))
				} else {
					presenter.Success("go fmt completed.")
					logger.Info("go fmt successful", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Go formatting.", toolGo)
				presenter.Warning(msg)
				logger.Warn("Go format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolGo))
			}
		}

		// --- Summary ---
		presenter.Newline()
		presenter.Header("Formatting Summary")

		if len(formatErrors) > 0 {
			errMsg := fmt.Sprintf("%d formatting tool(s) reported errors.", len(formatErrors))
			presenter.Error(errMsg)
			presenter.Advice("Review the errors above.")
			// Return the first error to signal failure
			logger.Error("Format command failed due to errors", slog.String("source_command", "format"), slog.Int("error_count", len(formatErrors)))
			return formatErrors[0]
		}

		presenter.Success("All formatting tools completed successfully.")
		logger.Info("Format command successful", slog.String("source_command", "format"))
		return nil
	},
}

// init adds the command to the root command.
func init() {
	rootCmd.AddCommand(formatCmd)
}
