// cmd/plan.go
package cmd

import (
	"context" // Import context
	"errors"
	"fmt"
	"log/slog" // For logging errors to AI file
	"os"
	osexec "os/exec" // Alias for standard library exec.ExitError, in case of ExitError check
	"strings"

	"github.com/contextvibes/cli/internal/project"
	// "github.com/contextvibes/cli/internal/tools" // No longer needed for exec functions
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
	// Ensure internal/exec is available if not already imported by other files in cmd
	// but we'll be using the global ExecClient from cmd/root.go
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generates an execution plan (e.g., terraform plan, pulumi preview).",
	Long: `Detects the project type (Terraform, Pulumi) and runs the appropriate
command to generate an execution plan, showing expected infrastructure changes.

- Terraform: Runs 'terraform plan -out=tfplan.out'
- Pulumi: Runs 'pulumi preview'`,
	Example: `  contextvibes plan  # Run in a Terraform or Pulumi project directory`,
	Args:    cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		// ExecClient should also be available from cmd/root.go
		if ExecClient == nil {
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}

		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Generating execution plan.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Plan: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return wrappedErr
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
			return executeTerraformPlan(ctx, presenter, logger, ExecClient, cwd) // Pass ExecClient
		case project.Pulumi:
			return executePulumiPreview(ctx, presenter, logger, ExecClient, cwd) // Pass ExecClient
		case project.Go:
			presenter.Info("Plan command is not applicable for Go projects.")
			return nil
		case project.Python:
			presenter.Info("Plan command is not applicable for Python projects.")
			return nil
		case project.Unknown:
			errMsgForUser := "Unknown project type detected. Cannot determine plan action."
			errMsgForError := "unknown project type detected"
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "plan"))
			return errors.New(errMsgForError)
		default:
			errMsgForUser := fmt.Sprintf("Internal error: Unhandled project type '%s'", projType)
			errMsgForError := fmt.Sprintf("internal error: unhandled project type '%s'", projType)
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "plan"))
			return errors.New(errMsgForError)
		}
	},
}

// Modified to accept execClient
func executeTerraformPlan(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execClientInterface, dir string) error {
	tool := "terraform"
	args := []string{"plan", "-out=tfplan.out"}

	if !execClient.CommandExists(tool) { // Use ExecClient
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Terraform is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Terraform plan prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	presenter.Info("Executing: %s %s", tool, strings.Join(args, " "))
	logger.Info("Executing terraform plan", slog.String("source_command", "plan"), slog.String("tool", tool), slog.Any("args", args))

	// Use ExecClient.Execute. Since terraform plan's output is important, CaptureOutput might be better,
	// but if the OSCommandExecutor pipes stdio for Execute, it might be fine.
	// For consistency with how it might have worked before (piping output), Execute is okay.
	// However, to check exit codes correctly, CaptureOutput is often more robust as Execute's error might be too generic.
	// Let's switch to CaptureOutput to analyze exit codes more reliably.
	_, stderr, err := execClient.CaptureOutput(ctx, dir, tool, args...) // Use CaptureOutput

	if err != nil {
		var exitErr *osexec.ExitError // from os/exec
		if errors.As(err, &exitErr) {
			// Exit code 2 from `terraform plan` means changes are needed (success for plan command)
			if exitErr.ExitCode() == 2 {
				presenter.Newline()
				// stderr might contain the plan output itself or useful info, so display it.
				if strings.TrimSpace(stderr) != "" {
					presenter.Detail("Terraform plan output (stderr):\n%s", stderr)
				}
				presenter.Info("Terraform plan indicates changes are needed.")
				presenter.Advice("Plan saved to tfplan.out. Run `contextvibes deploy` to apply.")
				logger.Info("Terraform plan successful (changes detected)", slog.String("source_command", "plan"), slog.Int("exit_code", 2))
				return nil
			}
			// Any other non-zero exit code is a failure
			errMsgForUser := fmt.Sprintf("'%s plan' command failed.", tool)
			errMsgForError := fmt.Sprintf("%s plan command failed", tool)
			presenter.Error(errMsgForUser)
			if strings.TrimSpace(stderr) != "" {
				presenter.Error("Details (stderr):\n%s", stderr)
			}
			logger.Error("Terraform plan command failed", slog.String("source_command", "plan"), slog.Int("exit_code", exitErr.ExitCode()), slog.String("error", err.Error()), slog.String("stderr", stderr))
			return errors.New(errMsgForError)
		}
		// Error wasn't an ExitError
		errMsgForUser := fmt.Sprintf("Failed to execute '%s plan': %v", tool, err)
		presenter.Error(errMsgForUser)
		if strings.TrimSpace(stderr) != "" {
			presenter.Error("Details (stderr):\n%s", stderr)
		}
		logger.Error("Terraform plan execution failed", slog.String("source_command", "plan"), slog.String("error", err.Error()), slog.String("stderr", stderr))
		return fmt.Errorf("failed to execute '%s plan': %w", tool, err)
	}

	// Exit code 0 means no changes detected
	// stdout from `terraform plan -out=...` is usually minimal, confirmation messages.
	// The actual plan is in the file or on stderr if not using -out.
	presenter.Newline()
	presenter.Info("Terraform plan successful (no changes detected).")
	presenter.Advice("Plan saved to tfplan.out (contains no changes).")
	logger.Info("Terraform plan successful (no changes)", slog.String("source_command", "plan"), slog.Int("exit_code", 0))
	return nil
}

// Modified to accept execClient
func executePulumiPreview(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execClientInterface, dir string) error {
	tool := "pulumi"
	args := []string{"preview"}

	if !execClient.CommandExists(tool) { // Use ExecClient
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Pulumi is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Pulumi preview prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	presenter.Info("Executing: %s %s", tool, strings.Join(args, " "))
	logger.Info("Executing pulumi preview", slog.String("source_command", "plan"), slog.String("tool", tool), slog.Any("args", args))

	// Pulumi preview prints to stdout/stderr itself.
	// Using ExecClient.Execute will pipe these streams.
	err := execClient.Execute(ctx, dir, tool, args...)
	if err != nil {
		// Error message from ExecClient.Execute should be informative enough
		// (includes exit code if that's the cause)
		errMsgForUser := fmt.Sprintf("'%s preview' command failed.", tool)
		errMsgForError := fmt.Sprintf("%s preview command failed", tool)
		presenter.Error(errMsgForUser) // The actual error details would have been piped to stderr by Pulumi
		logger.Error("Pulumi preview command failed", slog.String("source_command", "plan"), slog.String("error", err.Error()))
		return errors.New(errMsgForError)
	}

	presenter.Newline()
	presenter.Success("Pulumi preview completed successfully.")
	logger.Info("Pulumi preview successful", slog.String("source_command", "plan"))
	return nil
}

// Define an interface for execClient to make testing/mocking easier for these functions.
// This interface matches the methods used from exec.ExecutorClient.
type execClientInterface interface {
	CommandExists(commandName string) bool
	Execute(ctx context.Context, dir string, commandName string, args ...string) error
	CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error)
}


func init() {
	rootCmd.AddCommand(planCmd)
}
