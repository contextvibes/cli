// cmd/quality.go

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings" // Added for trimming space

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"    // Use Presenter
	"github.com/spf13/cobra"
)

// Assume AppLogger is initialized in rootCmd for AI file logging
// var AppLogger *slog.Logger // Defined in root.go

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Runs code formatting and linting checks (Terraform, Python, Go).",
	Long: `Detects project type (Terraform, Python, Go) and runs common formatters and linters.
Checks performed depend on available tools in PATH.

- Terraform: Runs 'terraform fmt -check', 'terraform validate', 'tflint'.
- Python: Runs 'isort --check', 'black --check', 'flake8'.
- Go: Runs 'go vet', 'go mod tidy', and checks 'go fmt' compliance.

Formatter/validator checks ('terraform fmt -check', 'terraform validate', 'isort --check',
'black --check', 'go vet') and dependency checks ('go mod tidy') will fail the command
if issues are found or errors occur.
The 'go fmt' check will also fail the command if files *are not* correctly formatted
(note: this check modifies files in place if needed to determine compliance).
Linter issues ('tflint', 'flake8') are reported as warnings.`,
	Example: `  contextvibes quality`,
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
		ctx := context.Background()

		presenter.Summary("Running code quality checks.")

		cwd, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Quality: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			logger.ErrorContext(ctx, "Quality: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "quality"), slog.String("type", string(projType)))

		hasTerraform := projType == project.Terraform
		hasPython := projType == project.Python
		hasGo := projType == project.Go

		if !hasTerraform && !hasPython && !hasGo {
			presenter.Info("No supported components (Terraform, Python, Go) found for quality checks in this directory.")
			return nil
		}

		var criticalErrors []error
		var warnings []string

		// TODO: Offload specific checks for each project type into dedicated internal functions/packages.

		// --- Terraform Checks ---
		if hasTerraform {
			presenter.Newline()
			presenter.Header("Terraform Quality Checks")
			tool := "terraform"
			if ExecClient.CommandExists(tool) {
				// --- terraform fmt -check ---
				presenter.Step("Checking Terraform formatting (terraform fmt -check)...")
				logger.Info("Executing terraform fmt -check -recursive .", slog.String("source_command", "quality"))
				errFmt := ExecClient.Execute(ctx, cwd, tool, "fmt", "-check", "-recursive", ".")
				if errFmt != nil {
					errMsg := "`terraform fmt -check` failed or found files needing formatting"
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errFmt))
					criticalErrors = append(criticalErrors, errors.New("terraform fmt check failed"))
					logger.Error("Terraform fmt check failed", slog.String("source_command", "quality"), slog.String("error", errFmt.Error()))
					presenter.Advice("Run `contextvibes format` or `terraform fmt -recursive .` to fix formatting.")
				} else {
					presenter.Success("terraform fmt check completed (no changes needed).")
					logger.Info("Terraform fmt check successful", slog.String("source_command", "quality"))
				}

				// --- terraform validate ---
				if errFmt == nil { // Skip if formatting failed
					presenter.Step("Running terraform validate...")
					logger.Info("Executing terraform validate", slog.String("source_command", "quality"))
					errValidate := ExecClient.Execute(ctx, cwd, tool, "validate")
					if errValidate != nil {
						errMsg := "`terraform validate` failed"
						presenter.Error(errMsg)
						warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errValidate))
						criticalErrors = append(criticalErrors, errors.New("terraform validate failed"))
						logger.Error("Terraform validate failed", slog.String("source_command", "quality"), slog.String("error", errValidate.Error()))
					} else {
						presenter.Success("terraform validate completed.")
						logger.Info("Terraform validate successful", slog.String("source_command", "quality"))
					}
				} else {
					presenter.Warning("Skipping terraform validate due to previous terraform fmt check failure.")
					logger.Warn("Skipping terraform validate due to fmt failure", slog.String("source_command", "quality"))
				}
			} else {
				// Handle missing terraform tool
				msg := fmt.Sprintf("'%s' command not found, skipping Terraform format/validate.", tool)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("Terraform checks skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", tool))
			}

			// --- tflint ---
			linter := "tflint"
			if ExecClient.CommandExists(linter) {
				presenter.Step("Running %s...", linter)
				logger.Info("Executing tflint", slog.String("source_command", "quality"))
				errLint := ExecClient.Execute(ctx, cwd, linter, "--recursive", ".")
				if errLint != nil {
					errMsg := fmt.Sprintf("`%s` reported issues or failed", linter)
					presenter.Warning(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errLint)) // Add to warnings, not critical
					logger.Warn("tflint reported issues or failed", slog.String("source_command", "quality"), slog.String("error", errLint.Error()))
				} else {
					presenter.Success("%s completed (no issues found).", linter)
					logger.Info("tflint successful", slog.String("source_command", "quality"))
				}
			} else {
				presenter.Info("'%s' command not found, skipping Terraform linting.", linter) // Info, as it's just a linter
				logger.Info("tflint check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", linter))
			}
		}

		// --- Python Checks ---
		if hasPython {
			presenter.Newline()
			presenter.Header("Python Quality Checks")
			pythonDir := "." // Assuming checks run from root

			// --- isort --check ---
			toolIsort := "isort"
			if ExecClient.CommandExists(toolIsort) {
				presenter.Step("Checking import sorting (%s --check)...", toolIsort)
				logger.Info("Executing isort --check", slog.String("source_command", "quality"))
				errIsort := ExecClient.Execute(ctx, cwd, toolIsort, "--check", pythonDir)
				if errIsort != nil {
					errMsg := fmt.Sprintf("`%s --check` failed or found files needing sorting", toolIsort)
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errIsort))
					criticalErrors = append(criticalErrors, errors.New("isort check failed"))
					logger.Error("isort check failed", slog.String("source_command", "quality"), slog.String("error", errIsort.Error()))
					presenter.Advice("Run `contextvibes format` or `isort .` to fix import sorting.")
				} else {
					presenter.Success("%s check completed (imports sorted).", toolIsort)
					logger.Info("isort check successful", slog.String("source_command", "quality"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping import sorting check.", toolIsort)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("isort check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", toolIsort))
			}

			// --- black --check ---
			toolBlack := "black"
			if ExecClient.CommandExists(toolBlack) {
				isortCheckFailed := false
				for _, e := range criticalErrors {
					if e.Error() == "isort check failed" {
						isortCheckFailed = true
						break
					}
				}
				if !isortCheckFailed {
					presenter.Step("Checking Python formatting (%s --check)...", toolBlack)
					logger.Info("Executing black --check", slog.String("source_command", "quality"))
					errBlack := ExecClient.Execute(ctx, cwd, toolBlack, "--check", pythonDir)
					if errBlack != nil {
						errMsg := fmt.Sprintf("`%s --check` failed or found files needing formatting", toolBlack)
						presenter.Error(errMsg)
						warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errBlack))
						criticalErrors = append(criticalErrors, errors.New("black check failed"))
						logger.Error("black check failed", slog.String("source_command", "quality"), slog.String("error", errBlack.Error()))
						presenter.Advice("Run `contextvibes format` or `black .` to fix formatting.")
					} else {
						presenter.Success("%s check completed (no changes needed).", toolBlack)
						logger.Info("black check successful", slog.String("source_command", "quality"))
					}
				} else {
					presenter.Warning("Skipping %s check due to previous python tool failure.", toolBlack)
					logger.Warn("Skipping black check due to prior failure", slog.String("source_command", "quality"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Python formatting check.", toolBlack)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("black check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", toolBlack))
			}

			// --- flake8 ---
			linterFlake8 := "flake8"
			if ExecClient.CommandExists(linterFlake8) {
				presenter.Step("Running %s...", linterFlake8)
				logger.Info("Executing flake8", slog.String("source_command", "quality"))
				errFlake8 := ExecClient.Execute(ctx, cwd, linterFlake8, pythonDir)
				if errFlake8 != nil {
					errMsg := fmt.Sprintf("`%s` reported issues or failed", linterFlake8)
					presenter.Warning(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errFlake8)) // Warning only
					logger.Warn("flake8 reported issues or failed", slog.String("source_command", "quality"), slog.String("error", errFlake8.Error()))
				} else {
					presenter.Success("%s completed (no issues found).", linterFlake8)
					logger.Info("flake8 successful", slog.String("source_command", "quality"))
				}
			} else {
				presenter.Info("'%s' command not found, skipping Python linting.", linterFlake8)
				logger.Info("flake8 check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", linterFlake8))
			}
		}

		// --- Go Checks ---
		if hasGo {
			presenter.Newline()
			presenter.Header("Go Quality Checks")
			goDir := "./..." // Target all subdirectories for Go tools

			toolGo := "go"
			if ExecClient.CommandExists(toolGo) {

				// --- go fmt ---
				// Check formatting compliance by running `go fmt` and capturing output.
				// Note: This command *modifies files in place* if they are not formatted.
				// It acts as both a check and a fix within the quality command for Go fmt.
				// We treat non-empty output (files were formatted) as a critical error
				// because the code was not compliant before the command ran.
				// TODO: Revisit this if a reliable check-only mode becomes standard in go fmt
				//       or if external formatters (like gofumpt -l) are adopted.
				presenter.Step("Checking Go formatting (running go fmt)...")
				logger.Info("Executing go fmt ./... (and checking output)", slog.String("source_command", "quality"))
				fmtOutput, fmtStderr, errFmt := ExecClient.CaptureOutput(ctx, cwd, toolGo, "fmt", goDir)

				// First, check for execution errors (e.g., syntax errors).
				if errFmt != nil {
					errMsg := "`go fmt` execution failed"
					presenter.Error(errMsg + ": " + errFmt.Error())
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errFmt))             // Log warning for summary
					criticalErrors = append(criticalErrors, errors.New("go fmt execution failed")) // Critical error
					logger.Error("go fmt execution failed", slog.String("source_command", "quality"), slog.String("error", errFmt.Error()), slog.String("stderr", fmtStderr))
				} else {
					// If execution succeeded, check if files were actually formatted (non-empty output).
					trimmedOutput := strings.TrimSpace(fmtOutput)
					if trimmedOutput != "" {
						// Files *were* formatted, meaning they were not compliant initially. Treat as critical error.
						errMsg := "Go files were not correctly formatted (fixed by `go fmt`)"
						presenter.Error(errMsg)
						warnings = append(warnings, errMsg)                                             // Add to warnings for summary visibility
						criticalErrors = append(criticalErrors, errors.New("go fmt compliance failed")) // Critical error
						logger.Error("go fmt compliance failed: files were modified", slog.String("source_command", "quality"), slog.String("files_formatted", trimmedOutput))
						presenter.Advice("Commit the formatting changes applied by `go fmt`.")
						// Optional: show which files were formatted using presenter.Detail(trimmedOutput)
					} else {
						// Success: command ran without error and produced no output, meaning files were already formatted.
						presenter.Success("go fmt check completed (files already formatted).")
						logger.Info("go fmt check successful (no changes needed)", slog.String("source_command", "quality"))
					}
				}

				// --- go vet ---
				// Checks for suspicious constructs. Failure is critical.
				presenter.Step("Running Go vet...")
				logger.Info("Executing go vet ./...", slog.String("source_command", "quality"))
				errVet := ExecClient.Execute(ctx, cwd, toolGo, "vet", goDir)
				if errVet != nil {
					errMsg := "`go vet` reported issues or failed"
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errVet))
					criticalErrors = append(criticalErrors, errors.New("go vet failed"))
					logger.Error("go vet reported issues or failed", slog.String("source_command", "quality"), slog.String("error", errVet.Error()))
				} else {
					presenter.Success("go vet completed (no issues found).")
					logger.Info("go vet successful", slog.String("source_command", "quality"))
				}

				// --- go mod tidy ---
				// Ensures go.mod and go.sum are consistent. Failure is critical.
				// TODO: Add check if go.mod or go.sum were modified by tidy, potentially make that a critical error too.
				//       This likely requires checking git status before/after or diffing files.
				presenter.Step("Running go mod tidy...")
				logger.Info("Executing go mod tidy", slog.String("source_command", "quality"))
				errTidy := ExecClient.Execute(ctx, cwd, toolGo, "mod", "tidy")
				if errTidy != nil {
					errMsg := "`go mod tidy` failed"
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errTidy))
					criticalErrors = append(criticalErrors, errors.New("go mod tidy failed"))
					logger.Error("go mod tidy failed", slog.String("source_command", "quality"), slog.String("error", errTidy.Error()))
				} else {
					presenter.Success("go mod tidy completed.")
					logger.Info("go mod tidy successful", slog.String("source_command", "quality"))
				}

			} else {
				// Handle missing 'go' tool
				msg := fmt.Sprintf("'%s' command not found, skipping Go quality checks.", toolGo)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("Go checks skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", toolGo))
			}
		}

		// --- Summary ---
		presenter.Newline()
		presenter.Header("Quality Checks Summary")

		if len(warnings) > 0 {
			presenter.Warning("Issues reported during checks (Includes non-critical linter findings and Go fmt results):") // Clarify scope
			for _, w := range warnings {
				presenter.Warning("  - %s", w)
			}
			presenter.Newline()
		}

		if len(criticalErrors) > 0 {
			errMsg := fmt.Sprintf("%d critical quality check(s) failed.", len(criticalErrors))
			presenter.Error(errMsg)
			presenter.Advice("Please review the errors above and fix them.")
			logger.Error("Quality command failed due to critical errors", slog.String("source_command", "quality"), slog.Int("error_count", len(criticalErrors)))
			return criticalErrors[0] // Return first critical error
		}

		// If there were warnings but no critical errors
		if len(warnings) > 0 {
			presenter.Success("All critical quality checks passed, but warnings were reported (check summary).")
		} else {
			presenter.Success("All quality checks passed successfully.")
		}
		logger.Info("Quality command finished", slog.String("source_command", "quality"), slog.Int("critical_errors", len(criticalErrors)), slog.Int("warnings_count", len(warnings)))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(qualityCmd)
}
