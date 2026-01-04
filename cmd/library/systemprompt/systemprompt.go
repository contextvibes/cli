// Package systemprompt provides the command to generate system prompts.
package systemprompt

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed systemprompt.md.tpl
var systemPromptLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	systemPromptTarget string
	systemPromptOutput string
)

// SystemPromptCmd represents the system-prompt command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var SystemPromptCmd = &cobra.Command{
	Use:     "system-prompt",
	Aliases: []string{"prompt"},
	Example: `  contextvibes library system-prompt --target idx`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		loadedAppConfig := globals.LoadedAppConfig

		basePath := "docs/prompts/system"

		content, err := os.ReadFile(filepath.Join(basePath, "core.md"))
		if err != nil {
			//nolint:wrapcheck // Wrapping is handled by caller.
			return err
		}

		var finalPrompt strings.Builder
		finalPrompt.Write(content)

		if systemPromptTarget != "" {

			targetContent, err := os.ReadFile(filepath.Join(basePath, systemPromptTarget+".md"))
			if err != nil {
				//nolint:wrapcheck // Wrapping is handled by caller.
				return err
			}
			finalPrompt.WriteString("\n\n")
			finalPrompt.Write(targetContent)
		}

		outputPath := systemPromptOutput
		if outputPath == "" {
			if defaultPath, ok := loadedAppConfig.SystemPrompt.DefaultOutputFiles[systemPromptTarget]; ok {
				outputPath = defaultPath
			} else {
				outputPath = fmt.Sprintf("contextvibes_%s_prompt.md", systemPromptTarget)
			}
		}

		if outputPath == "-" {

			_, _ = fmt.Fprint(presenter.Out(), finalPrompt.String())
		} else {
			//nolint:mnd // 0600 is standard file permission.
			err := os.WriteFile(outputPath, []byte(finalPrompt.String()), 0o600)
			if err != nil {
				//nolint:wrapcheck // Wrapping is handled by caller.
				return err
			}
			presenter.Success("Successfully generated system prompt at %s.", outputPath)
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(systemPromptLongDescription, nil)
	if err != nil {
		panic(err)
	}

	SystemPromptCmd.Short = desc.Short
	SystemPromptCmd.Long = desc.Long
	//nolint:lll // Flag description is long.
	SystemPromptCmd.Flags().
		StringVar(&systemPromptTarget, "target", "aistudio", "The target environment for the system prompt (e.g., aistudio, idx)")
	SystemPromptCmd.Flags().
		StringVarP(&systemPromptOutput, "output", "o", "", "Output file path. Use '-' for stdout.")
}
