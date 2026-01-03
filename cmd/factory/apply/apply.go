// Package apply provides the command to apply changes to the project.
package apply

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/contextvibes/cli/internal/apply"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed apply.md.tpl
var applyLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var scriptPath string

// ApplyCmd represents the apply command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ApplyCmd = &cobra.Command{
	Use:     "apply [--script <file>]",
	Example: `  contextvibes factory apply --script ./plan.json`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		scriptContent, _, err := readInput(scriptPath)
		if err != nil {
			presenter.Error("Failed to read input: %v", err)

			return err
		}

		if len(scriptContent) == 0 {
			presenter.Info("Input is empty. Nothing to apply.")

			return nil
		}

		if isJSON(scriptContent) {
			return handleJSONPlan(ctx, presenter, scriptContent)
		}

		return handleShellScript(ctx, presenter, scriptContent)
	},
}

func readInput(scriptPath string) ([]byte, string, error) {
	if scriptPath != "" {
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, "file", fmt.Errorf("failed to read script file: %w", err)
		}

		return content, "file", nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		//nolint:err113 // Dynamic error is appropriate here.
		return nil, "", errors.New("no script provided via --script flag or standard input")
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, "standard input", fmt.Errorf("failed to read from stdin: %w", err)
	}

	return content, "standard input", nil
}

func isJSON(data []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(data), []byte("{"))
}

//nolint:cyclop // Complexity is acceptable for plan handling.
func handleJSONPlan(ctx context.Context, presenter *ui.Presenter, data []byte) error {
	var plan apply.ChangePlan

	err := json.Unmarshal(data, &plan)
	if err != nil {
		presenter.Error("Failed to parse JSON Change Plan: %v", err)

		return fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	presenter.Header("--- Change Plan Summary ---")

	for i, step := range plan.Steps {
		presenter.Step("Step %d: [%s] %s", i+1, step.Type, step.Description)
	}

	if !globals.AssumeYes {
		confirmed, err := presenter.PromptForConfirmation("Execute the structured plan?")
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			presenter.Info("Execution aborted.")

			return nil
		}
	}

	for _, step := range plan.Steps {
		switch step.Type {
		case "file_modification":
			for _, changeSet := range step.Changes {
				original, _ := os.ReadFile(changeSet.FilePath)
				current := string(original)

				for _, operation := range changeSet.Operations {
					if operation.Type == "create_or_overwrite" {
						current = *operation.Content
					}

					if operation.Type == "regex_replace" {
						re, _ := regexp.Compile(operation.FindRegex)
						current = re.ReplaceAllString(current, operation.ReplaceWith)
					}
				}

				//nolint:mnd // 0750 is standard directory permission.
				_ = os.MkdirAll(filepath.Dir(changeSet.FilePath), 0o750)
				//nolint:mnd // 0600 is standard file permission.
				_ = os.WriteFile(changeSet.FilePath, []byte(current), 0o600)
			}
		case "command_execution":
			err := globals.ExecClient.Execute(ctx, ".", step.Command, step.Args...)
			if err != nil {
				return fmt.Errorf("command execution failed: %w", err)
			}
		}
	}

	presenter.Success("Plan executed successfully.")

	return nil
}

func handleShellScript(ctx context.Context, presenter *ui.Presenter, scriptContent []byte) error {
	presenter.Header("--- Script to be Applied ---")

	fmt.Fprintln(presenter.Out(), "```bash\n"+string(scriptContent)+"\n```")

	if !globals.AssumeYes {
		confirmed, err := presenter.PromptForConfirmation("Execute the shell script?")
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			presenter.Info("Execution aborted.")

			return nil
		}
	}

	tempFile, _ := os.CreateTemp("", "contextvibes-*.sh")

	defer func() { _ = os.Remove(tempFile.Name()) }()

	_, _ = tempFile.Write(scriptContent)
	_ = tempFile.Close()

	err := globals.ExecClient.Execute(ctx, ".", "bash", tempFile.Name())
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(applyLongDescription, nil)
	if err != nil {
		panic(err)
	}

	ApplyCmd.Short = desc.Short
	ApplyCmd.Long = desc.Long
	ApplyCmd.Flags().
		StringVarP(&scriptPath, "script", "s", "", "Path to the Change Plan (JSON) or shell script to apply.")
}
