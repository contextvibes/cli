// Package plan provides commands to plan project work.
package plan

import (
	"github.com/contextvibes/cli/cmd/project/plan/refine"
	suggest "github.com/contextvibes/cli/cmd/project/plan/suggest-refinement"
	"github.com/spf13/cobra"
)

// PlanCmd represents the base command for the 'plan' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var PlanCmd = &cobra.Command{
	Use:     "plan",
	Short:   "Commands for backlog grooming and sprint planning.",
	Aliases: []string{"scrum"},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	PlanCmd.AddCommand(refine.RefineCmd)
	PlanCmd.AddCommand(suggest.SuggestRefinementCmd)
	// Future subcommands like 'link' and 'board' will be added here.
}
