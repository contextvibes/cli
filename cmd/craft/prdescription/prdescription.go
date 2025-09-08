// cmd/craft/prdescription/prdescription.go
package prdescription

import (
	_ "embed"

	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed prdescription.md.tpl
var prDescriptionLongDescription string

// PRDescriptionCmd represents the craft pr-description command
var PRDescriptionCmd = &cobra.Command{
	Use:     "pr-description",
	Aliases: []string{"pr"},
	Short:   "Generates a suggested pull request description.",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		// This command would use the git client to get a diff against main,
		// then pass that to an LLM. For now, we simulate.

		presenter.Summary("Crafting a pull request description...")
		presenter.Info("AI analysis complete. Suggested description:")
		presenter.Newline()

		simulatedPRBody := `### Summary

This change introduces the new 'craft' pillar to the CLI, providing a dedicated space for AI-assisted creative tasks.

### Changes
- Added 'craft message' to generate commit messages.
- Added 'craft pr-description' as a placeholder for generating PR bodies.
- Refactored the strategic kickoff into 'craft kickoff'.`

		fmt.Fprintln(presenter.Out(), simulatedPRBody)
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(prDescriptionLongDescription, nil)
	if err != nil {
		panic(err)
	}
	PRDescriptionCmd.Short = desc.Short
	PRDescriptionCmd.Long = desc.Long
}
