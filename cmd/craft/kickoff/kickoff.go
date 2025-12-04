// Package kickoff provides the command to start a strategic kickoff session.
package kickoff

import (
	_ "embed"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed kickoff.md.tpl
var kickoffLongDescription string

// KickoffCmd represents the craft kickoff command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var KickoffCmd = &cobra.Command{
	Use:   "kickoff",
	Short: "Starts an AI-guided strategic project planning session.",

	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		// This is the new home for the logic from the old 'kickoff --strategic'.
		// The full implementation of the orchestrator would be called from here.

		presenter.Summary("Initiating Strategic Kickoff Session...")
		presenter.Info(
			"This feature will guide you through generating a master prompt for your AI.",
		)
		presenter.Warning("Full implementation is pending.")

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(kickoffLongDescription, nil)
	if err != nil {
		panic(err)
	}

	KickoffCmd.Short = desc.Short
	KickoffCmd.Long = desc.Long
}
