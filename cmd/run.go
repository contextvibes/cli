// cmd/run.go
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Interactively runs one of the project's example applications.",
	Long: `Discovers runnable example applications within the './examples' directory
and presents an interactive menu to choose one to execute with 'go run'.`,
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
			// Check for empty choice which indicates user aborted (e.g., Ctrl+C)
			if choice == "" {
				presenter.Info("No selection made. Exiting.")

				return nil
			}

			return fmt.Errorf("interactive menu failed: %w", err)
		}

		presenter.Newline()
		presenter.Step("Executing example: %s...", presenter.Highlight(choice))
		presenter.Newline()

		// The ExecClient will pipe the stdout/stderr of the example directly to the user's terminal.
		err = ExecClient.Execute(ctx, ".", "go", "run", "./"+choice)
		if err != nil {
			// The ExecClient already logs the command and its failure.
			// We just need to provide a user-friendly message.
			presenter.Error("Failed to run example '%s'. See output above for details.", choice)

			return errors.New("example execution failed")
		}

		AppLogger.InfoContext(ctx, "Successfully launched example", "example_path", choice)

		return nil
	},
}

// findRunnableExamples scans the given root directory for subdirectories within 'examples/'.
func findRunnableExamples(rootDir string) ([]string, error) {
	var examples []string

	examplesDir := filepath.Join(rootDir, "examples")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		if os.IsNotExist(err) {
			// It's not an error if the examples directory doesn't exist.
			return nil, nil
		}

		return nil, fmt.Errorf("could not read examples directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			examples = append(examples, filepath.Join("examples", entry.Name()))
		}
	}

	return examples, nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
