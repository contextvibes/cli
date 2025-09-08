// cmd/project/plan/plan.go
package plan

import (
	"github.com/contextvibes/cli/cmd/project/plan/refine"
	"github.com/contextvibes/cli/cmd/project/plan/suggest-refinement"
	"github.com/spf13/cobra"
)

// PlanCmd represents the base command for the 'plan' subcommand group.
var PlanCmd = &cobra.Command{
	Use:     "plan",
	Short:   "Commands for backlog grooming and sprint planning.",
	Aliases: []string{"scrum"},
}

func init() {
	PlanCmd.AddCommand(refine.RefineCmd)
	PlanCmd.AddCommand(suggest.SuggestRefinementCmd)
	// Future subcommands like 'link' and 'board' will be added here.
}
