// Package factory groups commands related to the mechanical execution of the workflow.
package factory

import (
	"github.com/contextvibes/cli/cmd/factory/apply"
	"github.com/contextvibes/cli/cmd/factory/bootstrap"
	"github.com/contextvibes/cli/cmd/factory/commit"
	"github.com/contextvibes/cli/cmd/factory/deploy"
	"github.com/contextvibes/cli/cmd/factory/diff"
	"github.com/contextvibes/cli/cmd/factory/finish"
	init_cmd "github.com/contextvibes/cli/cmd/factory/init"
	"github.com/contextvibes/cli/cmd/factory/kickoff"
	"github.com/contextvibes/cli/cmd/factory/plan"
	"github.com/contextvibes/cli/cmd/factory/scaffold"
	"github.com/contextvibes/cli/cmd/factory/scrub"
	"github.com/contextvibes/cli/cmd/factory/setupidentity"
	"github.com/contextvibes/cli/cmd/factory/status"
	"github.com/contextvibes/cli/cmd/factory/sync"
	"github.com/contextvibes/cli/cmd/factory/tidy"
	"github.com/contextvibes/cli/cmd/factory/tools"
	"github.com/contextvibes/cli/cmd/factory/upgradecli"
	"github.com/spf13/cobra"
)

// FactoryCmd represents the base command for the 'factory' subcommand group.
var FactoryCmd = &cobra.Command{
	Use:   "factory",
	Short: "Commands for your workflow and machinery (the 'how').",
}

func init() {
	FactoryCmd.AddCommand(bootstrap.BootstrapCmd)
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
	FactoryCmd.AddCommand(tools.ToolsCmd)
	FactoryCmd.AddCommand(scaffold.ScaffoldCmd)
	FactoryCmd.AddCommand(upgradecli.UpgradeCliCmd)
}
