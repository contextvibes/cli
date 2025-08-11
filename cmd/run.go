// cmd/run.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Interactively runs one of the project's example applications after verification.",
	Long: `Discovers runnable example applications within the './examples' directory,
runs any configured prerequisite checks from '.contextvibes.yaml',
and then presents an interactive menu to choose an example to execute with 'go run'.`,
	Example: `  contextvibes run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context()

		presenter.Header("--- Example Runner ---")

		examples, err := findRunnableExamples(".")
		if err != nil {
			presenter.Error("Failed to find examples: %v", err)

			return err
		}

		if len(examples) == 0 {
			presenter.Warning("No runnable examples found in the './examples' directory.")

			return nil
		}

		presenter.Info("Discovered %d example(s).", len(examples))
		presenter.Newline()

		choice, err := presenter.PromptForSelect("Please select an example application to run:", examples)
		if err != nil {
			if choice == "" {
				presenter.Info("No selection made. Exiting.")

				return nil
			}

			return fmt.Errorf("interactive menu failed: %w", err)
		}

		// --- Verification Step ---
		err = runVerificationChecks(ctx, presenter, choice)
		if err != nil {
			// runVerificationChecks already prints detailed errors
			return errors.New("prerequisite verification failed")
		}
		// --- End Verification Step ---

		presenter.Newline()
		presenter.Step("Executing example: %s...", presenter.Highlight(choice))
		presenter.Newline()

		err = ExecClient.Execute(ctx, ".", "go", "run", "./"+choice)
		if err != nil {
			presenter.Error("Failed to run example '%s'. See output above for details.", choice)

			return errors.New("example execution failed")
		}

		AppLogger.InfoContext(ctx, "Successfully launched example", "example_path", choice)

		return nil
	},
}

// runVerificationChecks looks for and executes checks for a given example.
func runVerificationChecks(ctx context.Context, presenter *ui.Presenter, examplePath string) error {
	if LoadedAppConfig.Run.Examples == nil {
		presenter.Info("No 'run.examples' configuration found. Proceeding without verification.")

		return nil
	}

	exampleSettings, ok := LoadedAppConfig.Run.Examples[examplePath]
	if !ok || len(exampleSettings.Verify) == 0 {
		presenter.Info("No verification checks configured for '%s'. Proceeding.", examplePath)

		return nil
	}

	presenter.Header("--- 🔍 Verifying Prerequisites for '%s' ---", examplePath)

	allPassed := true

	for i, check := range exampleSettings.Verify {
		checkTitle := check.Name
		if check.Description != "" {
			checkTitle = check.Description
		}

		presenter.Step("Running check %d/%d: %s...", i+1, len(exampleSettings.Verify), checkTitle)

		// Use CaptureOutput to prevent check's stdout from cluttering the main output,
		// but show it if there's an error.
		_, stderr, err := ExecClient.CaptureOutput(ctx, ".", check.Command, check.Args...)
		if err != nil {
			allPassed = false

			presenter.Error("  ❌ FAILED: Command '%s' failed.", check.Command)

			if stderr != "" {
				presenter.Detail("    Stderr: %s", stderr)
			}

			presenter.Detail("    Error: %v", err)
		} else {
			presenter.Success("  ✅ PASSED")
		}
	}

	presenter.Newline()

	if !allPassed {
		presenter.Error("One or more prerequisite checks failed. Please resolve the issues above.")

		return errors.New("verification failed")
	}

	presenter.Success("All prerequisite checks passed!")

	return nil
}

// findRunnableExamples scans the given root directory for subdirectories within 'examples/'.
func findRunnableExamples(rootDir string) ([]string, error) {
	var examples []string

	examplesDir := filepath.Join(rootDir, "examples")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("could not read examples directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Use ToSlash for consistent path separators in config keys
			examples = append(examples, filepath.ToSlash(filepath.Join("examples", entry.Name())))
		}
	}

	return examples, nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
