package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/contextgen"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Generates context for various development goals.",
}

var generateCommitCmd = &cobra.Command{
	Use:   "generate-commit",
	Short: "Generates context for a commit message.",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		return runGenerateCommitContext(cmd.Context(), presenter)
	},
}

var generatePrCmd = &cobra.Command{
	Use:   "generate-pr",
	Short: "Generates context for a Pull Request description.",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		return runGeneratePrContext(cmd.Context(), presenter)
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports large-scale project context for AI onboarding.",
}

var exportAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Exports a comprehensive snapshot of the entire project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		outputFile := "context_export_project.md"
		presenter.Summary("Exporting full project context to %s...", outputFile)
		header, err := contextgen.GenerateReportHeader(
			"export-project-context.md",
			"Full Project Context",
		)
		if err != nil {
			return err
		}
		if err := os.WriteFile(outputFile, []byte(header), 0o600); err != nil {
			return err
		}
		if err := contextgen.ExportBook(cmd.Context(), ExecClient, outputFile, "Project Files", LoadedAppConfig.Export.ExcludePatterns, "."); err != nil {
			return err
		}
		presenter.Success("Full project export complete.")
		return nil
	},
}

func runGenerateCommitContext(ctx context.Context, presenter *ui.Presenter) error {
	presenter.Summary("Generating context for commit message...")
	header, err := contextgen.GenerateReportHeader(
		"generate-commit-message.md",
		"Generate Conventional Commit Command",
	)
	if err != nil {
		return err
	}
	status, _, err := ExecClient.CaptureOutput(ctx, ".", "git", "status")
	if err != nil {
		return err
	}
	stagedDiff, _, err := ExecClient.CaptureOutput(ctx, ".", "git", "diff", "--staged")
	if err != nil {
		return err
	}
	unstagedDiff, _, err := ExecClient.CaptureOutput(ctx, ".", "git", "diff")
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n## Git Status\n```\n" + status + "\n```\n")
	sb.WriteString(
		"---\n## Diff of All Uncommitted Changes\n```diff\n" + stagedDiff + unstagedDiff + "\n```\n",
	)
	return os.WriteFile("context_commit.md", []byte(sb.String()), 0o600)
}

func runGeneratePrContext(ctx context.Context, presenter *ui.Presenter) error {
	presenter.Summary("Generating context for Pull Request description...")
	header, err := contextgen.GenerateReportHeader(
		"generate-pr-description.md",
		"Generate Pull Request Description",
	)
	if err != nil {
		return err
	}
	gitClient, err := git.NewClient(
		ctx,
		".",
		git.GitClientConfig{Logger: AppLogger, Executor: ExecClient.UnderlyingExecutor()},
	)
	if err != nil {
		return err
	}
	mainBranchRef := "origin/" + gitClient.MainBranchName()
	presenter.Step("Fetching latest updates from remote...")
	_ = ExecClient.Execute(ctx, ".", "git", "fetch", "origin")
	log, diff, err := gitClient.GetLogAndDiffFromMergeBase(ctx, mainBranchRef)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString(fmt.Sprintf("\n## Commit History on This Branch\n```\n%s\n```\n", log))
	sb.WriteString(
		fmt.Sprintf(
			"---\n## Full Diff for Branch (vs. %s)\n```diff\n%s\n```\n",
			mainBranchRef,
			diff,
		),
	)
	return os.WriteFile("context_pr.md", []byte(sb.String()), 0o600)
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(generateCommitCmd)
	contextCmd.AddCommand(generatePrCmd)
	contextCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportAllCmd)
}
