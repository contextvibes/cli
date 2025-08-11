// cmd/apply.go
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/apply"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	scriptPath       string
	promptOutputPath string
)

// applyCmd is the parent command group for apply-related actions.
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Applies a structured Change Plan or executes a shell script.",
	Long: `A group of commands for applying changes to a project.

The 'apply' command is the primary executor for AI-generated solutions. It can operate in two modes:
1. Structured Plan (JSON): This is the preferred and safer mode of operation.
2. Fallback Script (Shell): For simple, imperative scripts.

Run 'contextvibes apply execute' to apply a plan or script.`,
}

var applyExecuteCmd = &cobra.Command{
	Use:     "execute [--script <file>]",
	Aliases: []string{"", "run"}, // Empty alias makes 'apply' default to this subcommand
	Short:   "Executes a Change Plan or shell script after confirmation.",
	Long: `Executes a structured Change Plan (JSON) or a raw shell script.

Input can be read from a file with --script or piped from standard input. The command
will detect the input type, show a summary, and require confirmation before running.`,
	Example: `  # Apply a structured plan from a file
  contextvibes apply --script ./plan.json

  # Pipe a plan from an AI or another tool (defaults to the execute subcommand)
  cat ./plan.json | contextvibes apply`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context()

		scriptContent, source, err := readInput(scriptPath)
		if err != nil {
			presenter.Error(err.Error())
			presenter.Advice("Usage: contextvibes apply execute --script <file> OR cat <file> | contextvibes apply")
			return err
		}

		if len(scriptContent) == 0 {
			presenter.Warning("The provided input from %s is empty. Nothing to do.", source)
			return nil
		}

		if isJSON(scriptContent) {
			return handleJSONPlan(ctx, presenter, scriptContent, source)
		} else {
			return handleShellScript(ctx, presenter, scriptContent, source)
		}
	},
}

var applyPromptCmd = &cobra.Command{
	Use:   "prompt [-o <output_file>]",
	Short: "Generates a detailed prompt for an AI on how to create a Change Plan.",
	Long: `Outputs a comprehensive prompt that can be given to an AI assistant.
This prompt instructs the AI on the exact JSON schema and best practices for generating
a structured Change Plan that the 'contextvibes apply' command can execute.`,
	Example: `  # Print the prompt to the console
  contextvibes apply prompt

  # Save the prompt to a file to use in an AI chat session
  contextvibes apply prompt -o ./ai_instruction.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		promptContent := apply.GetChangePlanPrompt()
		if promptOutputPath != "" {
			if err := os.WriteFile(promptOutputPath, []byte(promptContent), 0644); err != nil {
				return fmt.Errorf("failed to write prompt to file %s: %w", promptOutputPath, err)
			}
			fmt.Printf("Prompt successfully written to %s\n", promptOutputPath)
		} else {
			fmt.Println(promptContent)
		}
		return nil
	},
}

// --- Business Logic (Shared by Subcommands) ---

func readInput(scriptPath string) ([]byte, string, error) {
	var scriptContent []byte
	var err error
	source := "standard input"

	if scriptPath != "" {
		source = fmt.Sprintf("file '%s'", scriptPath)
		scriptContent, err = os.ReadFile(scriptPath)
		if err != nil {
			return nil, source, fmt.Errorf("failed to read script from %s: %w", source, err)
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, source, errors.New("no script provided via --script flag or standard input")
		}
		scriptContent, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, source, fmt.Errorf("failed to read script from standard input: %w", err)
		}
	}
	return scriptContent, source, nil
}

func isJSON(data []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(data), []byte("{"))
}

func handleJSONPlan(ctx context.Context, presenter *ui.Presenter, data []byte, source string) error {
	var plan apply.ChangePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		presenter.Error("Failed to parse JSON Change Plan from %s: %v", source, err)
		return err
	}

	presenter.Header("--- Change Plan Summary ---")
	presenter.Info("Source: %s", source)
	presenter.Info("Description: %s", plan.Description)
	presenter.Newline()

	for i, step := range plan.Steps {
		presenter.Step("Step %d: [%s] %s", i+1, step.Type, step.Description)
	}
	presenter.Newline()

	confirmed, err := confirmExecution(presenter, "Execute the structured plan shown above?")
	if err != nil || !confirmed {
		return err
	}

	presenter.Header("--- Executing Plan ---")
	for i, step := range plan.Steps {
		presenter.Step("Executing Step %d: %s", i+1, step.Description)
		switch step.Type {
		case "file_modification":
			err = executeFileModificationStep(ctx, presenter, step)
		case "command_execution":
			err = executeCommandExecutionStep(ctx, presenter, step)
		default:
			err = fmt.Errorf("unknown step type: '%s'", step.Type)
		}

		if err != nil {
			presenter.Error("Failed to execute step %d: %v", i+1, err)
			return err
		}
	}

	presenter.Success("Plan executed successfully.")
	return nil
}

func executeFileModificationStep(ctx context.Context, presenter *ui.Presenter, step apply.Step) error {
	for _, changeSet := range step.Changes {
		presenter.Detail("  Applying changes to: %s", changeSet.FilePath)
		originalContent, readErr := os.ReadFile(changeSet.FilePath)
		if readErr != nil && !os.IsNotExist(readErr) {
			return fmt.Errorf("could not read file %s: %w", changeSet.FilePath, readErr)
		}

		currentContent := string(originalContent)

		for _, op := range changeSet.Operations {
			switch op.Type {
			case "create_or_overwrite":
				if op.Content == nil {
					return fmt.Errorf("'create_or_overwrite' operation for %s is missing 'content' field", changeSet.FilePath)
				}
				currentContent = *op.Content
			case "regex_replace":
				re, err := regexp.Compile(op.FindRegex)
				if err != nil {
					return fmt.Errorf("invalid regex '%s' for file %s: %w", op.FindRegex, changeSet.FilePath, err)
				}
				currentContent = re.ReplaceAllString(currentContent, op.ReplaceWith)
			default:
				return fmt.Errorf("unsupported file operation type: '%s'", op.Type)
			}
		}

		if err := os.MkdirAll(filepath.Dir(changeSet.FilePath), 0750); err != nil {
			return fmt.Errorf("could not create parent directories for %s: %w", changeSet.FilePath, err)
		}
		if err := os.WriteFile(changeSet.FilePath, []byte(currentContent), 0644); err != nil {
			return fmt.Errorf("could not write file %s: %w", changeSet.FilePath, err)
		}
	}
	return nil
}

func executeCommandExecutionStep(ctx context.Context, presenter *ui.Presenter, step apply.Step) error {
	presenter.Detail("  Running command: %s %s", step.Command, strings.Join(step.Args, " "))
	return ExecClient.Execute(ctx, ".", step.Command, step.Args...)
}

func handleShellScript(ctx context.Context, presenter *ui.Presenter, scriptContent []byte, source string) error {
	presenter.Header("--- Script to be Applied (Fallback Mode) ---")
	presenter.Detail("Source: %s\n", source)
	_, _ = fmt.Fprintln(presenter.Out(), "```bash")
	_, _ = presenter.Out().Write(scriptContent)
	_, _ = fmt.Fprintln(presenter.Out(), "\n```")
	presenter.Newline()

	confirmed, err := confirmExecution(presenter, "Execute the shell script shown above?")
	if err != nil || !confirmed {
		return err
	}

	presenter.Header("--- Script Execution ---")
	tempFile, err := os.CreateTemp("", "contextvibes-apply-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temporary script file: %w", err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			AppLogger.Warn("Failed to clean up temporary script file", "path", tempFile.Name(), "error", err)
		}
	}()

	if _, err := tempFile.Write(scriptContent); err != nil {
		return fmt.Errorf("failed to write to temporary script file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary script file: %w", err)
	}

	if err := ExecClient.Execute(ctx, ".", "bash", tempFile.Name()); err != nil {
		presenter.Error("Script execution failed. See output above for details.")
		return errors.New("script execution returned a non-zero exit code")
	}

	presenter.Success("Script executed successfully.")
	return nil
}

func confirmExecution(presenter *ui.Presenter, prompt string) (bool, error) {
	if assumeYes {
		presenter.Info("Confirmation bypassed via --yes flag.")
		return true, nil
	}
	confirmed, err := presenter.PromptForConfirmation(prompt)
	if err != nil {
		return false, err
	}
	if !confirmed {
		presenter.Info("Execution aborted by user.")
		return false, nil
	}
	return true, nil
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Flags for execute subcommand
	applyExecuteCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Path to the Change Plan (JSON) or shell script to apply.")
	applyCmd.AddCommand(applyExecuteCmd)

	// Flags for prompt subcommand
	applyPromptCmd.Flags().StringVarP(&promptOutputPath, "output", "o", "", "Path to save the generated AI prompt.")
	applyCmd.AddCommand(applyPromptCmd)
}
