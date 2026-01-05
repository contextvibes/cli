package workflow

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/contextvibes/cli/internal/git"
)

const (
	// filePermUserRW represents read/write permissions for the user (0600).
	filePermUserRW = 0o600
)

//go:embed assets/commit_prompt.md.tpl
var commitPromptTemplate string

//go:embed assets/pr_description_prompt.md.tpl
var prDescriptionPromptTemplate string

//go:embed assets/refactor_prompt.md.tpl
var refactorPromptTemplate string

//go:embed assets/review_prompt.md.tpl
var reviewPromptTemplate string

//go:embed assets/strategic_kickoff.md.tpl
var strategicKickoffPromptTemplate string

// --- Existing Steps (Commit & PR) ---

// GenerateCommitPromptStep reads the state and outputs the AI prompt.
type GenerateCommitPromptStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

type promptData struct {
	Branch string
	Diff   string
}

func (s *GenerateCommitPromptStep) Description() string {
	return "Generate AI prompt for commit message"
}

func (s *GenerateCommitPromptStep) PreCheck(_ context.Context) error { return nil }

func (s *GenerateCommitPromptStep) Execute(ctx context.Context) error {
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

	data := promptData{
		Branch: currentBranch,
		Diff:   diff,
	}

	tmpl, err := template.New("commit_prompt").Parse(commitPromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute prompt template: %w", err)
	}

	outputFile := "context_commit.md"
	if err := os.WriteFile(outputFile, buf.Bytes(), filePermUserRW); err != nil {
		return fmt.Errorf("failed to write prompt to %s: %w", outputFile, err)
	}

	s.Presenter.Success("Prompt generated: %s", outputFile)
	s.Presenter.Info("Pass this file to your AI to generate the commit message.")

	return nil
}

// GeneratePRDescriptionPromptStep generates a prompt for PR descriptions.
type GeneratePRDescriptionPromptStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

type prPromptData struct {
	Log  string
	Diff string
}

func (s *GeneratePRDescriptionPromptStep) Description() string {
	return "Generate AI prompt for PR description"
}

func (s *GeneratePRDescriptionPromptStep) PreCheck(_ context.Context) error { return nil }

func (s *GeneratePRDescriptionPromptStep) Execute(ctx context.Context) error {
	mainBranch := s.GitClient.MainBranchName()
	log, diff, err := s.GitClient.GetLogAndDiffFromMergeBase(ctx, mainBranch)
	if err != nil {
		return fmt.Errorf("failed to get branch changes against '%s': %w", mainBranch, err)
	}

	data := prPromptData{
		Log:  log,
		Diff: diff,
	}

	tmpl, err := template.New("pr_prompt").Parse(prDescriptionPromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute prompt template: %w", err)
	}

	outputFile := "context_pr.md"
	if err := os.WriteFile(outputFile, buf.Bytes(), filePermUserRW); err != nil {
		return fmt.Errorf("failed to write prompt to %s: %w", outputFile, err)
	}

	s.Presenter.Success("Prompt generated: %s", outputFile)
	s.Presenter.Info("Pass this file to your AI to generate the PR description.")

	return nil
}

// --- New Steps (Refactor & Review) ---

type codePromptData struct {
	CodeBlocks string
}

// GenerateRefactorPromptStep generates a prompt for code refactoring.
type GenerateRefactorPromptStep struct {
	Files     []string
	Presenter PresenterInterface
}

func (s *GenerateRefactorPromptStep) Description() string {
	return fmt.Sprintf("Generate AI refactoring prompt for %d file(s)", len(s.Files))
}

func (s *GenerateRefactorPromptStep) PreCheck(_ context.Context) error {
	if len(s.Files) == 0 {
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("no files specified for refactoring")
	}
	return nil
}

func (s *GenerateRefactorPromptStep) Execute(_ context.Context) error {
	codeContent, err := readFilesAsMarkdown(s.Files)
	if err != nil {
		return err
	}

	tmpl, err := template.New("refactor_prompt").Parse(refactorPromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, codePromptData{CodeBlocks: codeContent}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	outputFile := "context_refactor.md"
	if err := os.WriteFile(outputFile, buf.Bytes(), filePermUserRW); err != nil {
		return fmt.Errorf("failed to write prompt to %s: %w", outputFile, err)
	}

	s.Presenter.Success("Prompt generated: %s", outputFile)
	s.Presenter.Info("Pass this file to your AI to generate a refactoring plan.")

	return nil
}

// GenerateReviewPromptStep generates a prompt for code review.
type GenerateReviewPromptStep struct {
	Files     []string
	Presenter PresenterInterface
}

func (s *GenerateReviewPromptStep) Description() string {
	return fmt.Sprintf("Generate AI review prompt for %d file(s)", len(s.Files))
}

func (s *GenerateReviewPromptStep) PreCheck(_ context.Context) error {
	if len(s.Files) == 0 {
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("no files specified for review")
	}
	return nil
}

func (s *GenerateReviewPromptStep) Execute(_ context.Context) error {
	codeContent, err := readFilesAsMarkdown(s.Files)
	if err != nil {
		return err
	}

	tmpl, err := template.New("review_prompt").Parse(reviewPromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, codePromptData{CodeBlocks: codeContent}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	outputFile := "context_review.md"
	if err := os.WriteFile(outputFile, buf.Bytes(), filePermUserRW); err != nil {
		return fmt.Errorf("failed to write prompt to %s: %w", outputFile, err)
	}

	s.Presenter.Success("Prompt generated: %s", outputFile)
	s.Presenter.Info("Pass this file to your AI to get a code review.")

	return nil
}

// --- Strategic Kickoff Step ---

// GenerateStrategicKickoffPromptStep generates the master prompt for a strategic kickoff.
type GenerateStrategicKickoffPromptStep struct {
	Presenter PresenterInterface
}

func (s *GenerateStrategicKickoffPromptStep) Description() string {
	return "Generate Strategic Kickoff Master Prompt"
}

func (s *GenerateStrategicKickoffPromptStep) PreCheck(_ context.Context) error { return nil }

func (s *GenerateStrategicKickoffPromptStep) Execute(_ context.Context) error {
	// In a full implementation, we might gather some initial project details here
	// to pre-fill the template. For now, we use the static template.

	tmpl, err := template.New("strategic_kickoff").Parse(strategicKickoffPromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	outputFile := "STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md"
	if err := os.WriteFile(outputFile, buf.Bytes(), filePermUserRW); err != nil {
		return fmt.Errorf("failed to write prompt to %s: %w", outputFile, err)
	}

	s.Presenter.Success("Strategic Kickoff Protocol generated: %s", outputFile)
	s.Presenter.Info("👉 Upload this file to your AI assistant to begin the strategic planning session.")

	return nil
}

// Helper to read files and format as markdown blocks
func readFilesAsMarkdown(files []string) (string, error) {
	var sb strings.Builder
	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
		}
		sb.WriteString(fmt.Sprintf("\n--- FILE: %s ---\n", filePath))
		sb.WriteString("```go\n")
		sb.Write(content)
		sb.WriteString("\n```\n")
	}
	return sb.String(), nil
}
