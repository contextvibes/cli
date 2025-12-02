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

// CheckOnMainBranchStep verifies that the current branch is the main branch.
type CheckOnMainBranchStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
}

// Description returns a description of the step.
func (s *CheckOnMainBranchStep) Description() string {
	return "Verify current branch is the main branch"
}

// Execute runs the step.
func (s *CheckOnMainBranchStep) Execute(
	_ context.Context,
) error {
	return nil
} // PreCheck does all the work.

// PreCheck performs pre-execution checks.
func (s *CheckOnMainBranchStep) PreCheck(ctx context.Context) error {
	mainBranch := s.GitClient.MainBranchName()

	currentBranch, err := s.GitClient.GetCurrentBranchName(ctx)
	if err != nil {
		s.Presenter.Error("Could not determine current branch: %v", err)

		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch != mainBranch {
		//nolint:err113 // Dynamic error is appropriate here.
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

// CheckAndPromptStashStep checks for a clean workspace and offers to stash changes.
type CheckAndPromptStashStep struct {
	GitClient *git.GitClient
	Presenter PresenterInterface
	AssumeYes bool
	DidStash  bool // ADDED: This field will track if a stash was performed.
}

// Description returns a description of the step.
func (s *CheckAndPromptStashStep) Description() string {
	return "Check for a clean working directory (and offer to stash)"
}

// Execute runs the step.
func (s *CheckAndPromptStashStep) Execute(
	_ context.Context,
) error {
	return nil
} // PreCheck does all the work.

// PreCheck performs pre-execution checks.
func (s *CheckAndPromptStashStep) PreCheck(ctx context.Context) error {
	isClean, err := s.GitClient.IsWorkingDirClean(ctx)
	if err != nil {
		s.Presenter.Error("Failed to check working directory status: %v", err)

		return fmt.Errorf("failed to check working directory status: %w", err)
	}

	if isClean {
		return nil
	}

	if s.AssumeYes {
		//nolint:err113 // Dynamic error is appropriate here.
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
		return fmt.Errorf("confirmation prompt failed: %w", err)
	}

	if !confirmed {
		//nolint:err113 // Dynamic error is appropriate here.
		err := errors.New("workflow aborted by user at stash prompt")
		s.Presenter.Info(err.Error())

		return err
	}

	s.Presenter.Step("Stashing uncommitted changes...")

	//nolint:noinlineerr // Inline check is standard for stash.
	if err := s.GitClient.StashPush(ctx); err != nil {
		s.Presenter.Error("Failed to stash changes: %v", err)

		return fmt.Errorf("stash failed: %w", err)
	}

	s.Presenter.Success("✓ Changes stashed.")
	s.DidStash = true // ADDED: Record that a stash was successfully performed.

	return nil
}

// UpdateMainBranchStep updates the main branch from the remote.
type UpdateMainBranchStep struct {
	GitClient *git.GitClient
}

// Description returns a description of the step.
func (s *UpdateMainBranchStep) Description() string {
	return "Update main branch from remote (pull --rebase)"
}

// PreCheck performs pre-execution checks.
func (s *UpdateMainBranchStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *UpdateMainBranchStep) Execute(ctx context.Context) error {
	err := s.GitClient.PullRebase(ctx, s.GitClient.MainBranchName())
	if err != nil {
		return fmt.Errorf("pull rebase failed: %w", err)
	}

	return nil
}

// CreateAndPushBranchStep creates and pushes a new branch.
type CreateAndPushBranchStep struct {
	GitClient  *git.GitClient
	BranchName string
}

// Description returns a description of the step.
func (s *CreateAndPushBranchStep) Description() string {
	return fmt.Sprintf("Create and push new branch '%s'", s.BranchName)
}

// PreCheck performs pre-execution checks.
func (s *CreateAndPushBranchStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *CreateAndPushBranchStep) Execute(ctx context.Context) error {
	err := s.GitClient.CreateAndSwitchBranch(ctx, s.BranchName, "")
	if err != nil {
		return fmt.Errorf("failed to create and switch branch: %w", err)
	}

	err = s.GitClient.PushAndSetUpstream(ctx, s.BranchName)
	if err != nil {
		return fmt.Errorf("failed to push and set upstream: %w", err)
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
			var err error

			branchName, err = getBranchNameInput(presenter, assumeYes)
			if err != nil {
				return "", err
			}
		}

		if !validationEnabled {
			return branchName, nil
		}

		valid, err := validateBranch(ctx, branchName, cfg, gitClient, presenter)
		if err != nil {
			return "", err
		}

		if valid {
			return branchName, nil
		}

		branchName = "" // Reset to prompt again
	}
}

func getBranchNameInput(presenter PresenterInterface, assumeYes bool) (string, error) {
	if assumeYes {
		//nolint:err113 // Dynamic error is appropriate here.
		return "", errors.New(
			"branch name must be provided via --branch flag in non-interactive mode",
		)
	}

	name, err := presenter.PromptForInput(
		"Enter new branch name (e.g., feature/TASK-123-new-thing)",
	)
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	return name, nil
}

func validateBranch(
	ctx context.Context,
	branchName string,
	cfg *config.Config,
	gitClient *git.GitClient,
	presenter PresenterInterface,
) (bool, error) {
	pattern := cfg.Validation.BranchName.Pattern
	if pattern == "" {
		pattern = config.DefaultBranchNamePattern
	}

	//nolint:varnamelen // 're' is standard for regex.
	re, err := regexp.Compile(pattern)
	if err != nil {
		presenter.Error("Invalid branch name validation regex in config: %s", pattern)

		return false, fmt.Errorf("invalid validation pattern: %w", err)
	}

	if !re.MatchString(branchName) {
		presenter.Error("Invalid branch name format: '%s'", branchName)
		presenter.Advice("Branch name must match the pattern: %s", pattern)

		return false, nil
	}

	exists, err := gitClient.LocalBranchExists(ctx, branchName)
	if err != nil {
		return false, fmt.Errorf("failed to check branch existence: %w", err)
	}

	if exists {
		presenter.Error("A local branch named '%s' already exists.", branchName)

		return false, nil
	}

	return true, nil
}
