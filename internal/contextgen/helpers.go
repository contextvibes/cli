// Package contextgen provides helpers for generating project context reports.
package contextgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
)

//nolint:mnd // 1024 is standard buffer size.
const bufferSize = 1024

const (
	// FilePermReadWrite is 0o644.
	FilePermReadWrite = 0o644
)

func isFileBinary(filePath string) (bool, error) {
	//nolint:gosec // Reading file to check type is intended.
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	buffer := make([]byte, bufferSize)

	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	return bytes.Contains(buffer[:n], []byte{0}), nil
}

// GenerateReportHeader generates the header for the context report.
func GenerateReportHeader(promptFile, defaultTitle, defaultTask string) (string, error) {
	searchPaths := []string{
		filepath.Join("docs", "prompts", promptFile),
		filepath.Join("..", "thea", "building-blocks", "prompts", promptFile),
	}
	for _, path := range searchPaths {
		_, err := os.Stat(path)
		if err == nil {
			//nolint:gosec // Reading prompt file is intended.
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return "", fmt.Errorf("failed to read prompt file %s: %w", path, readErr)
			}

			return string(content), nil
		}
	}
	// Fallback prompt is now more descriptive.
	return fmt.Sprintf(
		"# AI Prompt: %s\n\nYour task is to: %s\n\n---\n",
		defaultTitle,
		defaultTask,
	), nil
}

// ExportBook exports a set of files to a single markdown file.
func ExportBook(
	ctx context.Context,
	execClient *exec.ExecutorClient,
	outputFile, title string,
	excludePatterns []string,
	paths ...string,
) (err error) {
	//nolint:gosec // Writing to output file is intended.
	outFile, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermReadWrite)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	defer func() {
		closeErr := outFile.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", closeErr)
		}
	}()

	_, err = fmt.Fprintf(outFile, "\n---\n## Book: %s\n\n", title)
	if err != nil {
		return fmt.Errorf("failed to write book header: %w", err)
	}

	gitArgs := append([]string{"ls-files", "--"}, paths...)

	gitFilesBytes, _, err := execClient.CaptureOutput(ctx, ".", "git", gitArgs...)
	if err != nil {
		return fmt.Errorf("failed to list git files: %w", err)
	}

	files := strings.Split(gitFilesBytes, "\n")

	return processFiles(outFile, files, excludePatterns)
}

func processFiles(writer io.Writer, files []string, excludePatterns []string) error {
fileLoop:
	for _, file := range files {
		if file == "" {
			continue
		}

		if shouldExclude(file, excludePatterns) {
			continue fileLoop
		}

		isBinary, checkErr := isFileBinary(file)
		if checkErr != nil {
			fmt.Fprintf(os.Stderr, "[WARN] Could not check file type for %s: %v. Skipping.\n", file, checkErr)

			continue
		}

		if isBinary {
			fmt.Fprintf(os.Stderr, "[INFO] Skipping binary file: %s\n", file)

			continue
		}

		err := appendFileContent(writer, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func shouldExclude(file string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(file, pattern) {
			fmt.Fprintf(os.Stderr, "[INFO] Excluding file in directory '%s': %s\n", pattern, file)

			return true
		}

		matched, _ := filepath.Match(pattern, file)
		if matched {
			fmt.Fprintf(os.Stderr, "[INFO] Excluding file matching pattern '%s': %s\n", pattern, file)

			return true
		}
	}

	return false
}

func appendFileContent(writer io.Writer, file string) error {
	//nolint:gosec // Reading source files is intended.
	content, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read file %s: %w", file, err)
	}

	ext := filepath.Ext(file)
	lang := strings.TrimPrefix(ext, ".")

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("======== FILE: %s ========\n", file))
	builder.WriteString("```" + lang + "\n")
	builder.Write(content)
	builder.WriteString("\n```\n")
	builder.WriteString(fmt.Sprintf("======== END FILE: %s ========\n\n", file))

	_, err = writer.Write([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("failed to write content for file %s: %w", file, err)
	}

	return nil
}
