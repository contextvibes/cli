package workflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/contextvibes/cli/internal/git"
)

// EnsureNotMainBranchStep ensures the user is NOT on the main branch.
type EnsureNotMainBranchStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

// Description returns a description of the step.
func (s *EnsureNotMainBranchStep) Description() string {
	return "Verify current branch is a feature branch (not main)"
}

// PreCheck performs pre-execution checks.
func (s *EnsureNotMainBranchStep) PreCheck(ctx context.Context) error {
	mainBranch := s.GitClient.MainBranchName()

	currentBranch, err := s.GitClient.GetCurrentBranchName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == mainBranch {
		s.Presenter.Error("You are on the '%s' branch.", currentBranch)
		s.Presenter.Advice("We do not commit directly to main.")
		s.Presenter.Advice("Run 'contextvibes factory kickoff' to start a feature branch.")
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("cannot run on %s branch", mainBranch)
	}

	return nil
}

// Execute runs the step.
func (s *EnsureNotMainBranchStep) Execute(_ context.Context) error { return nil }

// EnsureStagedStep ensures there are staged changes, prompting to add if necessary.
type EnsureStagedStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	AssumeYes bool
}

// Description returns a description of the step.
func (s *EnsureStagedStep) Description() string {
	return "Ensure changes are staged for commit"
}

// PreCheck performs pre-execution checks.
func (s *EnsureStagedStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *EnsureStagedStep) Execute(ctx context.Context) error {
	// 1. Check for existing staged changes
	hasStaged, err := s.GitClient.HasStagedChanges(ctx)
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}

	if hasStaged {
		return nil // Already good to go
	}

	// 2. No staged changes? Check for unstaged/untracked.
	isClean, err := s.GitClient.IsWorkingDirClean(ctx)
	if err != nil {
		return fmt.Errorf("failed to check working dir: %w", err)
	}

	if isClean {
		s.Presenter.Info("Working directory is clean. Nothing to commit.")
		//nolint:err113 // Dynamic error is appropriate here.
		return errors.New("no changes to commit")
	}

	// 3. Found unstaged changes. Prompt to stage.
	s.Presenter.Warning("No staged changes found, but the working directory has modifications.")

	if !s.AssumeYes {
		confirm, err := s.Presenter.PromptForConfirmation("Stage all changes (git add .)?")
		if err != nil {
			//nolint:wrapcheck // Wrapping is handled by caller.
			return err
		}

		if !confirm {
			s.Presenter.Info("Aborted. Please stage specific files manually.")
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("user aborted staging")
		}
	}

	s.Presenter.Step("Staging all changes...")

	err = s.GitClient.AddAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	return nil
}
