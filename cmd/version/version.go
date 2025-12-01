// Package version provides the version command.
package version

import (
	"fmt"

	"github.com/contextvibes/cli/internal/globals"
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ContextVibes CLI",
	Run: func(cmd *cobra.Command, _ []string) {
		// Use Fprintf to write to the configured output stream, satisfying forbidigo.
		fmt.Fprintf(cmd.OutOrStdout(), "ContextVibes CLI version %s\n", globals.AppVersion)
	},
}
