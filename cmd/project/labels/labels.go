// Package labels provides commands to manage project labels.
package labels

import (
	"github.com/contextvibes/cli/cmd/project/labels/create"
	"github.com/spf13/cobra"
)

// LabelsCmd represents the base command for the 'labels' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var LabelsCmd = &cobra.Command{
	Use:     "labels",
	Short:   "Manage project labels.",
	Aliases: []string{"label"},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	LabelsCmd.AddCommand(create.CreateCmd)
}
