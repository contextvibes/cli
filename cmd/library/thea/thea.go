// Package thea provides commands to interact with the THEA framework.
package thea

import (
	"github.com/contextvibes/cli/cmd/library/thea/getartifact"
	"github.com/spf13/cobra"
)

// TheaCmd represents the base command for the 'thea' subcommand group.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design..
var TheaCmd = &cobra.Command{
	Use:   "thea",
	Short: "Interact with the THEA framework (fetch artifacts, etc).",
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	TheaCmd.AddCommand(getartifact.GetArtifactCmd)
}
