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
	"github.com/contextvibes/cli/cmd/factory/squash"
	"github.com/contextvibes/cli/cmd/factory/status"
	"github.com/contextvibes/cli/cmd/factory/sync"
	"github.com/contextvibes/cli/cmd/factory/tidy"
	"github.com/contextvibes/cli/cmd/factory/tools"
	"github.com/contextvibes/cli/cmd/factory/upgradecli"
	"github.com/spf13/cobra"
)

// NewFactoryCmd creates and configures the `factory` command and its subcommands.
func NewFactoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "factory",
		Short: "Commands for your workflow and machinery (the 'how').",
		Long: `The factory commands provide the tools to manage the mechanics of your
development workflow. These are the commands that are typically chained
together by higher-level commands.`,
		Example:       "contextvibes factory --help",
		GroupID:       "core",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	// Define Command Groups and add subcommands
	cmd.AddGroup(&cobra.Group{ID: "factory", Title: "Factory Operations"})
	addSubCommands(cmd)

	return cmd
}

// addSubCommands is a helper to keep the NewFactoryCmd cleaner.
func addSubCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		bootstrap.BootstrapCmd,
		init_cmd.InitCmd,
		kickoff.KickoffCmd,
		commit.CommitCmd,
		status.StatusCmd,
		diff.DiffCmd,
		sync.SyncCmd,
		finish.FinishCmd,
		tidy.TidyCmd,
		plan.PlanCmd,
		apply.ApplyCmd,
		deploy.DeployCmd,
		scrub.ScrubCmd,
		setupidentity.NewSetupIdentityCmd(),
		tools.ToolsCmd,
		scaffold.ScaffoldCmd,
		upgradecli.UpgradeCliCmd,
		squash.NewSquashCmd(),
	)
}
