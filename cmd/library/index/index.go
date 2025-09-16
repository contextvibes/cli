// cmd/library/index/index.go
package index

import (
	"bufio"
	_ "embed"
	"encoding/json"
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

var (
	indexPathTHEA     string
	indexPathTemplate string
	indexPathOut      string
)

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
var IndexCmd = &cobra.Command{
	Use:     "index --thea-path <path> --template-path <path> [-o <output-file>]",
	Example: `  contextvibes library index --thea-path ../THEA/docs -o manifest.json`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if err := os.WriteFile(indexPathOut, jsonData, 0o600); err != nil {
			return fmt.Errorf("failed to write index file to %s: %w", indexPathOut, err)
		}

		presenter.Success("Successfully created document manifest at: %s", indexPathOut)
		return nil
	},
}

func processDirectory(
	rootPath, baseDirName string,
	processedFiles map[string]bool,
	logger *slog.Logger,
) ([]DocumentMetadata, error) {
	var metadataList []DocumentMetadata
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(
		absRootPath,
		func(currentPath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
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
			if parseErr == nil && docMeta != nil {
				metadataList = append(metadataList, *docMeta)
			}
			return nil
		},
	)
	return metadataList, err
}

func parseFrontMatterAndDerive(
	filePath, rootPath, baseDirName string,
	fileInfo fs.FileInfo,
) (*DocumentMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
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
		return nil, nil
	}

	var fmData tempFrontMatter
	if err := yaml.Unmarshal([]byte(strings.Join(frontMatterLines, "\n")), &fmData); err != nil {
		return nil, err
	}
	if strings.TrimSpace(fmData.Title) == "" {
		return nil, nil
	}

	relPath, _ := filepath.Rel(rootPath, filePath)
	ext := filepath.Ext(relPath)
	id := strings.TrimSuffix(relPath, ext)

	docMeta := &DocumentMetadata{
		ID:               id,
		FileExtension:    strings.TrimPrefix(ext, "."),
		Title:            fmData.Title,
		LastModifiedDate: fileInfo.ModTime().UTC().Format(time.RFC3339),
		// ... (other fields) ...
	}
	return docMeta, nil
}

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
