// cmd/system_prompt.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	systemPromptTarget string
	systemPromptOutput string
)

var systemPromptCmd = &cobra.Command{
	Use:   "system-prompt",
	Short: "Generates a system prompt for a target environment and saves it to a file.",
	Long: `Builds a complete system prompt by combining the universal 'core.md' rules
with a target-specific instruction file (e.g., 'aistudio.md', 'idx.md')
from the 'docs/prompts/system/' directory.

By default, it saves the output to a file path determined by the target
(e.g., '.idx/airules.md' for --target=idx). Use '-o -' to print to the console.`,
	Example: `  # Generate the default prompt for AI Studio and save it to the default file
  contextvibes system-prompt --target aistudio

  # Generate the prompt for IDX, overwriting the required .idx/airules.md file
  contextvibes system-prompt --target idx

  # Generate the AI Studio prompt but print it to the console to be piped
  contextvibes system-prompt --target aistudio -o - | pbcopy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		basePath := "docs/prompts/system"

		// --- Assemble the prompt content ---
		coreContent, err := os.ReadFile(filepath.Join(basePath, "core.md"))
		if err != nil {
			presenter.Error("Failed to read core system prompt 'docs/prompts/system/core.md': %v", err)
			return err
		}

		var finalPrompt strings.Builder
		finalPrompt.Write(coreContent)

		if systemPromptTarget != "" {
			targetFilename := systemPromptTarget + ".md"
			targetPath := filepath.Join(basePath, targetFilename)

			targetContent, err := os.ReadFile(targetPath)
			if err != nil {
				presenter.Error("Failed to read target prompt file '%s': %v", targetPath, err)
				presenter.Advice("Ensure a file named '%s' exists in the '%s' directory.", targetFilename, basePath)
				return err
			}
			finalPrompt.WriteString("\n\n")
			finalPrompt.Write(targetContent)
		}

		// --- Determine output destination ---
		outputPath := systemPromptOutput
		if outputPath == "" { // If -o flag is not used, determine default path from config
			if defaultPath, ok := LoadedAppConfig.SystemPrompt.DefaultOutputFiles[systemPromptTarget]; ok {
				outputPath = defaultPath
			} else {
				// Construct a fallback path if the target is unknown to the config
				outputPath = fmt.Sprintf("contextvibes_%s_prompt.md", systemPromptTarget)
			}
		}

		// --- Write the output ---
		if outputPath == "-" {
			// Write to stdout
			fmt.Fprint(presenter.Out(), finalPrompt.String())
		} else {
			// Write to file
			presenter.Summary("Generating system prompt for target '%s'...", systemPromptTarget)
			err := os.WriteFile(outputPath, []byte(finalPrompt.String()), 0644)
			if err != nil {
				presenter.Error("Failed to write to output file '%s': %v", outputPath, err)
				return err
			}
			presenter.Success("Successfully generated system prompt at %s.", outputPath)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(systemPromptCmd)
	systemPromptCmd.Flags().StringVar(&systemPromptTarget, "target", "aistudio", "The target environment for the system prompt (e.g., aistudio, idx)")
	systemPromptCmd.Flags().StringVarP(&systemPromptOutput, "output", "o", "", "Output file path. Use '-' to print to standard output.")
}
