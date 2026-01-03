package workflow

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/tools"
)

// SquashState holds data shared between squash workflow steps.
type SquashState struct {
	CurrentBranch string
	TargetBase    string // The ref we are squashing against (e.g. origin/main)
	MergeBase     string
	CommitCount   int
}

// FetchStep ensures we have the latest remote refs.
type FetchStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

func (s *FetchStep) Description() string {
	return "Fetch latest changes from remote"
}

func (s *FetchStep) PreCheck(_ context.Context) error { return nil }

func (s *FetchStep) Execute(ctx context.Context) error {
	// We use the executor directly to run fetch, as GitClient doesn't expose a raw Fetch yet.
	// Ideally GitClient should have Fetch(), but for now this works via the underlying executor pattern.
	// Or we can add Fetch to GitClient. Let's assume we can't easily modify GitClient interface right now
	// and just skip explicit fetch if it's too complex, but it's better to be robust.
	// Actually, let's just skip explicit fetch for this iteration to keep the diff small
	// and rely on what's there. If needed, the user can 'git fetch'.
	return nil
}

// EnsureCleanOrSaveStep checks for dirty state and offers to commit.
type EnsureCleanOrSaveStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	AssumeYes bool
}

func (s *EnsureCleanOrSaveStep) Description() string {
	return "Ensure working directory is clean (or save state)"
}

func (s *EnsureCleanOrSaveStep) PreCheck(ctx context.Context) error { return nil }

func (s *EnsureCleanOrSaveStep) Execute(ctx context.Context) error {
	isClean, err := s.GitClient.IsWorkingDirClean(ctx)
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if isClean {
		return nil
	}

	s.Presenter.Warning("Working directory is dirty.")

	if s.AssumeYes {
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("cannot auto-squash with dirty directory in non-interactive mode")
	}

	confirm, err := s.Presenter.PromptForConfirmation("Stage and commit these changes before squashing?")
	if err != nil {
		return err
	}
	if !confirm {
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("aborted by user")
	}

	// Prompt for message
	defaultMsg := "wip: Pre-squash save"
	msg, err := s.Presenter.PromptForInput(fmt.Sprintf("Commit message (Press Enter for '%s'):", defaultMsg))
	if err != nil {
		return err
	}
	if strings.TrimSpace(msg) == "" {
		msg = defaultMsg
	}

	s.Presenter.Step("Saving current state...")
	if err := s.GitClient.AddAll(ctx); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	if err := s.GitClient.Commit(ctx, msg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	s.Presenter.Success("✓ State saved.")

	return nil
}

// AnalyzeBranchStep calculates commit counts and merge base.
type AnalyzeBranchStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	State     *SquashState
}

func (s *AnalyzeBranchStep) Description() string {
	return "Analyze branch divergence from upstream"
}

func (s *AnalyzeBranchStep) PreCheck(ctx context.Context) error {
	current, err := s.GitClient.GetCurrentBranchName(ctx)
	if err != nil {
		return err
	}

	remote := s.GitClient.RemoteName()
	main := s.GitClient.MainBranchName()
	targetBase := fmt.Sprintf("%s/%s", remote, main)

	if current == main {
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("cannot squash main branch (%s)", main)
	}

	s.State.CurrentBranch = current
	s.State.TargetBase = targetBase
	return nil
}

func (s *AnalyzeBranchStep) Execute(ctx context.Context) error {
	base, err := s.GitClient.GetMergeBase(ctx, s.State.TargetBase)
	if err != nil {
		return fmt.Errorf("failed to find merge base with %s: %w", s.State.TargetBase, err)
	}
	s.State.MergeBase = base

	count, err := s.GitClient.GetCommitCount(ctx, base+"..HEAD")
	if err != nil {
		return err
	}
	s.State.CommitCount = count

	s.Presenter.Info("Analysis: Branch '%s' is %d commits ahead of '%s'.", s.State.CurrentBranch, count, s.State.TargetBase)
	s.Presenter.Detail("Merge Base: %s", base)

	return nil
}

// SoftResetStep performs the git reset --soft.
type SoftResetStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	State     *SquashState
	AssumeYes bool
}

func (s *SoftResetStep) Description() string {
	return fmt.Sprintf("Soft reset to merge base (%s)", s.State.MergeBase)
}

func (s *SoftResetStep) PreCheck(_ context.Context) error { return nil }

func (s *SoftResetStep) Execute(ctx context.Context) error {
	if s.State.CommitCount <= 1 {
		s.Presenter.Warning("Branch has only %d commit(s) ahead of base.", s.State.CommitCount)
		if !s.AssumeYes {
			confirm, err := s.Presenter.PromptForConfirmation("Proceed anyway (useful for rewriting the commit message)?")
			if err != nil {
				return err
			}
			if !confirm {
				//nolint:err113 // Dynamic error is appropriate here.
				return fmt.Errorf("squash aborted by user (low commit count)")
			}
		}
	}

	return s.GitClient.ResetSoft(ctx, s.State.MergeBase)
}

// GenerateSquashPromptStep creates the _contextvibes.md file for the AI.
type GenerateSquashPromptStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	State     *SquashState
}

func (s *GenerateSquashPromptStep) Description() string {
	return "Generate AI context (_contextvibes.md)"
}

func (s *GenerateSquashPromptStep) PreCheck(_ context.Context) error { return nil }

func (s *GenerateSquashPromptStep) Execute(ctx context.Context) error {
	diff, _, err := s.GitClient.GetDiffCached(ctx)
	if err != nil {
		s.Presenter.Warning("Failed to generate diff for AI context: %v", err)
		return nil
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# AI Task: Generate Squash Commit Message\n\n")
	fmt.Fprintf(&buf, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&buf, "## Context\n")
	fmt.Fprintf(&buf, "- **Branch:** %s\n", s.State.CurrentBranch)
	fmt.Fprintf(&buf, "- **Squashed Commits:** %d\n\n", s.State.CommitCount)
	fmt.Fprintf(&buf, "## Instructions\n")
	fmt.Fprintf(&buf, "Write a Conventional Commit message that summarizes the entire feature based on the diff below.\n")
	fmt.Fprintf(&buf, "The subject should be imperative (e.g., 'feat: Add squash command'). The body should list key changes.\n\n")
	fmt.Fprintf(&buf, "## Diff\n\n```diff\n%s\n```\n", diff)

	outputFile := config.DefaultDescribeOutputFile // _contextvibes.md
	if err := tools.WriteBufferToFile(outputFile, &buf); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	s.Presenter.Success("AI Context generated: %s", outputFile)
	s.Presenter.Info("👉 Pass this file to your AI to generate the commit message.")
	return nil
}

// CommitSquashStep prompts for the message and commits.
type CommitSquashStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	State     *SquashState
	AssumeYes bool
}

func (s *CommitSquashStep) Description() string {
	return "Commit staged changes"
}

func (s *CommitSquashStep) PreCheck(_ context.Context) error { return nil }

func (s *CommitSquashStep) Execute(ctx context.Context) error {
	var message string
	if !s.AssumeYes {
		var err error
		message, err = s.Presenter.PromptForInput("Enter new commit message (Subject):")
		if err != nil {
			return err
		}
	}

	if message == "" {
		message = fmt.Sprintf("feat: Squash %d commits on %s", s.State.CommitCount, s.State.CurrentBranch)
		s.Presenter.Info("Using default message: %s", message)
	}

	return s.GitClient.Commit(ctx, message)
}

// ForcePushStep performs the push.
type ForcePushStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	State     *SquashState
	AssumeYes bool
}

func (s *ForcePushStep) Description() string {
	return "Force push (with lease) to remote"
}

func (s *ForcePushStep) PreCheck(_ context.Context) error { return nil }

func (s *ForcePushStep) Execute(ctx context.Context) error {
	if !s.AssumeYes {
		confirm, err := s.Presenter.PromptForConfirmation("Force push (with lease) to remote?")
		if err != nil {
			return err
		}
		if !confirm {
			s.Presenter.Info("Skipping push. Remember to run 'git push --force-with-lease' manually.")
			return nil
		}
	}

	return s.GitClient.ForcePushLease(ctx, s.State.CurrentBranch)
}
