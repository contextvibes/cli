// cmd/plan.go

package cmd

import (
	"context" // Import context
	"errors"
	"fmt"
	"log/slog" // For logging errors to AI file
	"os"
	"os/exec"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
)

// Assume AppLogger is initialized in rootCmd for AI file logging
// var AppLogger *slog.Logger // Defined in root.go

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generates an execution plan (e.g., terraform plan, pulumi preview).",
	Long: `Detects the project type (Terraform, Pulumi) and runs the appropriate
command to generate an execution plan, showing expected infrastructure changes.

- Terraform: Runs 'terraform plan -out=tfplan.out'
- Pulumi: Runs 'pulumi preview'`,
	Example: `  contextvibes plan  # Run in a Terraform or Pulumi project directory`,
	Args:    cobra.NoArgs,
	// Add Silence flags as we handle output/errors via Presenter
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			// Note: Internal errors often don't strictly follow ST1005, but lowercasing is harmless.
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background() // Use context

		presenter.Summary("Generating execution plan.")

		cwd, err := os.Getwd()
		if err != nil {
			// Log actual error, return wrapped error
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Plan: Failed getwd", slog.String("error", err.Error())) // Log original error
			presenter.Error("Failed to get current working directory: %v", err)               // Show original error to user
			return wrappedErr                                                                 // Return wrapped error
		}

		presenter.Info("Detecting project type in %s...", presenter.Highlight(cwd))
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(ctx, "Plan: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return wrappedErr
		}

		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "plan"), slog.String("type", string(projType)))

		switch projType {
		case project.Terraform:
			return executeTerraformPlan(ctx, presenter, logger, cwd) // Pass context
		case project.Pulumi:
			return executePulumiPreview(ctx, presenter, logger, cwd) // Pass context
		case project.Go:
			presenter.Info("Plan command is not applicable for Go projects.")
			return nil
		case project.Python:
			presenter.Info("Plan command is not applicable for Python projects.")
			return nil
		case project.Unknown:
			errMsgForUser := "Unknown project type detected. Cannot determine plan action."
			errMsgForError := "unknown project type detected" // ST1005 compliant
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "plan")) // Log user message
			return errors.New(errMsgForError)
		default:
			errMsgForUser := fmt.Sprintf("Internal error: Unhandled project type '%s'", projType)
			errMsgForError := fmt.Sprintf("internal error: unhandled project type '%s'", projType) // ST1005 compliant
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "plan")) // Log user message
			return errors.New(errMsgForError)
		}
	},
}

// Modified to accept context, presenter and logger
func executeTerraformPlan(_ context.Context, presenter *ui.Presenter, logger *slog.Logger, dir string) error {
	tool := "terraform"
	args := []string{"plan", "-out=tfplan.out"}
	if !tools.CommandExists(tool) {
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Terraform is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool) // ST1005 compliant
		presenter.Error(errMsgForUser)
		logger.Error("Terraform plan prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	presenter.Info("Executing: %s %s", tool, strings.Join(args, " "))
	logger.Info("Executing terraform plan", slog.String("source_command", "plan"), slog.String("tool", tool), slog.Any("args", args))

	// Using context with ExecuteCommand (assuming ExecuteCommand is updated or CaptureCommandOutput is used if context matters)
	// Let's assume ExecuteCommand now accepts context or we switch if needed.
	// For now, stick with ExecuteCommand as the code uses it.
	err := tools.ExecuteCommand(dir, tool, args...) // TODO: Pass ctx if ExecuteCommand supports it

	// Handle Terraform plan's specific exit codes
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Exit code 2 means changes are needed - this is SUCCESS for the plan command itself.
			if exitErr.ExitCode() == 2 {
				presenter.Newline()
				presenter.Info("Terraform plan indicates changes are needed.")
				presenter.Advice("Plan saved to tfplan.out. Run `contextvibes deploy` to apply.")
				logger.Info("Terraform plan successful (changes detected)", slog.String("source_command", "plan"), slog.Int("exit_code", 2))
				return nil // Success
			}
			// Any other non-zero exit code is a failure
			// Keep the user-facing message formatted as before
			errMsgForUser := fmt.Sprintf("'%s plan' command failed.", tool)
			// Create a separate error value following conventions
			errMsgForError := fmt.Sprintf("%s plan command failed", tool) // Lowercase, no punctuation - FIXED
			presenter.Error(errMsgForUser)                                // Show user-facing message
			logger.Error("Terraform plan command failed", slog.String("source_command", "plan"), slog.Int("exit_code", exitErr.ExitCode()), slog.String("error", err.Error()))
			return errors.New(errMsgForError) // Return the conventional error value
		}
		// Error wasn't an ExitError (e.g., command not found, though checked above)
		// Use original error wrapped for more context if possible
		errMsgForUser := fmt.Sprintf("Failed to execute '%s plan': %v", tool, err)
		presenter.Error(errMsgForUser)
		logger.Error("Terraform plan execution failed", slog.String("source_command", "plan"), slog.String("error", err.Error()))
		// Return the original error (or a wrapped version)
		return fmt.Errorf("failed to execute '%s plan': %w", tool, err)
	}

	// Exit code 0 means no changes detected
	presenter.Newline()
	presenter.Info("Terraform plan successful (no changes detected).")
	presenter.Advice("Plan saved to tfplan.out (contains no changes).")
	logger.Info("Terraform plan successful (no changes)", slog.String("source_command", "plan"), slog.Int("exit_code", 0))
	return nil
}

// Modified to accept context, presenter and logger
func executePulumiPreview(_ context.Context, presenter *ui.Presenter, logger *slog.Logger, dir string) error {
	tool := "pulumi"
	args := []string{"preview"}
	if !tools.CommandExists(tool) {
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Pulumi is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool) // ST1005 compliant
		presenter.Error(errMsgForUser)
		logger.Error("Pulumi preview prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	presenter.Info("Executing: %s %s", tool, strings.Join(args, " "))
	logger.Info("Executing pulumi preview", slog.String("source_command", "plan"), slog.String("tool", tool), slog.Any("args", args))

	err := tools.ExecuteCommand(dir, tool, args...) // TODO: Pass ctx if ExecuteCommand supports it
	if err != nil {
		// Pulumi preview usually exits 0 on success (even with changes), non-zero on error.
		errMsgForUser := fmt.Sprintf("'%s preview' command failed.", tool)
		errMsgForError := fmt.Sprintf("%s preview command failed", tool) // ST1005 compliant
		presenter.Error(errMsgForUser)
		logger.Error("Pulumi preview command failed", slog.String("source_command", "plan"), slog.String("error", err.Error()))
		return errors.New(errMsgForError) // Return simpler error
	}

	presenter.Newline()
	presenter.Success("Pulumi preview completed successfully.")
	logger.Info("Pulumi preview successful", slog.String("source_command", "plan"))
	return nil
}

func init() {
	rootCmd.AddCommand(planCmd)
}
