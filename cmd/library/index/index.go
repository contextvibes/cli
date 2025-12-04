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

//go:embed index.md.tpl
var indexLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	indexPathTHEA     string
	indexPathTemplate string
	indexPathOut      string
)

// ErrSkipDocument is returned when a document should be skipped during indexing.
var ErrSkipDocument = errors.New("skip document")

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

// IndexCmd represents the index command
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var IndexCmd = &cobra.Command{
	Use:     "index --thea-path <path> --template-path <path> [-o <output-file>]",
	Example: `  contextvibes library index --thea-path ../THEA/docs -o manifest.json`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		logger := globals.AppLogger

		var allMetadata []DocumentMetadata
		processedFiles := make(map[string]bool)

		if indexPathTHEA != "" {
			theaMetadata, err := processDirectory(indexPathTHEA, "THEA", processedFiles, logger)
			if err != nil {
				presenter.Error("Error processing THEA directory: %v", err)
			}
			allMetadata = append(allMetadata, theaMetadata...)
		}

		if indexPathTemplate != "" {
			templateMetadata, err := processDirectory(
				indexPathTemplate,
				"Template",
				processedFiles,
				logger,
			)
			if err != nil {
				presenter.Error("Error processing Template directory: %v", err)
			}
			allMetadata = append(allMetadata, templateMetadata...)
		}

		jsonData, err := json.MarshalIndent(allMetadata, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
		}

		//nolint:mnd,noinlineerr // 0600 is standard file permission, inline check is standard.
		if err := os.WriteFile(indexPathOut, jsonData, 0o600); err != nil {
			return fmt.Errorf("failed to write index file to %s: %w", indexPathOut, err)
		}

		presenter.Success("Successfully created document manifest at: %s", indexPathOut)

		return nil
	},
}

func processDirectory(
	rootPath, baseDirName string,
	_ map[string]bool,
	logger *slog.Logger,
) ([]DocumentMetadata, error) {
	var metadataList []DocumentMetadata

	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	err = filepath.WalkDir(
		absRootPath,
		//nolint:varnamelen // 'd' is standard for DirEntry.
		func(currentPath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				//nolint:nilerr // Returning nil to continue walking is intended.
				return nil
			}

			if !(strings.HasSuffix(d.Name(), ".md")) {
				return nil
			}

			fileInfo, _ := d.Info()
			docMeta, parseErr := parseFrontMatterAndDerive(
				currentPath,
				absRootPath,
				baseDirName,
				fileInfo,
			)

			if errors.Is(parseErr, ErrSkipDocument) {
				return nil
			}

			if parseErr != nil {
				// Log warning but continue walking
				logger.Warn("Failed to parse document", "path", currentPath, "error", parseErr)

				return nil
			}

			if docMeta != nil {
				metadataList = append(metadataList, *docMeta)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return metadataList, nil
}

func parseFrontMatterAndDerive(
	filePath, rootPath, _ string,
	fileInfo fs.FileInfo,
) (*DocumentMetadata, error) {
	//nolint:gosec // Reading file is intended.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	//nolint:errcheck // Defer close is sufficient.
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var frontMatterLines []string

	inFrontMatter := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if !inFrontMatter {
				inFrontMatter = true

				continue
			}

			break
		}

		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		}
	}

	if !inFrontMatter || len(frontMatterLines) == 0 {
		return nil, ErrSkipDocument
	}

	var fmData tempFrontMatter
	if err := yaml.Unmarshal([]byte(strings.Join(frontMatterLines, "\n")), &fmData); err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	if strings.TrimSpace(fmData.Title) == "" {
		return nil, ErrSkipDocument
	}

	relPath, _ := filepath.Rel(rootPath, filePath)
	ext := filepath.Ext(relPath)
	id := strings.TrimSuffix(relPath, ext)

	//nolint:exhaustruct // Partial initialization is sufficient.
	docMeta := &DocumentMetadata{
		ID:               id,
		FileExtension:    strings.TrimPrefix(ext, "."),
		Title:            fmData.Title,
		LastModifiedDate: fileInfo.ModTime().UTC().Format(time.RFC3339),
		// ... (other fields) ...
	}

	return docMeta, nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(indexLongDescription, nil)
	if err != nil {
		panic(err)
	}

	IndexCmd.Short = desc.Short
	IndexCmd.Long = desc.Long
	IndexCmd.Flags().
		StringVar(&indexPathTHEA, "thea-path", "", "Path to the THEA directory to index.")
	IndexCmd.Flags().
		StringVar(&indexPathTemplate, "template-path", "", "Path to the template directory to index.")
	IndexCmd.Flags().
		StringVarP(&indexPathOut, "output", "o", "project_manifest.json", "Output path for the JSON manifest.")
}
