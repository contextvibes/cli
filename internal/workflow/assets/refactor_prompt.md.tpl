cat << 'EOF' > internal/workflow/assets/review_prompt.md.tpl
# AI Meta-Prompt: Code Quality Review

## Your Role
You are an expert Go developer and a meticulous code reviewer.

## The Task
Analyze the following code. Provide a comprehensive code review in Markdown format.

## Rules for Your Review
1. **Start with a High-Level Summary.**
2. **Categorize Feedback:** Correctness, Simplicity, Idiomatic Go, Testing, Nitpicks.
3. **Provide Actionable Suggestions:** Show "before" and "after" snippets.

## The Code
{{ .CodeBlocks }}
EOF

# 2. Update internal/workflow/ai_steps.go with new steps
cat << 'EOF' > internal/workflow/ai_steps.go
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

//go:embed assets/refactor_prompt.md.tpl
var refactorPromptTemplate string

//go:embed assets/review_prompt.md.tpl
var reviewPromptTemplate string

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

	prompt := fmt.Sprintf(`
# Role
You are a senior software engineer.

# Goal
Write a clear and comprehensive Pull Request description based on the following changes.

# Instructions
1.  **Summary**: Write a high-level summary of the problem solved and the solution.
2.  **Changes**: Use a bulleted list to detail specific changes.
3.  **Format**: Output raw Markdown suitable for a GitHub PR body.

# Commit History
%s

# Code Diff
~~~diff
%s
~~~
`, log, diff)

	outputFile := "context_pr.md"
	if err := os.WriteFile(outputFile, []byte(prompt), filePermUserRW); err != nil {
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
EOF

# 3. Update codemod command
cat << 'EOF' > cmd/product/codemod/codemod.go
// Package codemod provides the command to apply code modifications.
package codemod

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/codemod"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed codemod.md.tpl
var codemodLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	codemodScriptPath string
	aiAssist          bool
)

// CodemodCmd represents the codemod command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CodemodCmd = &cobra.Command{
	Use: "codemod [files...] [--script <file.json>] | --ai",
	Example: `  contextvibes product codemod --script ./plan.json
  contextvibes product codemod cmd/main.go --ai`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		// 1. Handle AI Mode (Generate Prompt)
		if aiAssist {
			runner := workflow.NewRunner(presenter, globals.AssumeYes)
			return runner.Run(
				ctx,
				"Generating AI Refactoring Prompt",
				&workflow.GenerateRefactorPromptStep{
					Files:     args,
					Presenter: presenter,
				},
			)
		}

		// 2. Standard Codemod Logic (Apply Script)
		scriptToLoad := codemodScriptPath
		if scriptToLoad == "" {
			scriptToLoad = "codemod.json"
		}

		scriptData, err := os.ReadFile(scriptToLoad)
		if err != nil {
			return fmt.Errorf("failed to read codemod script '%s': %w", scriptToLoad, err)
		}

		var script codemod.ChangeScript
		err = json.Unmarshal(scriptData, &script)
		if err != nil {
			return fmt.Errorf("failed to parse codemod script JSON: %w", err)
		}

		for _, fileChangeSet := range script {
			presenter.Header("Processing target: %s", fileChangeSet.FilePath)

			contentBytes, err := os.ReadFile(fileChangeSet.FilePath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to read target file: %w", err)
			}
			currentContent := string(contentBytes)

			//nolint:varnamelen // 'op' is standard for operation.
			for _, op := range fileChangeSet.Operations {
				switch op.Type {
				case "regex_replace":
					re, err := regexp.Compile(op.FindRegex)
					if err != nil {
						return fmt.Errorf("invalid regex '%s': %w", op.FindRegex, err)
					}
					currentContent = re.ReplaceAllString(currentContent, op.ReplaceWith)
				case "create_or_overwrite":
					if op.Content != nil {
						currentContent = *op.Content
					}
				}
			}

			if !globals.AssumeYes {
				confirmed, err := presenter.PromptForConfirmation(
					fmt.Sprintf("Write changes to %s?", fileChangeSet.FilePath),
				)
				if err != nil || !confirmed {
					continue
				}
			}
			//nolint:mnd // 0600 is standard file permission.
			err = os.WriteFile(fileChangeSet.FilePath, []byte(currentContent), 0o600)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			globals.AppLogger.Info("Applied codemod", "file", fileChangeSet.FilePath)
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(codemodLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CodemodCmd.Short = desc.Short
	CodemodCmd.Long = desc.Long
	CodemodCmd.Flags().
		StringVarP(&codemodScriptPath, "script", "s", "", "Path to the JSON codemod script file")
	CodemodCmd.Flags().
		BoolVar(&aiAssist, "ai", false, "Generate a prompt for an AI to refactor the specified files")
}
EOF

# 4. Update quality command
cat << 'EOF' > cmd/product/quality/quality.go
package quality

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/config/assets"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	// aiContextFile is the local constant for the AI context file.
	aiContextFile = "_contextvibes.md"

	// Quality check modes.
	modeEssential  = "essential"
	modeStrict     = "strict"
	modeStyle      = "style"
	modeComplexity = "complexity"
	modeSecurity   = "security"
	modeLocal      = "local"
)

// CheckResult holds the outcome of a single quality check.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
	Err     error  `json:"-"`
}

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var aiAssist bool

// NewQualityCmd creates the quality command and its subcommands.
func NewQualityCmd() *cobra.Command {
	var qualityMode string

	cmd := &cobra.Command{
		Use:   "quality [paths...] | --ai",
		Args:  cobra.ArbitraryArgs,
		Short: "Runs a series of quality checks against the codebase.",
		Long: `The quality command runs a configurable pipeline of code quality checks.

MODES:
  - essential:  (Default) Basic sanity checks (build, vet, vuln). Fast & recommended for local dev.
  - strict:     Enforces strict linting rules using the embedded 'strict' configuration.
  - style:      Focuses purely on code style, formatting, and naming conventions.
  - complexity: Checks for cyclomatic complexity and function length.
  - security:   Deep security scan (gitleaks, gosec).
  - local:      Uses the project's own .golangci.yml configuration (if present).

EXAMPLES:
  contextvibes product quality                  # Run essential checks on whole project
  contextvibes product quality --mode strict    # Run strict checks
  contextvibes product quality cmd/factory --ai # Generate AI review prompt for package`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuality(cmd, args, qualityMode)
		},
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}

	usage := fmt.Sprintf("Quality check mode (%s)", strings.Join(supportedModesAsString(), "|"))
	cmd.Flags().StringVarP(&qualityMode, "mode", "m", modeEssential, usage)
	cmd.Flags().BoolVar(&aiAssist, "ai", false, "Generate a prompt for an AI to review the code")

	_ = cmd.RegisterFlagCompletionFunc("mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return supportedModesAsString(), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.AddCommand(serveCmd)

	return cmd
}

func runQuality(cmd *cobra.Command, args []string, qualityMode string) error {
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
	ctx := cmd.Context()

	// 1. Handle AI Mode
	if aiAssist {
		runner := workflow.NewRunner(presenter, globals.AssumeYes)
		return runner.Run(
			ctx,
			"Generating AI Code Review Prompt",
			&workflow.GenerateReviewPromptStep{
				Files:     args,
				Presenter: presenter,
			},
		)
	}

	// 2. Standard Quality Logic
	isValidMode := false
	for _, validMode := range supportedModesAsString() {
		if qualityMode == validMode {
			isValidMode = true
			break
		}
	}

	if !isValidMode {
		presenter.Error("Invalid quality mode specified: " + qualityMode)
		return fmt.Errorf("invalid quality mode: %s", qualityMode)
	}

	_, err := RunQualityChecks(ctx, presenter, qualityMode, args)
	return err
}

// RunQualityChecks executes the full pipeline of code quality checks.
//
//nolint:cyclop,funlen // Orchestration logic requires complexity.
func RunQualityChecks(
	ctx context.Context,
	presenter *ui.Presenter,
	mode string,
	args []string,
) ([]CheckResult, error) {
	presenter.Header("--- Code Quality Pipeline ---")
	presenter.Info("Mode: %s", mode)

	results := []CheckResult{}

	// 1. Determine which checks to run based on mode
	runBuild := mode == modeEssential || mode == modeStrict || mode == modeLocal
	runVet := mode == modeEssential || mode == modeStrict || mode == modeLocal
	runVuln := mode == modeEssential || mode == modeStrict || mode == modeSecurity || mode == modeLocal
	runGitleaks := mode == modeSecurity

	// Determine Linter Config
	var linterConfig config.AssetType

	runLint := true

	switch mode {
	case modeEssential:
		runLint = false // Essential is just build/vet/vuln
	case modeStrict:
		linterConfig = config.AssetLintStrict
	case modeStyle:
		linterConfig = config.AssetLintStyle
	case modeComplexity:
		linterConfig = config.AssetLintComplexity
	case modeSecurity:
		linterConfig = config.AssetLintSecurity
	case modeLocal:
		linterConfig
