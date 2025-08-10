package contextgen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
)

func GenerateReportHeader(promptFile, defaultTitle string) (string, error) {
	searchPaths := []string{
		filepath.Join("docs", "prompts", promptFile),
		filepath.Join("..", "thea", "building-blocks", "prompts", promptFile),
	}
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return "", fmt.Errorf("failed to read prompt file %s: %w", path, readErr)
			}
			return string(content), nil
		}
	}
	return fmt.Sprintf("# AI Prompt: %s\n\n---\n", defaultTitle), nil
}

func ExportBook(ctx context.Context, execClient *exec.ExecutorClient, outputFile, title string, paths ...string) error {
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { return fmt.Errorf("failed to open output file: %w", err) }
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n---\n## Book: %s\n\n", title); err != nil {
		return fmt.Errorf("failed to write book header: %w", err)
	}
	gitArgs := append([]string{"ls-files", "--"}, paths...)
	gitFilesBytes, _, err := execClient.CaptureOutput(ctx, ".", "git", gitArgs...)
	if err != nil { return fmt.Errorf("failed to list git files: %w", err) }
	for _, file := range strings.Split(gitFilesBytes, "\n") {
		if file == "" { continue }
		content, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) { continue }
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
