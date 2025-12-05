// Package factory groups commands related to the mechanical execution of the workflow.
package factory

import (
	"github.com/contextvibes/cli/cmd/factory/apply"
	"github.com/contextvibes/cli/cmd/factory/commit"
	"github.com/contextvibes/cli/cmd/factory/deploy"
	"github.com/contextvibes/cli/cmd/factory/diff"
	"github.com/contextvibes/cli/cmd/factory/finish"
	init_cmd "github.com/contextvibes/cli/cmd/factory/init"
	"github.com/contextvibes/cli/cmd/factory/kickoff"
	"github.com/contextvibes/cli/cmd/factory/plan"
	"github.com/contextvibes/cli/cmd/factory/scrub"
	"github.com/contextvibes/cli/cmd/factory/setupidentity"
	"github.com/contextvibes/cli/cmd/factory/status"
	"github.com/contextvibes/cli/cmd/factory/sync"
	"github.com/contextvibes/cli/cmd/factory/tidy"
	"github.com/contextvibes/cli/cmd/factory/tools" // Added
	"github.com/spf13/cobra"
)

// FactoryCmd represents the base command for the 'factory' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var FactoryCmd = &cobra.Command{
	Use:   "factory",
	Short: "Commands for your workflow and machinery (the 'how').",
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	FactoryCmd.AddCommand(init_cmd.InitCmd)
	FactoryCmd.AddCommand(kickoff.KickoffCmd)
	FactoryCmd.AddCommand(commit.CommitCmd)
	FactoryCmd.AddCommand(status.StatusCmd)
	FactoryCmd.AddCommand(diff.DiffCmd)
	FactoryCmd.AddCommand(sync.SyncCmd)
	FactoryCmd.AddCommand(finish.FinishCmd)
	FactoryCmd.AddCommand(tidy.TidyCmd)
	FactoryCmd.AddCommand(plan.PlanCmd)
	FactoryCmd.AddCommand(apply.ApplyCmd)
	FactoryCmd.AddCommand(deploy.DeployCmd)
	FactoryCmd.AddCommand(scrub.ScrubCmd)
	FactoryCmd.AddCommand(setupidentity.SetupIdentityCmd)
	FactoryCmd.AddCommand(tools.ToolsCmd) // Added
}
