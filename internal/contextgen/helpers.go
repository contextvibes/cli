package contextgen

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
)

func isFileBinary(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	return bytes.Contains(buffer[:n], []byte{0}), nil
}

func GenerateReportHeader(promptFile, defaultTitle string) (string, error) {
	searchPaths := []string{
		filepath.Join("docs", "prompts", promptFile),
		filepath.Join("..", "thea", "building-blocks", "prompts", promptFile),
	}
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			//gosec:G304
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return "", fmt.Errorf("failed to read prompt file %s: %w", path, readErr)
			}
			return string(content), nil
		}
	}
	return fmt.Sprintf("# AI Prompt: %s\n\n---\n", defaultTitle), nil
}

func ExportBook(ctx context.Context, execClient *exec.ExecutorClient, outputFile, title string, excludePatterns []string, paths ...string) (err error) {
	//gosec:G304
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", closeErr)
		}
	}()

	if _, err := fmt.Fprintf(f, "\n---\n## Book: %s\n\n", title); err != nil {
		return fmt.Errorf("failed to write book header: %w", err)
	}

	gitArgs := append([]string{"ls-files", "--"}, paths...)
	gitFilesBytes, _, err := execClient.CaptureOutput(ctx, ".", "git", gitArgs...)
	if err != nil {
		return fmt.Errorf("failed to list git files: %w", err)
	}

	files := strings.Split(gitFilesBytes, "\n")

fileLoop:
	for _, file := range files {
		if file == "" {
			continue
		}

		// Perform robust exclusion matching in Go.
		for _, pattern := range excludePatterns {
			// If pattern ends with '/', treat it as a directory prefix.
			if strings.HasSuffix(pattern, "/") && strings.HasPrefix(file, pattern) {
				fmt.Fprintf(os.Stderr, "[INFO] Excluding file in directory '%s': %s\n", pattern, file)
				continue fileLoop
			}
			// Otherwise, use standard glob matching for files.
			matched, _ := filepath.Match(pattern, file)
			if matched {
				fmt.Fprintf(os.Stderr, "[INFO] Excluding file matching pattern '%s': %s\n", pattern, file)
				continue fileLoop
			}
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

		//gosec:G304
		content, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}
		ext := filepath.Ext(file)
		lang := strings.TrimPrefix(ext, ".")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("======== FILE: %s ========\n", file))
		sb.WriteString("```" + lang + "\n")
		sb.Write(content)
		sb.WriteString("\n```\n")
		sb.WriteString(fmt.Sprintf("======== END FILE: %s ========\n\n", file))
		if _, err := f.WriteString(sb.String()); err != nil {
			return fmt.Errorf("failed to write content for file %s: %w", file, err)
		}
	}
	return nil
}
