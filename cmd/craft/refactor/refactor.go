// Package refactor provides the command to generate refactoring prompts.
package refactor

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed refactor.md.tpl
var refactorLongDescription string

//go:embed prompt.md.tpl
var refactorPromptTemplate string

// RefactorCmd represents the craft refactor command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var RefactorCmd = &cobra.Command{
	Use:   "refactor [files...]",
	Short: "Generates a prompt for AI-assisted code refactoring.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		// 1. Render the Prompt Template
		tmpl, err := template.New("refactor-prompt").Parse(refactorPromptTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse prompt template: %w", err)
		}

		var promptBuf bytes.Buffer
		// Future: Pass a struct with config/context here instead of nil
		if err := tmpl.Execute(&promptBuf, nil); err != nil {
			return fmt.Errorf("failed to render prompt template: %w", err)
		}

		// 2. Read and Format File Content
		var codeContent strings.Builder
		for _, filePath := range args {

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file '%s': %w", filePath, err)
			}

			codeContent.WriteString(fmt.Sprintf("\n--- FILE: %s ---\n", filePath))
			codeContent.WriteString("```go\n")
			codeContent.Write(content)
			codeContent.WriteString("\n```\n")
		}

		// 3. Output
		presenter.Header("--- Copy the text below to your AI ---")
		//nolint:forbidigo // Printing prompt to stdout is the core feature.
		fmt.Println(promptBuf.String())
		//nolint:forbidigo // Printing prompt to stdout is the core feature.
		fmt.Println(codeContent.String())
		presenter.Header("--- End of Prompt ---")

		presenter.Success("Refactoring prompt generated for %d file(s).", len(args))

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(refactorLongDescription, nil)
	if err != nil {
		panic(err)
	}

	RefactorCmd.Short = desc.Short
	RefactorCmd.Long = desc.Long
}
