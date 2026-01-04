// Package index provides the command to index project documentation.
package index

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	filePermUserRW = 0o600
	magicNumber    = 100
)

//go:embed index.md.tpl
var indexLongDescription string

var (
	// ErrSkipDocument is returned when a document should be skipped during indexing.
	ErrSkipDocument = errors.New("skip document")
)

// DocumentMetadata represents the metadata extracted from a document.
type DocumentMetadata struct {
	ID                string   `json:"id"`
	FileExtension     string   `json:"fileExtension"`
	Title             string   `json:"title"`
	ArtifactVersion   string   `json:"artifactVersion,omitempty"`
	Summary           string   `json:"summary,omitempty"`
	UsageGuidance     []string `json:"usageGuidance,omitempty"`
	Owner             string   `json:"owner,omitempty"`
	CreatedDate       string   `json:"createdDate,omitempty"`
	LastModifiedDate  string   `json:"lastModifiedDate,omitempty"`
	DefaultTargetPath string   `json:"defaultTargetPath,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	SourceFilePath    string   `json:"sourceFilePath"`
}

// tempFrontMatter is used for unmarshalling the YAML front matter.
type tempFrontMatter struct {
	Title             string   `yaml:"title"`
	ArtifactVersion   string   `yaml:"artifactVersion"`
	Summary           string   `yaml:"summary"`
	UsageGuidance     []string `yaml:"usageGuidance"`
	Owner             string   `yaml:"owner"`
	CreatedDate       string   `yaml:"createdDate"`
	LastModifiedDate  string   `yaml:"lastModifiedDate"`
	DefaultTargetPath string   `yaml:"defaultTargetPath"`
	Tags              []string `yaml:"tags"`
}

// NewIndexCmd creates and configures the `index` command.
func NewIndexCmd() *cobra.Command {
	// Use local variables for flags to avoid global state.
	var indexPathTHEA, indexPathTemplate, indexPathOut string

	cmd := &cobra.Command{
		Use:     "index --thea-path <path> --template-path <path> [-o <output-file>]",
		Short:   "Indexes documentation files into a JSON manifest.",
		Example: `  contextvibes library index --thea-path ../THEA/docs -o manifest.json`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Pass flag values to the actual run function.
			return runIndex(cmd, indexPathTHEA, indexPathTemplate, indexPathOut)
		},
		// Boilerplate
		GroupID:                    "",
		Long:                       "", // Will be set from embedded doc
		Aliases:                    []string{},
		SuggestFor:                 []string{},
		ValidArgs:                  []string{},
		ValidArgsFunction:          nil,
		ArgAliases:                 []string{},
		BashCompletionFunction:     "",
		Deprecated:                 "",
		Annotations:                nil,
		Version:                    "",
		PersistentPreRun:           nil,
		PersistentPreRunE:          nil,
		PreRun:                     nil,
		PreRunE:                    nil,
		Run:                        nil,
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{UnknownFlags: true},
		CompletionOptions:          cobra.CompletionOptions{DisableDefaultCmd: true},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              true,
		SilenceUsage:               true,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          true,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}

	desc, err := cmddocs.ParseAndExecute(indexLongDescription, nil)
	if err == nil {
		cmd.Long = desc.Long
	}

	cmd.Flags().StringVar(&indexPathTHEA, "thea-path", "", "Path to the THEA directory to index.")
	cmd.Flags().StringVar(&indexPathTemplate, "template-path", "", "Path to the template directory to index.")
	cmd.Flags().StringVarP(&indexPathOut, "output", "o", "project_manifest.json", "Output path for the JSON manifest.")

	return cmd
}

// runIndex executes the core logic of the index command.
func runIndex(cmd *cobra.Command, theaPath, templatePath, outPath string) error {
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
	logger := globals.AppLogger

	allMetadata := make([]DocumentMetadata, 0, magicNumber)

	if theaPath != "" {
		theaMetadata, err := processDirectory(theaPath, logger)
		if err != nil {
			presenter.Error("Error processing THEA directory: %v", err)
		}

		allMetadata = append(allMetadata, theaMetadata...)
	}

	if templatePath != "" {
		templateMetadata, err := processDirectory(templatePath, logger)
		if err != nil {
			presenter.Error("Error processing Template directory: %v", err)
		}

		allMetadata = append(allMetadata, templateMetadata...)
	}

	jsonData, err := json.MarshalIndent(allMetadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
	}

	if err := os.WriteFile(outPath, jsonData, filePermUserRW); err != nil {
		return fmt.Errorf("failed to write index file to %s: %w", outPath, err)
	}

	presenter.Success("Successfully created document manifest at: %s", outPath)

	return nil
}

// processDirectory walks a directory and parses metadata from markdown files.
func processDirectory(rootPath string, logger *slog.Logger) ([]DocumentMetadata, error) {
	var metadataList []DocumentMetadata

	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", rootPath, err)
	}

	err = filepath.WalkDir(absRootPath, func(currentPath string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr // Stop walking on file system errors.
		}

		if dirEntry.IsDir() || !strings.HasSuffix(dirEntry.Name(), ".md") {
			return nil // Continue walking.
		}

		fileInfo, err := dirEntry.Info()
		if err != nil {
			logger.Warn("Failed to get file info, skipping", "path", currentPath, "error", err)

			return nil // Continue walking.
		}

		docMeta, parseErr := parseFrontMatterAndDerive(currentPath, absRootPath, fileInfo)
		if errors.Is(parseErr, ErrSkipDocument) {
			return nil // Intentionally skip this document and continue.
		}

		if parseErr != nil {
			logger.Warn("Failed to parse document, skipping", "path", currentPath, "error", parseErr)

			return nil // Continue walking.
		}

		metadataList = append(metadataList, *docMeta)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", rootPath, err)
	}

	return metadataList, nil
}

// parseFrontMatterAndDerive opens a file and orchestrates the parsing of its metadata.
func parseFrontMatterAndDerive(filePath, rootPath string, fileInfo fs.FileInfo) (*DocumentMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	defer func() { _ = file.Close() }()

	frontMatterLines, found := extractFrontMatter(bufio.NewScanner(file))
	if !found {
		return nil, ErrSkipDocument
	}

	fmData, err := parseFrontMatterData(frontMatterLines)
	if err != nil {
		return nil, fmt.Errorf("parsing front matter YAML: %w", err)
	}

	if strings.TrimSpace(fmData.Title) == "" {
		return nil, ErrSkipDocument // Skip if title is missing.
	}

	return buildDocumentMetadata(fmData, filePath, rootPath, fileInfo)
}

// extractFrontMatter scans a file line-by-line to find and return the front matter content.
func extractFrontMatter(scanner *bufio.Scanner) ([]string, bool) {
	var lines []string

	inFrontMatter := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if !inFrontMatter {
				inFrontMatter = true
			} else {
				// End of front matter found.
				return lines, true
			}
		} else if inFrontMatter {
			lines = append(lines, line)
		}
	}

	// Reached end of file.
	return lines, inFrontMatter
}

// parseFrontMatterData unmarshals the raw front matter lines into a struct.
func parseFrontMatterData(frontMatterLines []string) (*tempFrontMatter, error) {
	var fmData tempFrontMatter

	yamlContent := strings.Join(frontMatterLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &fmData); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}

	return &fmData, nil
}

// buildDocumentMetadata constructs the final metadata object from parsed and derived data.
func buildDocumentMetadata(
	fmData *tempFrontMatter,
	filePath,
	rootPath string,
	fileInfo fs.FileInfo,
) (*DocumentMetadata, error) {
	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("could not determine relative path: %w", err)
	}

	ext := filepath.Ext(relPath)
	id := strings.TrimSuffix(relPath, ext)

	return &DocumentMetadata{
		ID:                id,
		FileExtension:     strings.TrimPrefix(ext, "."),
		Title:             fmData.Title,
		LastModifiedDate:  fileInfo.ModTime().UTC().Format(time.RFC3339),
		ArtifactVersion:   fmData.ArtifactVersion,
		Summary:           fmData.Summary,
		UsageGuidance:     fmData.UsageGuidance,
		Owner:             fmData.Owner,
		CreatedDate:       fmData.CreatedDate,
		DefaultTargetPath: fmData.DefaultTargetPath,
		Tags:              fmData.Tags,
		SourceFilePath:    relPath,
	}, nil
}
