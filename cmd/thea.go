// cmd/thea.go
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	// Internal imports.
	"github.com/contextvibes/cli/internal/thea" // Assuming this is your THEA client package
	"github.com/contextvibes/cli/internal/ui"

	// External imports.
	"github.com/spf13/cobra"
)

// --- Parent 'thea' Command Definition ---.
var theaCmd = &cobra.Command{
	Use:   "thea",
	Short: "Interact with THEA framework artifacts and services.",
	Long: `Provides commands to fetch artifacts, list available resources (future), 
and other interactions related to the THEA framework.`,
	// No RunE needed if it only groups subcommands and doesn't execute itself.
}

// --- Subcommand 'get-artifact' Definitions ---

// Flags for get-artifact.
var (
	artifactVersionHintFlag string
	outputFilePathFlag      string
	forceOutputFlag         bool
)

// Hardcoded default values for THEA Service Configuration used by get-artifact.
const (
	getArtifactDefaultManifestURL       = "https://raw.githubusercontent.com/contextvibes/THEA/main/thea-manifest.json"
	getArtifactDefaultRawContentBaseURL = "https://raw.githubusercontent.com/contextvibes/THEA"
	getArtifactDefaultArtifactRef       = "main"
	getArtifactDefaultRequestTimeout    = 30 * time.Second
)

var getArtifactCmd = &cobra.Command{
	Use:   "get-artifact <artifact-id>",
	Short: "Fetches a specific artifact document from the THEA framework repository.",
	Long: `Downloads a specified artifact (e.g., playbook, template, guide)
from the central THEA framework repository using its unique artifact ID.

The artifact ID typically follows a path-like structure (e.g., "playbooks/initiation/kickoff").
The fetched content is saved to a local file. Default THEA repository URLs are used.`,
	Example: `  contextvibes thea get-artifact playbooks/project_initiation/master_strategic_kickoff_prompt -o kickoff_template.md
  contextvibes thea get-artifact docs/style-guide --version v1.2.0`,
	Args: cobra.ExactArgs(1), // artifact-id
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // Assumes AppLogger is initialized globally
		if logger == nil {
			logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
			logger.Warn("AppLogger was nil, using basic stderr logger for get-artifact.")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context() // Use command's context
		artifactID := args[0]

		hardcodedTHEACfg := &thea.THEAServiceConfig{
			ManifestURL:        getArtifactDefaultManifestURL,
			RawContentBaseURL:  getArtifactDefaultRawContentBaseURL,
			DefaultArtifactRef: getArtifactDefaultArtifactRef,
			RequestTimeout:     getArtifactDefaultRequestTimeout,
		}

		theaClt, errClient := thea.NewClient(ctx, hardcodedTHEACfg, logger)
		if errClient != nil {
			presenter.Error("Failed to initialize THEA artifact service: %v", errClient)

			return fmt.Errorf("initializing THEA artifact service: %w", errClient)
		}

		presenter.Info("Fetching artifact '%s' from THEA framework repository...", artifactID)
		if artifactVersionHintFlag != "" {
			presenter.Info("Using version hint: '%s'", artifactVersionHintFlag)
		}

		content, err := theaClt.FetchArtifactContentByID(ctx, artifactID, artifactVersionHintFlag)
		if err != nil {
			presenter.Error("Failed to fetch artifact '%s': %v", artifactID, err)

			return fmt.Errorf("fetching artifact '%s': %w", artifactID, err)
		}

		outputPath := outputFilePathFlag
		if outputPath == "" {
			manifest, manifestErr := theaClt.LoadManifest(ctx)
			var artifactFilename string
			if manifestErr == nil && manifest != nil {
				artDetails, artErr := manifest.GetArtifactByID(artifactID)
				if artErr == nil && artDetails != nil {
					if artDetails.DefaultTargetPath != "" {
						artifactFilename = artDetails.DefaultTargetPath
					} else {
						if artDetails.FileExtension != "" {
							artifactFilename = artDetails.ID + "." + artDetails.FileExtension
						} else {
							artifactFilename = artDetails.ID
						}
						artifactFilename = filepath.Base(artifactFilename)
					}
				}
			}
			if artifactFilename == "" {
				baseName := filepath.Base(artifactID)
				if baseName == "." || baseName == "/" || baseName == "" {
					baseName = "fetched_thea_artifact"
				}
				if ext := filepath.Ext(baseName); ext != "" {
					artifactFilename = baseName
				} else {
					artifactFilename = baseName + ".md"
				}
			}
			outputPath = artifactFilename
			presenter.Info("No output file specified with -o, will save to: %s (in current directory)", outputPath)
		}

		outputDir := filepath.Dir(outputPath)
		if outputDir != "" && outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				presenter.Error("Failed to create output directory '%s': %v", outputDir, err)

				return fmt.Errorf("creating output directory %s: %w", outputDir, err)
			}
		}

		if _, statErr := os.Stat(outputPath); statErr == nil { // File exists
			if !forceOutputFlag {
				presenter.Error("Output file '%s' already exists. Use --force to overwrite.", outputPath)

				return fmt.Errorf("output file %s already exists", outputPath)
			}
			presenter.Info("Output file '%s' exists, overwriting due to --force.", outputPath)
		} else if !os.IsNotExist(statErr) { // Some other error (e.g. permissions)
			presenter.Error("Error checking output file '%s': %v", outputPath, statErr)

			return fmt.Errorf("checking output file %s: %w", outputPath, statErr)
		}

		err = os.WriteFile(outputPath, []byte(content), 0644)
		if err != nil {
			presenter.Error("Failed to write artifact to '%s': %v", outputPath, err)

			return fmt.Errorf("writing artifact to %s: %w", outputPath, err)
		}

		presenter.Success("Successfully fetched artifact '%s' and saved to: %s", artifactID, outputPath)
		logger.InfoContext(ctx, "Artifact fetched successfully",
			slog.String("artifact_id", artifactID),
			slog.String("output_path", outputPath))

		return nil
	},
}

// --- init Function ---.
func init() {
	// Add the parent 'thea' command to the root command
	rootCmd.AddCommand(theaCmd)

	// Add 'get-artifact' as a subcommand of 'thea'
	theaCmd.AddCommand(getArtifactCmd)

	// Define flags for 'get-artifact'
	getArtifactCmd.Flags().StringVarP(&artifactVersionHintFlag, "version", "v", "", "Version hint (e.g., git tag/branch) for the artifact.")
	getArtifactCmd.Flags().StringVarP(&outputFilePathFlag, "output", "o", "", "Path to save the fetched artifact. If empty, uses a default name.")
	getArtifactCmd.Flags().BoolVarP(&forceOutputFlag, "force", "f", false, "Overwrite the output file if it already exists.")
}
