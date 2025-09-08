// cmd/product/codemod/codemod.go
package codemod

import (
	"encoding/json"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/codemod"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed codemod.md.tpl
var codemodLongDescription string

var codemodScriptPath string

// CodemodCmd represents the codemod command
var CodemodCmd = &cobra.Command{
	Use:   "codemod [--script <file.json>]",
	Example: `  contextvibes product codemod # Looks for codemod.json
  contextvibes product codemod --script ./my_refactor_script.json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		
		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		assumeYes, ok := cmd.Context().Value("assumeYes").(bool)
		if !ok { return errors.New("assumeYes not found in context") }

		scriptToLoad := codemodScriptPath
		if scriptToLoad == "" {
			scriptToLoad = "codemod.json"
		}

		scriptData, err := os.ReadFile(scriptToLoad)
		if err != nil {
			return fmt.Errorf("failed to read codemod script '%s': %w", scriptToLoad, err)
		}

		var script codemod.ChangeScript
		if err := json.Unmarshal(scriptData, &script); err != nil {
			return fmt.Errorf("failed to parse codemod script JSON: %w", err)
		}

		for _, fileChangeSet := range script {
			presenter.Header("Processing target: %s", fileChangeSet.FilePath)
			
			contentBytes, err := os.ReadFile(fileChangeSet.FilePath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			currentContent := string(contentBytes)

			for _, op := range fileChangeSet.Operations {
				switch op.Type {
				case "regex_replace":
					re, err := regexp.Compile(op.FindRegex)
					if err != nil { return err }
					currentContent = re.ReplaceAllString(currentContent, op.ReplaceWith)
				// Add other operations here
				}
			}

			if !assumeYes {
				confirmed, err := presenter.PromptForConfirmation(fmt.Sprintf("Write changes to %s?", fileChangeSet.FilePath))
				if err != nil || !confirmed {
					continue
				}
			}
			if err := os.WriteFile(fileChangeSet.FilePath, []byte(currentContent), 0644); err != nil {
				return err
			}
			logger.Info("Applied codemod", "file", fileChangeSet.FilePath)
		}
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(codemodLongDescription, nil)
	if err != nil {
		panic(err)
	}
	CodemodCmd.Short = desc.Short
	CodemodCmd.Long = desc.Long
	CodemodCmd.Flags().StringVarP(&codemodScriptPath, "script", "s", "", "Path to the JSON codemod script file")
}
