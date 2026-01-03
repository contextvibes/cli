// Package craft provides commands for AI-assisted creative tasks.
package craft

import (
	"github.com/contextvibes/cli/cmd/craft/kickoff"
	"github.com/contextvibes/cli/cmd/craft/message"
	"github.com/contextvibes/cli/cmd/craft/prdescription"
	"github.com/contextvibes/cli/cmd/craft/refactor"
	"github.com/contextvibes/cli/cmd/craft/review"
	"github.com/spf13/cobra"
)

// CraftCmd represents the base command for the 'craft' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CraftCmd = &cobra.Command{
	Use:   "craft",
	Short: "Commands for AI-assisted creative tasks (the 'who' & 'how').",
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	CraftCmd.AddCommand(message.MessageCmd)
	CraftCmd.AddCommand(prdescription.PRDescriptionCmd)
	CraftCmd.AddCommand(kickoff.KickoffCmd)
	CraftCmd.AddCommand(refactor.RefactorCmd)
	CraftCmd.AddCommand(review.ReviewCmd)
}
