// Package getartifact provides the command to fetch THEA artifacts.
package getartifact

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/thea"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals // Cobra flags require package-level variables..
var (
	versionFlag string
	outputFlag  string
	forceFlag   bool
)

// GetArtifactCmd represents the get-artifact command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var GetArtifactCmd = &cobra.Command{
	Use:     "get-artifact <artifact-id>",
	Short:   "Fetch a specific artifact from the THEA framework.",
	Example: `  contextvibes library thea get-artifact playbooks/project_initiation/master_strategic_kickoff_prompt`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()
		artifactID := args[0]

		presenter.Summary("Fetching THEA Artifact: %s", artifactID)

		// Initialize THEA Client
		//nolint:exhaustruct // Partial config is sufficient.
		cfg := &thea.ServiceConfig{
			ManifestURL:        "https://raw.githubusercontent.com/contextvibes/THEA/main/thea-manifest.json",
			RawContentBaseURL:  "https://raw.githubusercontent.com/contextvibes/THEA",
			DefaultArtifactRef: "main",
			RequestTimeout:     30 * time.Second, //nolint:mnd // 30s timeout.
		}

		client, err := thea.NewClient(ctx, cfg, globals.AppLogger)
		if err != nil {
			return fmt.Errorf("failed to initialize THEA client: %w", err)
		}

		// Fetch Content
		presenter.Step("Downloading content...")

		content, err := client.FetchArtifactContentByID(ctx, artifactID, versionFlag)
		if err != nil {
			presenter.Error("Failed to fetch artifact: %v", err)

			return fmt.Errorf("fetch failed: %w", err)
		}

		// Determine Output Path
		targetPath := outputFlag
		if targetPath == "" {
			// Default to ID + .md if no output specified (simplification)
			targetPath = filepath.Base(artifactID) + ".md"
		}

		// Check for overwrite
		_, statErr := os.Stat(targetPath)
		if statErr == nil && !forceFlag {
			presenter.Error("File '%s' already exists.", targetPath)
			presenter.Advice("Use --force to overwrite.")
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("file exists")
		}

		// Write File
		//nolint:mnd // 0600 is standard secure file permission.
		err = os.WriteFile(targetPath, []byte(content), 0o600)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		presenter.Success("Artifact saved to: %s", targetPath)

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	GetArtifactCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Version hint (git tag/branch)")
	GetArtifactCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path")
	GetArtifactCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing file")
}
