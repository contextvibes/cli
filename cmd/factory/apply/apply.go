// cmd/factory/apply/apply.go
package apply

import (
	"bytes"
	"context"
	"encoding/json"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/contextvibes/cli/internal/apply"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed apply.md.tpl
var applyLongDescription string

var scriptPath string

// ApplyCmd represents the apply command
var ApplyCmd = &cobra.Command{
	Use:     "apply [--script <file>]",
	Example: `  contextvibes factory apply --script ./plan.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		
		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		execClient, ok := cmd.Context().Value("execClient").(*exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		assumeYes, ok := cmd.Context().Value("assumeYes").(bool)
		if !ok { return errors.New("assumeYes not found in context") }
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
			return handleJSONPlan(ctx, presenter, execClient, logger, assumeYes, scriptContent)
		}
		return handleShellScript(ctx, presenter, execClient, logger, assumeYes, scriptContent)
	},
}

func readInput(scriptPath string) ([]byte, string, error) {
	if scriptPath != "" {
		content, err := os.ReadFile(scriptPath)
		return content, "file", err
	}
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, "", errors.New("no script provided via --script flag or standard input")
	}
	content, err := io.ReadAll(os.Stdin)
	return content, "standard input", err
}

func isJSON(data []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(data), []byte("{"))
}

func handleJSONPlan(ctx context.Context, presenter *ui.Presenter, execClient *exec.ExecutorClient, logger *slog.Logger, assumeYes bool, data []byte) error {
	var plan apply.ChangePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		presenter.Error("Failed to parse JSON Change Plan: %v", err)
		return err
	}
	
	presenter.Header("--- Change Plan Summary ---")
	for i, step := range plan.Steps {
		presenter.Step("Step %d: [%s] %s", i+1, step.Type, step.Description)
	}

	if !assumeYes {
		confirmed, err := presenter.PromptForConfirmation("Execute the structured plan?")
		if err != nil || !confirmed {
			presenter.Info("Execution aborted.")
			return err
		}
	}

	for _, step := range plan.Steps {
		switch step.Type {
		case "file_modification":
			for _, changeSet := range step.Changes {
				original, _ := os.ReadFile(changeSet.FilePath)
				current := string(original)
				for _, op := range changeSet.Operations {
					if op.Type == "create_or_overwrite" { current = *op.Content }
					if op.Type == "regex_replace" { re, _ := regexp.Compile(op.FindRegex); current = re.ReplaceAllString(current, op.ReplaceWith) }
				}
				os.MkdirAll(filepath.Dir(changeSet.FilePath), 0o750)
				os.WriteFile(changeSet.FilePath, []byte(current), 0o600)
			}
		case "command_execution":
			if err := execClient.Execute(ctx, ".", step.Command, step.Args...); err != nil { return err }
		}
	}
	presenter.Success("Plan executed successfully.")
	return nil
}

func handleShellScript(ctx context.Context, presenter *ui.Presenter, execClient *exec.ExecutorClient, logger *slog.Logger, assumeYes bool, scriptContent []byte) error {
	presenter.Header("--- Script to be Applied ---")
	fmt.Fprintln(presenter.Out(), "```bash\n"+string(scriptContent)+"\n```")
	
	if !assumeYes {
		confirmed, err := presenter.PromptForConfirmation("Execute the shell script?")
		if err != nil || !confirmed {
			presenter.Info("Execution aborted.")
			return err
		}
	}

	tempFile, _ := os.CreateTemp("", "contextvibes-*.sh")
	defer os.Remove(tempFile.Name())
	tempFile.Write(scriptContent)
	tempFile.Close()
	return execClient.Execute(ctx, ".", "bash", tempFile.Name())
}

func init() {
	desc, err := cmddocs.ParseAndExecute(applyLongDescription, nil)
	if err != nil {
		panic(err)
	}
	ApplyCmd.Short = desc.Short
	ApplyCmd.Long = desc.Long
	ApplyCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "Path to the Change Plan (JSON) or shell script to apply.")
}
