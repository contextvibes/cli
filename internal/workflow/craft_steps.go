package workflow

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/contextvibes/cli/internal/git"
)

const (
	// filePermUserRW represents read/write permissions for the user (0600).
	filePermUserRW = 0o600
)

//go:embed assets/commit_prompt.md.tpl
var commitPromptTemplate string

// GenerateCommitPromptStep reads the state and outputs the AI prompt.
type GenerateCommitPromptStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

type promptData struct {
	Branch string
	Diff   string
}

// Description returns a description of the step.
func (s *GenerateCommitPromptStep) Description() string {
	return "Generate AI prompt"
}

// PreCheck performs pre-execution checks.
func (s *GenerateCommitPromptStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *GenerateCommitPromptStep) Execute(ctx context.Context) error {
	// 1. Get Context Data
	currentBranch, err := s.GitClient.GetCurrentBranchName(ctx)
	if err != nil {
		//nolint:wrapcheck // Wrapping is handled by caller.
		return err
	}

	diff, _, err := s.GitClient.GetDiffCached(ctx)
	if err != nil {
		//nolint:wrapcheck // Wrapping is handled by caller.
		return err
	}

	// 2. Prepare Template Data
	data := promptData{
		Branch: currentBranch,
		Diff:   diff,
	}

	// 3. Execute Template
	tmpl, err := template.New("commit_prompt").Parse(commitPromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute prompt template: %w", err)
	}

	// 4. Output to File
	outputFile := "_contextvibes.md"
	if err := os.WriteFile(outputFile, buf.Bytes(), filePermUserRW); err != nil {
		return fmt.Errorf("failed to write prompt to %s: %w", outputFile, err)
	}

	s.Presenter.Success("Prompt generated: %s", outputFile)
	s.Presenter.Info("Pass this file to your AI to generate the commit message.")

	return nil
}
