package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/contextvibes/cli/internal/ui"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the Context Vibes CLI",
	Long:  `Display the version number of the Context Vibes CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use cmd.OutOrStdout() so that output can be captured in tests
		p := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Stdin)
		p.Summary(fmt.Sprintf("Context Vibes CLI Version: %s", AppVersion))
		return nil
	},
}

// init is called after all the variable declarations in the package have evaluated
// their initializers, and after all imported packages have been initialized.
// It is used here to add the versionCmd to the rootCmd.
// The AppVersion variable is expected to be initialized in root.go's init().
func init() {
	rootCmd.AddCommand(versionCmd)
}
