// Package project groups commands related to project planning and management.
package project

import (
	"github.com/contextvibes/cli/cmd/project/board"
	"github.com/contextvibes/cli/cmd/project/describe"
	"github.com/contextvibes/cli/cmd/project/issues"
	"github.com/contextvibes/cli/cmd/project/labels"
	"github.com/contextvibes/cli/cmd/project/onboard" // Added
	"github.com/contextvibes/cli/cmd/project/plan"
	"github.com/contextvibes/cli/cmd/project/summary"
	"github.com/spf13/cobra"
)

// NewProjectCmd creates and configures the `project` command.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "project",
		Short:             "Commands for project planning and management (the 'why').",
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(describe.DescribeCmd)
	cmd.AddCommand(issues.IssuesCmd)
	cmd.AddCommand(plan.PlanCmd)
	cmd.AddCommand(labels.LabelsCmd)
	cmd.AddCommand(board.BoardCmd)
	cmd.AddCommand(summary.SummaryCmd)
	cmd.AddCommand(onboard.OnboardCmd) // Added

	return cmd
}
