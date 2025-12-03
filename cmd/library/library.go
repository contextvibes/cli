// Package library provides commands to manage the knowledge library.
package library

import (
	"github.com/contextvibes/cli/cmd/library/index"
	"github.com/contextvibes/cli/cmd/library/systemprompt"
	"github.com/contextvibes/cli/cmd/library/thea" // Imported
	"github.com/spf13/cobra"
)

// LibraryCmd represents the base command for the 'library' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var LibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Commands for knowledge and standards (the 'where').",
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	LibraryCmd.AddCommand(index.IndexCmd)
	LibraryCmd.AddCommand(thea.TheaCmd) // Uncommented/Added
	LibraryCmd.AddCommand(systemprompt.SystemPromptCmd)
}
