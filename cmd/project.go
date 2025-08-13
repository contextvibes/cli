// cmd/project.go
package cmd

import (
	"github.com/spf13/cobra"
)

// projectCmd represents the base command for the 'project' domain.
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage your current task (the 'what'). Start, save, and finish work here.",
	Long:  `The project domain represents the management of a specific unit of work. It is the "what"—the task at hand, the feature being built, the bug being fixed.`,
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
