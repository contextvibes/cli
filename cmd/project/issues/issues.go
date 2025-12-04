// Package issues provides commands to manage project issues.
package issues

import (
	closecmd "github.com/contextvibes/cli/cmd/project/issues/close" // Imports package closecmd
	"github.com/contextvibes/cli/cmd/project/issues/create"
	"github.com/contextvibes/cli/cmd/project/issues/list"
	"github.com/contextvibes/cli/cmd/project/issues/tree"
	"github.com/contextvibes/cli/cmd/project/issues/view"
	"github.com/spf13/cobra"
)

// IssuesCmd represents the base command for the 'issues' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var IssuesCmd = &cobra.Command{
	Use:     "issues",
	Short:   "Manage project issues (work tickets, blueprints).",
	Aliases: []string{"issue"},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	IssuesCmd.AddCommand(create.CreateCmd)
	IssuesCmd.AddCommand(list.ListCmd)
	IssuesCmd.AddCommand(view.ViewCmd)
	IssuesCmd.AddCommand(tree.TreeCmd)
	IssuesCmd.AddCommand(closecmd.CloseCmd) // Updated reference
}
