package workflow

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
)

// --- Step 1: Check if on the main branch ---
type CheckOnMainBranchStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

func (s *CheckOnMainBranchStep) Description() string {
	return "Verify current branch is the main branch"
}

func (s *CheckOnMainBranchStep) Execute(
	ctx context.Context,
) error {
	return nil
} // PreCheck does all the work.
func (s *CheckOnMainBranchStep) PreCheck(ctx context.Context) error {
	mainBranch := s.GitClient.MainBranchName()
	currentBranch, err := s.GitClient.GetCurrentBranchName(ctx)
	if err != nil {
		s.Presenter.Error("Could not determine current branch: %v", err)
		return err
	}
	if currentBranch != mainBranch {
		err := fmt.Errorf(
			"must be run from the main branch ('%s'), but you are on '%s'",
			mainBranch,
			currentBranch,
		)
		s.Presenter.Error(err.Error())
		return err
	}
	return nil
}

// --- Step 2: Check for clean workspace and offer to stash ---
type CheckAndPromptStashStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	AssumeYes bool
	DidStash  bool // ADDED: This field will track if a stash was performed.
}

func (s *CheckAndPromptStashStep) Description() string {
	return "Check for a clean working directory (and offer to stash)"
}

func (s *CheckAndPromptStashStep) Execute(
	ctx context.Context,
) error {
	return nil
} // PreCheck does all the work.
func (s *CheckAndPromptStashStep) PreCheck(ctx context.Context) error {
	isClean, err := s.GitClient.IsWorkingDirClean(ctx)
	if err != nil {
		s.Presenter.Error("Failed to check working directory status: %v", err)
		return err
	}
	if isClean {
		return nil
	}

	if s.AssumeYes {
		err := errors.New(
			"working directory is not clean and cannot prompt in non-interactive mode",
		)
		s.Presenter.Error(err.Error())
		s.Presenter.Advice(
			"Please commit or stash your changes before running with the --yes flag.",
		)
		return err
	}

	confirmed, err := s.Presenter.PromptForConfirmation(
		"Your working directory has uncommitted changes. Stash them to proceed?",
	)
	if err != nil {
		return err
	}

	if !confirmed {
		err := errors.New("workflow aborted by user at stash prompt")
		s.Presenter.Info(err.Error())
		return err
	}

	s.Presenter.Step("Stashing uncommitted changes...")
	if err := s.GitClient.StashPush(ctx); err != nil {
		s.Presenter.Error("Failed to stash changes: %v", err)
		return err
	}
	s.Presenter.Success("✓ Changes stashed.")
	s.DidStash = true // ADDED: Record that a stash was successfully performed.
	return nil
}

// --- Step 3: Update the main branch from remote ---
type UpdateMainBranchStep struct {
	GitClient *git.GitClient
}

func (s *UpdateMainBranchStep) Description() string {
	return "Update main branch from remote (pull --rebase)"
}
func (s *UpdateMainBranchStep) PreCheck(ctx context.Context) error { return nil }
func (s *UpdateMainBranchStep) Execute(ctx context.Context) error {
	return s.GitClient.PullRebase(ctx, s.GitClient.MainBranchName())
}

// --- Step 4: Create and push the new branch ---
type CreateAndPushBranchStep struct {
	GitClient  *git.GitClient
	BranchName string
}

func (s *CreateAndPushBranchStep) Description() string {
	return fmt.Sprintf("Create and push new branch '%s'", s.BranchName)
}
func (s *CreateAndPushBranchStep) PreCheck(ctx context.Context) error { return nil }
func (s *CreateAndPushBranchStep) Execute(ctx context.Context) error {
	if err := s.GitClient.CreateAndSwitchBranch(ctx, s.BranchName, ""); err != nil {
		return err
	}
	if err := s.GitClient.PushAndSetUpstream(ctx, s.BranchName); err != nil {
		return err
	}
	return nil
}

// GetValidatedBranchName is a helper function that can be used by the command
// before initializing the workflow. It's not a step itself.
func GetValidatedBranchName(
	ctx context.Context,
	branchNameFlag string,
	cfg *config.Config,
	presenter PresenterInterface,
	gitClient *git.GitClient,
	assumeYes bool,
) (string, error) {
	branchName := strings.TrimSpace(branchNameFlag)
	validationRule := cfg.Validation.BranchName
	validationEnabled := validationRule.Enable == nil || *validationRule.Enable

	for {
		if branchName == "" {
			if assumeYes {
				return "", errors.New(
					"branch name must be provided via --branch flag in non-interactive mode",
				)
			}
			var promptErr error
			branchName, promptErr = presenter.PromptForInput(
				"Enter new branch name (e.g., feature/TASK-123-new-thing)",
			)
			if promptErr != nil {
				return "", promptErr
			}
		}

		if !validationEnabled {
			return branchName, nil
		}

		pattern := validationRule.Pattern
		if pattern == "" {
			pattern = config.DefaultBranchNamePattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			presenter.Error("Invalid branch name validation regex in config: %s", pattern)
			return "", fmt.Errorf("invalid validation pattern: %w", err)
		}

		if re.MatchString(branchName) {
			exists, err := gitClient.LocalBranchExists(ctx, branchName)
			if err != nil {
				return "", err
			}
			if exists {
				presenter.Error("A local branch named '%s' already exists.", branchName)
				branchName = "" // Reset to prompt again
				continue
			}
			return branchName, nil
		}

		presenter.Error("Invalid branch name format: '%s'", branchName)
		presenter.Advice("Branch name must match the pattern: %s", pattern)
		branchName = "" // Reset to prompt again
	}
}
