// cmd/index.go
package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time" // For handling dates and file modification times

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for the 'index' command (remain the same)
var (
	indexPathTHEA     string
	indexPathTemplate string
	indexPathOut      string
)

// DocumentMetadata holds the parsed front matter and derived info from a file.
// This struct will be marshalled to JSON for the output of 'contextvibes index'.
type DocumentMetadata struct {
	ID                string   `json:"id"`                          // e.g., "category/path/artifact-name" (derived: relative path w/o ext)
	FileExtension     string   `json:"fileExtension"`               // e.g., "md", "yml" (derived from source file)
	Title             string   `json:"title"`                       // From front matter
	ArtifactVersion   string   `json:"artifactVersion,omitempty"`   // From front matter (semver)
	Summary           string   `json:"summary,omitempty"`           // From front matter
	UsageGuidance     []string `json:"usageGuidance,omitempty"`     // From front matter (can be a list)
	Owner             string   `json:"owner,omitempty"`             // From front matter (role/nickname)
	CreatedDate       string   `json:"createdDate,omitempty"`       // From front matter (ISO 8601)
	LastModifiedDate  string   `json:"lastModifiedDate,omitempty"`  // From front matter (ISO 8601) or file mod time
	DefaultTargetPath string   `json:"defaultTargetPath,omitempty"` // From front matter, or derived (id + fileExtension)
	Tags              []string `json:"tags,omitempty"`              // From front matter
	SourceFilePath    string   `json:"sourceFilePath"`              // Relative path of the source file processed (baseDirName/relPath)
}

// tempFrontMatter is used to unmarshal the YAML front matter from files.
// It includes all fields that 'contextvibes index' expects to find.
type tempFrontMatter struct {
	Title             string   `yaml:"title"`
	ArtifactVersion   string   `yaml:"artifactVersion"`
	Summary           string   `yaml:"summary"`
	UsageGuidance     []string `yaml:"usageGuidance"`
	Owner             string   `yaml:"owner"`
	CreatedDate       string   `yaml:"createdDate"`       // Expected in ISO 8601 string format
	LastModifiedDate  string   `yaml:"lastModifiedDate"`  // Expected in ISO 8601 string format
	DefaultTargetPath string   `yaml:"defaultTargetPath"` // Optional in front matter
	Tags              []string `yaml:"tags"`
}

// indexCmd represents the base 'index' command.
var indexCmd = &cobra.Command{
	Use:   "index --thea-path <path-to-thea> --template-path <path-to-template> [-o <output-file>]",
	Short: "Indexes documents (e.g., THEA, templates) to create a structured JSON manifest.",
	Long: `Crawls the specified THEA framework and project template directories, 
parses metadata (including version, summary, usage guidance, owner, dates, tags)
from YAML front matter in supported files (e.g., .md, .yaml), and generates a 
structured JSON manifest file.

The 'id' of each document in the manifest is derived from its relative file path 
minus the extension. 'fileExtension' is also captured.

This manifest can be used for various purposes, including providing context to AI models
or for programmatic access to document metadata.`,
	Example: `  contextvibes index --thea-path ../THEA-main/docs --template-path ../THEA-main/templates -o project_manifest.json`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context()

		presenter.Summary("Starting Document Indexing Process...")

		if indexPathTHEA == "" && indexPathTemplate == "" { // Adjusted: allow one or both
			return errors.New("--thea-path and/or --template-path flag is required")
		}

		logger.InfoContext(ctx, "Index command initiated",
			slog.String("thea_path", indexPathTHEA),
			slog.String("template_path", indexPathTemplate),
			slog.String("output_path", indexPathOut),
		)

		allMetadata := []DocumentMetadata{}
		processedFiles := make(map[string]bool) // To avoid processing same file twice if paths overlap

		// Process the THEA directory
		if indexPathTHEA != "" {
			presenter.Step("Scanning THEA directory: %s", indexPathTHEA)
			// Assuming "THEA" as the base name for paths derived from this source
			theaMetadata, err := processDirectory(indexPathTHEA, "THEA", processedFiles, logger)
			if err != nil {
				// Log error but continue if possible, processDirectory should handle skippable errors
				presenter.Error("Error processing THEA directory: %v (continuing if possible)", err)
			}
			allMetadata = append(allMetadata, theaMetadata...)
			presenter.Info("Found %d documents in THEA directory.", len(theaMetadata))
		}

		// Process the template directory
		if indexPathTemplate != "" {
			presenter.Step("Scanning Template directory: %s", indexPathTemplate)
			// Assuming "Template" as the base name for paths derived from this source
			templateMetadata, err := processDirectory(indexPathTemplate, "Template", processedFiles, logger)
			if err != nil {
				presenter.Error("Error processing Template directory: %v (continuing if possible)", err)
			}
			allMetadata = append(allMetadata, templateMetadata...)
			presenter.Info("Found %d documents in Template directory.", len(templateMetadata))
		}

		if len(allMetadata) == 0 {
			presenter.Warning("No documents found or processed. The output file will be empty or not created.")
			// Optionally, write an empty JSON array or handle as needed
			// For now, let it proceed to write an empty array if that's the case.
		}

		presenter.Step("Generating JSON manifest...")
		jsonData, err := json.MarshalIndent(allMetadata, "", "  ")
		if err != nil {
			logger.ErrorContext(ctx, "Failed to marshal metadata to JSON", slog.String("error", err.Error()))
			return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
		}

		err = os.WriteFile(indexPathOut, jsonData, 0644)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to write index file", slog.String("path", indexPathOut), slog.String("error", err.Error()))
			return fmt.Errorf("failed to write index file to %s: %w", indexPathOut, err)
		}

		presenter.Success("Successfully created document manifest at: %s (%d documents indexed)", indexPathOut, len(allMetadata))
		return nil
	},
}

// processDirectory walks a directory and extracts metadata from supported files.
// baseDirName is used to prefix the 'SourceFilePath' for clarity on origin.
// processedFiles map is used to avoid processing the same absolute file path multiple times.
func processDirectory(rootPath string, baseDirName string, processedFiles map[string]bool, logger *slog.Logger) ([]DocumentMetadata, error) {
	var metadataList []DocumentMetadata
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		logger.Error("Could not get absolute path for root", slog.String("rootPath", rootPath), slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not get absolute path for %s: %w", rootPath, err)
	}

	err = filepath.WalkDir(absRootPath, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			logger.Warn("Error accessing path during walk, skipping entry", slog.String("path", currentPath), slog.String("error", walkErr.Error()))
			if d != nil && d.IsDir() && errors.Is(walkErr, fs.ErrPermission) { // Example: skip permission-denied dirs
				return fs.SkipDir
			}
			return nil // Skip this problematic entry, try to continue
		}

		// Check if already processed (handles symlinks or overlapping input paths)
		absCurrentPath, err := filepath.Abs(currentPath)
		if err != nil {
			logger.Warn("Could not get absolute path for item, skipping", slog.String("path", currentPath), slog.String("error", err.Error()))
			return nil
		}
		if processedFiles[absCurrentPath] {
			return nil // Already processed
		}

		// We only care about files, and specific types (e.g., .md for front matter)
		// Add other extensions if they can also contain YAML front matter (e.g., .yaml, .yml itself)
		if !d.IsDir() && (strings.HasSuffix(d.Name(), ".md") || strings.HasSuffix(d.Name(), ".markdown")) {
			fileInfo, statErr := d.Info()
			if statErr != nil {
				logger.Warn("Could not get file info, skipping file", slog.String("path", currentPath), slog.String("error", statErr.Error()))
				return nil
			}

			docMeta, parseErr := parseFrontMatterAndDerive(currentPath, absRootPath, baseDirName, fileInfo)
			if parseErr != nil {
				logger.Warn("Could not parse metadata, skipping file", slog.String("path", currentPath), slog.String("error", parseErr.Error()))
				// Optionally, use presenter.Warning here if user needs to know about individual skips.
				return nil // Skip files we can't parse or that don't meet criteria (e.g., no title)
			}

			if docMeta != nil { // parseFrontMatterAndDerive returns nil if essential fields missing (e.g. title)
				metadataList = append(metadataList, *docMeta)
				processedFiles[absCurrentPath] = true // Mark as processed
			}
		}
		return nil
	})

	if err != nil {
		// This error is from WalkDir itself if it encountered a non-skippable error.
		logger.Error("Fatal error walking directory", slog.String("rootPath", absRootPath), slog.String("error", err.Error()))
		return nil, fmt.Errorf("error walking directory %s: %w", rootPath, err)
	}
	return metadataList, nil
}

// parseFrontMatterAndDerive reads a file, parses its YAML front matter,
// and derives additional metadata fields.
func parseFrontMatterAndDerive(filePath string, rootPath string, baseDirName string, fileInfo fs.FileInfo) (*DocumentMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	var frontMatterLines []string
	inFrontMatter := false
	linesRead := 0
	maxFrontMatterLines := 100 // Increased slightly

	for scanner.Scan() {
		linesRead++
		if linesRead > maxFrontMatterLines && inFrontMatter {
			return nil, fmt.Errorf("front matter block in %s seems unusually large or unterminated after %d lines", filePath, maxFrontMatterLines)
		}
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				break // End of front matter
			}
		}
		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line) // Keep original lines for YAML parser
		}
		if !inFrontMatter && linesRead > 10 && len(frontMatterLines) == 0 { // Heuristic: if no '---' after 10 lines
			return nil, nil // Assume no front matter
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}
	if !inFrontMatter || len(frontMatterLines) == 0 {
		return nil, nil // No valid front matter found
	}

	fmData := tempFrontMatter{}
	err = yaml.Unmarshal([]byte(strings.Join(frontMatterLines, "\n")), &fmData)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling YAML from %s: %w", filePath, err)
	}

	// Essential field: Title. If not present, skip this document.
	if strings.TrimSpace(fmData.Title) == "" {
		return nil, nil // No title, consider it not a valid indexed document
	}

	// Derive ID and FileExtension
	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		// Fallback to full path if Rel fails, though this should ideally not happen if rootPath is an ancestor.
		AppLogger.Warn("Could not make path relative, using full path for ID basis", slog.String("filePath", filePath), slog.String("rootPath", rootPath))
		relPath = filePath
	}
	ext := filepath.Ext(relPath)
	id := strings.TrimSuffix(relPath, ext)
	// Normalize ID to use forward slashes for consistency, regardless of OS
	id = filepath.ToSlash(id)

	// Handle LastModifiedDate: use from front matter if present, else file system mod time.
	lastModified := fmData.LastModifiedDate
	if lastModified == "" && fileInfo != nil {
		lastModified = fileInfo.ModTime().UTC().Format(time.RFC3339)
	}

	// Handle DefaultTargetPath: use from front matter if present, else derive from id + ext.
	defaultTarget := fmData.DefaultTargetPath
	if defaultTarget == "" {
		defaultTarget = id + ext // or id + fmData.FileExtension if we add that to tempFrontMatter
		// but ext is directly available here.
		defaultTarget = filepath.ToSlash(defaultTarget) // Ensure consistent slashes
	}

	docMeta := &DocumentMetadata{
		ID:                id,
		FileExtension:     strings.TrimPrefix(ext, "."), // Store extension without the dot
		Title:             fmData.Title,
		ArtifactVersion:   fmData.ArtifactVersion,
		Summary:           fmData.Summary,
		UsageGuidance:     fmData.UsageGuidance,
		Owner:             fmData.Owner,
		CreatedDate:       fmData.CreatedDate,
		LastModifiedDate:  lastModified,
		DefaultTargetPath: defaultTarget,
		Tags:              fmData.Tags,
		SourceFilePath:    filepath.ToSlash(filepath.Join(baseDirName, relPath)), // Store with consistent forward slashes
	}

	return docMeta, nil
}

func init() {
	rootCmd.AddCommand(indexCmd)

	indexCmd.Flags().StringVar(&indexPathTHEA, "thea-path", "", "Path to the root of a THEA-like structured directory to index.")
	indexCmd.Flags().StringVar(&indexPathTemplate, "template-path", "", "Path to the root of a project template directory to index.")
	indexCmd.Flags().StringVarP(&indexPathOut, "output", "o", "project_manifest.json", "Output path for the generated JSON manifest file.")
	// Example of marking flags required:
	// indexCmd.MarkFlagRequired("output") // Though not strictly required as it has a default
}
