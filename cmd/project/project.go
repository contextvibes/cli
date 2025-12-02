// Package project groups commands related to project planning and management.
package project

import (
	"github.com/contextvibes/cli/cmd/project/board"
	"github.com/contextvibes/cli/cmd/project/describe"
	"github.com/contextvibes/cli/cmd/project/exportupstream"
	"github.com/contextvibes/cli/cmd/project/issues"
	"github.com/contextvibes/cli/cmd/project/labels"
	"github.com/contextvibes/cli/cmd/project/plan"
	"github.com/spf13/cobra"
)

// ProjectCmd represents the base command for the 'project' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Commands for project planning and management (the 'why').",
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	ProjectCmd.AddCommand(describe.DescribeCmd)
	ProjectCmd.AddCommand(issues.IssuesCmd)
	ProjectCmd.AddCommand(plan.PlanCmd)
	ProjectCmd.AddCommand(labels.LabelsCmd)
	ProjectCmd.AddCommand(board.BoardCmd)
	ProjectCmd.AddCommand(exportupstream.ExportUpstreamCmd)
}
