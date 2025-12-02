// Package board provides commands to manage project boards.
package board

import (
	"github.com/contextvibes/cli/cmd/project/board/add"
	"github.com/contextvibes/cli/cmd/project/board/list"
	"github.com/spf13/cobra"
)

// BoardCmd represents the base command for the 'board' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var BoardCmd = &cobra.Command{
	Use:     "board",
	Short:   "Manage project boards.",
	Aliases: []string{"boards"},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	BoardCmd.AddCommand(list.ListCmd)
	BoardCmd.AddCommand(add.AddCmd)
}
