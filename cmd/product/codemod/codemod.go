// Package codemod provides the command to apply code modifications.
package codemod

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/codemod"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed codemod.md.tpl
var codemodLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var codemodScriptPath string

// CodemodCmd represents the codemod command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CodemodCmd = &cobra.Command{
	Use: "codemod [--script <file.json>]",
	Example: `  contextvibes product codemod # Looks for codemod.json
  contextvibes product codemod --script ./my_refactor_script.json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		scriptToLoad := codemodScriptPath
		if scriptToLoad == "" {
			scriptToLoad = "codemod.json"
		}

		scriptData, err := os.ReadFile(scriptToLoad)
		if err != nil {
			return fmt.Errorf("failed to read codemod script '%s': %w", scriptToLoad, err)
		}

		var script codemod.ChangeScript
		err = json.Unmarshal(scriptData, &script)
		if err != nil {
			return fmt.Errorf("failed to parse codemod script JSON: %w", err)
		}

		for _, fileChangeSet := range script {
			presenter.Header("Processing target: %s", fileChangeSet.FilePath)

			contentBytes, err := os.ReadFile(fileChangeSet.FilePath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to read target file: %w", err)
			}
			currentContent := string(contentBytes)

			//nolint:varnamelen // 'op' is standard for operation.
			for _, op := range fileChangeSet.Operations {
				switch op.Type {
				case "regex_replace":
					re, err := regexp.Compile(op.FindRegex)
					if err != nil {
						return fmt.Errorf("invalid regex '%s': %w", op.FindRegex, err)
					}
					currentContent = re.ReplaceAllString(currentContent, op.ReplaceWith)
					// Add other operations here
				}
			}

			if !globals.AssumeYes {
				confirmed, err := presenter.PromptForConfirmation(
					fmt.Sprintf("Write changes to %s?", fileChangeSet.FilePath),
				)
				if err != nil || !confirmed {
					continue
				}
			}
			//nolint:mnd // 0600 is standard file permission.
			err = os.WriteFile(fileChangeSet.FilePath, []byte(currentContent), 0o600)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			globals.AppLogger.Info("Applied codemod", "file", fileChangeSet.FilePath)
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(codemodLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CodemodCmd.Short = desc.Short
	CodemodCmd.Long = desc.Long
	CodemodCmd.Flags().
		StringVarP(&codemodScriptPath, "script", "s", "", "Path to the JSON codemod script file")
}
