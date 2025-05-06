// cmd/test.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/tools" // For CommandExists and ExecuteCommand
	"github.com/contextvibes/cli/internal/ui"    // Use Presenter
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test [args...]",
	Short: "Runs project-specific tests (e.g., go test, pytest).",
	Long: `Detects the project type (Go, Python) and runs the appropriate test command.
Any arguments passed to 'contextvibes test' will be forwarded to the underlying test runner.

- Go: Runs 'go test ./...'
- Python: Runs 'pytest' (if available). Falls back to 'python -m unittest discover' if pytest not found.

For other project types, or if no specific test runner is found, it will indicate no action.`,
	Example: `  contextvibes test
  contextvibes test -v  # Passes '-v' to 'go test' or 'pytest'
  contextvibes test tests/my_specific_test.py # Passes path to pytest`,
	// Args: cobra.ArbitraryArgs, // Allow arbitrary arguments to pass to the underlying test command
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Stdin)
		ctx := context.Background()

		presenter.Summary("Running project tests.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Test: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return wrappedErr
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(ctx, "Test: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return wrappedErr
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "test"), slog.String("type", string(projType)))

		var testErr error
		testExecuted := false

		switch projType {
		case project.Go:
			presenter.Header("Go Project Tests")
			testErr = executeGoTests(ctx, presenter, logger, cwd, args)
			testExecuted = true
		case project.Python:
			presenter.Header("Python Project Tests")
			testErr = executePythonTests(ctx, presenter, logger, cwd, args)
			testExecuted = true
		case project.Terraform, project.Pulumi:
			presenter.Info("Automated testing for %s projects is not yet implemented in this command.", projType)
			presenter.Advice("Consider using tools like Terratest or language-specific test frameworks manually.")
			// TODO: Add support for Terraform/Pulumi testing frameworks if desired
		case project.Unknown:
			presenter.Warning("Unknown project type. Cannot determine how to run tests.")
		default:
			presenter.Warning("No specific test execution logic for project type: %s", projType)
		}

		presenter.Newline()
		if !testExecuted && testErr == nil {
			presenter.Info("No tests were executed for the detected project type or configuration.")
			return nil
		}

		if testErr != nil {
			// The underlying tool (go test, pytest) should have printed its own errors.
			// The presenter.Error here just summarizes that the test suite failed.
			presenter.Error("Project tests failed.")
			logger.Error("Test command finished with errors", slog.String("source_command", "test"), slog.String("error", testErr.Error()))
			return testErr // Return the specific error from the test execution
		}

		presenter.Success("Project tests completed successfully.")
		logger.Info("Test command successful", slog.String("source_command", "test"))
		return nil
	},
}

func executeGoTests(_ context.Context, presenter *ui.Presenter, logger *slog.Logger, dir string, passThroughArgs []string) error {
	tool := "go"
	if !tools.CommandExists(tool) {
		errMsgForUser := fmt.Sprintf("'%s' command not found. Ensure Go is installed and in your PATH.", tool)
		errMsgForError := "go command not found"
		presenter.Error(errMsgForUser)
		logger.Error("Go test prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	// Base arguments for go test
	testArgs := []string{"test", "./..."}
	// Append any passthrough arguments
	testArgs = append(testArgs, passThroughArgs...)

	presenter.Info("Executing: %s %s", tool, strings.Join(testArgs, " "))
	logger.Info("Executing go test", slog.String("source_command", "test"), slog.String("tool", tool), slog.Any("args", testArgs))

	// tools.ExecuteCommand pipes stdio directly
	err := tools.ExecuteCommand(dir, tool, testArgs...)
	if err != nil {
		// No need for presenter.Error here, as go test output is already on stderr.
		// tools.ExecuteCommand already wraps the error with exit code info.
		return fmt.Errorf("go test failed: %w", err)
	}
	return nil
}

func executePythonTests(_ context.Context, presenter *ui.Presenter, logger *slog.Logger, dir string, passThroughArgs []string) error {
	pytestTool := "pytest"
	pythonTool := "python" // Or python3

	if tools.CommandExists(pytestTool) {
		presenter.Info("Executing: %s %s", pytestTool, strings.Join(passThroughArgs, " "))
		logger.Info("Executing pytest", slog.String("source_command", "test"), slog.String("tool", pytestTool), slog.Any("args", passThroughArgs))
		err := tools.ExecuteCommand(dir, pytestTool, passThroughArgs...)
		if err != nil {
			return fmt.Errorf("pytest failed: %w", err)
		}
		return nil
	}

	presenter.Info("`pytest` not found. Attempting `python -m unittest discover`...")
	if tools.CommandExists(pythonTool) {
		// Base arguments for unittest
		unittestArgs := []string{"-m", "unittest", "discover"}
		// Append any passthrough arguments (though unittest discover has fewer common ones)
		// For simplicity, we might not pass all args here or make it conditional.
		// unittestArgs = append(unittestArgs, passThroughArgs...) // Be cautious with this

		presenter.Info("Executing: %s %s", pythonTool, strings.Join(unittestArgs, " "))
		logger.Info("Executing python unittest", slog.String("source_command", "test"), slog.String("tool", pythonTool), slog.Any("args", unittestArgs))
		err := tools.ExecuteCommand(dir, pythonTool, unittestArgs...)
		if err != nil {
			return fmt.Errorf("python -m unittest discover failed: %w", err)
		}
		return nil
	}

	errMsgForUser := "Neither `pytest` nor `python` found. Cannot run Python tests."
	errMsgForError := "no python test runner found"
	presenter.Error(errMsgForUser)
	logger.Error("Python test prerequisite failed", slog.String("reason", errMsgForUser))
	return errors.New(errMsgForError)
}

func init() {
	// Allow arbitrary arguments for the test command by setting this
	// after the command has been defined and before adding to root.
	// However, Cobra's handling of arbitrary args can be tricky with subcommands
	// if not careful. For simple passthrough, it's often easier to just
	// iterate over cmd.Flags().Args() directly if cobra.ArbitraryArgs is not behaving as expected.
	// For now, we'll rely on `args` passed to RunE.
	// testCmd.Args = cobra.ArbitraryArgs
	rootCmd.AddCommand(testCmd)
}