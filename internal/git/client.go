package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	osexec "os/exec" // Alias for standard library exec.ExitError
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec" // Use the new executor
)

// GitClient provides methods for interacting with a Git repository.
//
//nolint:revive // GitClient is the established name.
type GitClient struct {
	repoPath string
	gitDir   string
	config   GitClientConfig
	logger   *slog.Logger         // Logger for GitClient's specific operations
	executor exec.CommandExecutor // Uses the new CommandExecutor from internal/exec
}

// NewClient creates and initializes a new GitClient.
// The GitClientConfig's Executor field should be pre-populated,
// or validateAndSetDefaults will create a default OSCommandExecutor.
//
//nolint:funlen // Initialization logic requires length.
func NewClient(ctx context.Context, workDir string, config GitClientConfig) (*GitClient, error) {
	// validateAndSetDefaults will ensure Logger and Executor are non-nil.
	validatedConfig, err := config.validateAndSetDefaults()
	if err != nil {
		return nil, fmt.Errorf("invalid GitClientConfig: %w", err)
	}

	logger := validatedConfig.Logger     // This is the GitClient's own logger
	executor := validatedConfig.Executor // This is the exec.CommandExecutor

	// Check if the configured git executable exists using the provided executor
	if !executor.CommandExists(validatedConfig.GitExecutable) {
		//nolint:err113 // Dynamic error is appropriate here.
		err := fmt.Errorf(
			"git executable '%s' not found in PATH or specified path",
			validatedConfig.GitExecutable,
		)
		logger.ErrorContext(ctx, "GitClient setup failed: executable check",
			slog.String("source_component", "GitClient.NewClient"),
			slog.String("error", err.Error()),
			slog.String("executable_path", validatedConfig.GitExecutable))

		return nil, err
	}

	effectiveWorkDir := workDir
	if effectiveWorkDir == "" {
		effectiveWorkDir, err = os.Getwd()
		if err != nil {
			logger.ErrorContext(
				ctx,
				"GitClient setup failed: getwd",
				slog.String("error", err.Error()),
			)

			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Use the executor to find repo top-level
	topLevelCmdArgs := []string{"rev-parse", "--show-toplevel"}

	topLevel, stderr, err := executor.CaptureOutput(
		ctx,
		effectiveWorkDir,
		validatedConfig.GitExecutable,
		topLevelCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --show-toplevel",
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("initial_workdir", effectiveWorkDir))

		//nolint:err113 // Dynamic error is appropriate here.
		return nil, fmt.Errorf(
			"path '%s' is not within a Git repository (or git command '%s' failed)",
			effectiveWorkDir,
			validatedConfig.GitExecutable,
		)
	}

	repoPath := strings.TrimSpace(topLevel)

	// Use the executor to find .git directory
	gitDirCmdArgs := []string{"rev-parse", "--git-dir"}

	gitDirOutput, stderr, err := executor.CaptureOutput(
		ctx,
		repoPath,
		validatedConfig.GitExecutable,
		gitDirCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --git-dir",
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("repo_path", repoPath))

		return nil, fmt.Errorf(
			"could not determine .git directory for repo at '%s': %w",
			repoPath,
			err,
		)
	}

	gitDir := strings.TrimSpace(gitDirOutput)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoPath, gitDir)
	}

	client := &GitClient{
		repoPath: repoPath,
		gitDir:   gitDir,
		config:   validatedConfig,
		logger:   logger,
		executor: executor,
	}
	logger.InfoContext(ctx, "GitClient initialized successfully",
		slog.String("repository_path", client.repoPath),
		slog.String("git_dir", client.gitDir))

	return client, nil
}

// Path returns the repository path.
func (c *GitClient) Path() string { return c.repoPath }

// GitDir returns the .git directory path.
func (c *GitClient) GitDir() string { return c.gitDir }

// MainBranchName returns the configured main branch name.
func (c *GitClient) MainBranchName() string { return c.config.DefaultMainBranchName }

// RemoteName returns the configured remote name.
func (c *GitClient) RemoteName() string { return c.config.DefaultRemoteName }

// Logger returns the logger.
func (c *GitClient) Logger() *slog.Logger { return c.logger }

// GetRemoteURL retrieves the URL for a given remote name.
func (c *GitClient) GetRemoteURL(ctx context.Context, remoteName string) (string, error) {
	if remoteName == "" {
		//nolint:err113 // Dynamic error is appropriate here.
		return "", errors.New("remote name cannot be empty")
	}

	stdout, _, err := c.captureGitOutput(ctx, "remote", "get-url", remoteName)
	if err != nil {
		return "", fmt.Errorf("could not get URL for remote '%s': %w", remoteName, err)
	}

	return strings.TrimSpace(stdout), nil
}

// --- Public Git Operation Methods ---
// (These methods remain largely the same but now internally call c.executor methods
//  which are of type exec.CommandExecutor)

// GetCurrentBranchName returns the name of the current branch.
func (c *GitClient) GetCurrentBranchName(ctx context.Context) (string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		c.logger.ErrorContext(
			ctx,
			"Failed to get current branch name",
			"error",
			err,
			"stderr",
			strings.TrimSpace(stderr),
		)

		return "", fmt.Errorf("failed to get current branch name: %w", err)
	}

	branch := strings.TrimSpace(stdout)
	if branch == "HEAD" {
		//nolint:err113 // Dynamic error is appropriate here.
		return "", errors.New("currently in detached HEAD state")
	}

	if branch == "" {
		//nolint:err113 // Dynamic error is appropriate here.
		return "", errors.New("could not determine current branch name (empty output)")
	}

	return branch, nil
}

// AddAll stages all changes.
func (c *GitClient) AddAll(ctx context.Context) error {
	err := c.runGit(ctx, "add", ".")
	if err != nil {
		return fmt.Errorf("git add . failed: %w", err)
	}

	return nil
}

// Commit commits staged changes with a message.
func (c *GitClient) Commit(ctx context.Context, message string) error {
	if strings.TrimSpace(message) == "" {
		//nolint:err113 // Dynamic error is appropriate here.
		return errors.New("commit message cannot be empty")
	}

	err := c.runGit(ctx, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("commit command failed: %w", err)
	}

	return nil
}

// HasStagedChanges checks if there are staged changes.
func (c *GitClient) HasStagedChanges(ctx context.Context) (bool, error) {
	_, _, err := c.captureGitOutput(ctx, "diff", "--quiet", "--cached")
	if err == nil {
		return false, nil // Exit 0: no changes
	}

	var exitErr *osexec.ExitError // Use osexec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil // Exit 1: changes found
	}

	return false, fmt.Errorf("failed to determine staged status: %w", err)
}

// GetStatusShort returns the short status of the repository.
func (c *GitClient) GetStatusShort(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "status", "--short")
}

// GetDiffCached returns the cached diff.
func (c *GitClient) GetDiffCached(ctx context.Context) (string, string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "diff", "--cached")
	if err != nil {
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return stdout, stderr, nil // Exit 1 (changes found) is not an error for this func
		}

		return stdout, stderr, err // Actual error
	}

	return stdout, stderr, nil // No error, no diff
}

// GetDiffUnstaged returns the unstaged diff.
func (c *GitClient) GetDiffUnstaged(ctx context.Context) (string, string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "diff", "HEAD")
	if err != nil {
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return stdout, stderr, nil // Exit 1 (changes found) is not an error for this func
		}

		return stdout, stderr, err // Actual error
	}

	return stdout, stderr, nil // No error, no diff
}

// ListUntrackedFiles returns a list of untracked files.
func (c *GitClient) ListUntrackedFiles(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "ls-files", "--others", "--exclude-standard")
}

// IsWorkingDirClean checks if the working directory is clean.
func (c *GitClient) IsWorkingDirClean(ctx context.Context) (bool, error) {
	_, _, errDiff := c.captureGitOutput(ctx, "diff", "--quiet")
	if errDiff != nil {
		var exitErr *osexec.ExitError
		if errors.As(errDiff, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil // Unstaged changes
		}

		return false, fmt.Errorf("failed checking unstaged changes: %w", errDiff)
	}

	hasStaged, errStaged := c.HasStagedChanges(ctx)
	if errStaged != nil {
		return false, fmt.Errorf("failed checking staged changes: %w", errStaged)
	}

	if hasStaged {
		return false, nil // Staged changes
	}

	untrackedOut, _, errUntracked := c.ListUntrackedFiles(ctx)
	if errUntracked != nil {
		return false, fmt.Errorf("failed checking untracked files: %w", errUntracked)
	}

	if strings.TrimSpace(untrackedOut) != "" {
		return false, nil // Untracked files
	}

	return true, nil
}

// PullRebase pulls changes from the remote and rebases.
func (c *GitClient) PullRebase(ctx context.Context, branch string) error {
	remote := c.RemoteName()

	err := c.runGit(ctx, "pull", "--rebase", remote, branch)
	if err != nil {
		return fmt.Errorf("git pull --rebase %s %s failed: %w", remote, branch, err)
	}

	return nil
}

// IsBranchAhead checks if the local branch is ahead of the remote.
func (c *GitClient) IsBranchAhead(ctx context.Context) (bool, error) {
	stdout, _, err := c.captureGitOutput(ctx, "status", "-sb")
	if err != nil {
		return false, fmt.Errorf("failed to get status to check if branch is ahead: %w", err)
	}

	return strings.Contains(stdout, "[ahead "), nil
}

// Push pushes changes to the remote.
func (c *GitClient) Push(ctx context.Context, branch string) error {
	remote := c.RemoteName()

	args := []string{"push", remote}
	if branch != "" {
		args = append(args, branch)
	}
	// Capture output to check for "up-to-date" messages, as runGit only returns error on non-zero exit.
	_, stderr, err := c.captureGitOutput(ctx, args...) // Use captureGitOutput
	if err != nil {
		// Check if stderr (or err.Error() if it includes stderr) indicates "up-to-date"
		// This is a bit fragile. A better way would be for CaptureOutput to return specific error types.
		errMsg := strings.ToLower(err.Error() + " " + stderr) // Combine for checking
		if strings.Contains(errMsg, "everything up-to-date") ||
			strings.Contains(errMsg, "already up-to-date") {
			c.logger.InfoContext(
				ctx,
				"'git push' reported everything up-to-date.",
				"remote",
				remote,
				"branch_arg",
				branch,
			)

			return nil // Not a failure
		}

		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}

// LocalBranchExists checks if a local branch exists.
func (c *GitClient) LocalBranchExists(ctx context.Context, branchName string) (bool, error) {
	ref := "refs/heads/" + branchName

	_, _, err := c.captureGitOutput(ctx, "show-ref", "--verify", "--quiet", ref)
	if err == nil {
		return true, nil // Exit 0 means it exists
	}

	var exitErr *osexec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil // Exit 1 means it doesn't exist
	}

	return false, fmt.Errorf("failed to check existence of local branch '%s': %w", branchName, err)
}

// SwitchBranch switches to a branch.
func (c *GitClient) SwitchBranch(ctx context.Context, branchName string) error {
	err := c.runGit(ctx, "switch", branchName)
	if err != nil {
		return fmt.Errorf("git switch %s failed: %w", branchName, err)
	}

	return nil
}

// CreateAndSwitchBranch creates a new branch and switches to it.
func (c *GitClient) CreateAndSwitchBranch(
	ctx context.Context,
	newBranchName string,
	baseBranch string,
) error {
	args := []string{"switch", "-c", newBranchName}
	if baseBranch != "" {
		args = append(args, baseBranch)
	}

	err := c.runGit(ctx, args...)
	if err != nil {
		return fmt.Errorf("git switch -c %s failed: %w", newBranchName, err)
	}

	return nil
}

// PushAndSetUpstream pushes the branch and sets the upstream.
func (c *GitClient) PushAndSetUpstream(ctx context.Context, branchName string) error {
	remote := c.RemoteName()

	err := c.runGit(ctx, "push", "--set-upstream", remote, branchName)
	if err != nil {
		return fmt.Errorf("git push --set-upstream %s %s failed: %w", remote, branchName, err)
	}

	return nil
}

// ListTrackedAndCachedFiles returns a list of tracked and cached files.
func (c *GitClient) ListTrackedAndCachedFiles(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "ls-files", "-co", "--exclude-standard")
}

// GetLogAndDiffFromMergeBase finds the common ancestor with a branch and returns the log and diff since that point.
//
//nolint:nonamedreturns // Named returns are used for clarity in return signature.
func (c *GitClient) GetLogAndDiffFromMergeBase(
	ctx context.Context,
	baseBranchRef string,
) (log, diff string, err error) {
	// First, check if the remote branch even exists.
	_, _, err = c.captureGitOutput(ctx, "rev-parse", "--verify", baseBranchRef)
	if err != nil {
		// This is not a fatal error; it often means the branch hasn't been pushed.
		// Return a specific error that the caller can check for.
		//nolint:err113 // Dynamic error is appropriate here.
		return "", "", fmt.Errorf("remote branch '%s' not found", baseBranchRef)
	}

	mergeBaseBytes, _, err := c.captureGitOutput(ctx, "merge-base", baseBranchRef, "HEAD")
	if err != nil {
		return "", "", fmt.Errorf("git merge-base against '%s' failed: %w", baseBranchRef, err)
	}

	//nolint:unconvert // Conversion is necessary for TrimSpace.
	mergeBase := strings.TrimSpace(string(mergeBaseBytes))

	log, _, err = c.captureGitOutput(
		ctx,
		"log",
		"--pretty=format:%h %s (%an, %cr)",
		mergeBase+"..HEAD",
	)
	if err != nil {
		return "", "", fmt.Errorf("git log failed: %w", err)
	}

	diff, _, err = c.captureGitOutput(ctx, "diff", mergeBase+"..HEAD")
	if err != nil {
		return "", "", fmt.Errorf("git diff failed: %w", err)
	}

	return log, diff, nil
}

// StashPush saves the current state of the working directory and the index, but leaves the working directory clean.
func (c *GitClient) StashPush(ctx context.Context) error {
	// Using -u to include untracked files, which is generally desired for this workflow.
	err := c.runGit(ctx, "stash", "push", "-u")
	if err != nil {
		return fmt.Errorf("git stash push failed: %w", err)
	}

	return nil
}

// GetMergeBase finds the common ancestor between HEAD and the target branch.
func (c *GitClient) GetMergeBase(ctx context.Context, targetBranch string) (string, error) {
	out, _, err := c.captureGitOutput(ctx, "merge-base", "HEAD", targetBranch)
	if err != nil {
		return "", fmt.Errorf("failed to find merge base with %s: %w", targetBranch, err)
	}

	return strings.TrimSpace(out), nil
}

// ResetSoft moves the current HEAD to the target commit, leaving changes staged.
func (c *GitClient) ResetSoft(ctx context.Context, targetCommit string) error {
	err := c.runGit(ctx, "reset", "--soft", targetCommit)
	if err != nil {
		return fmt.Errorf("git reset --soft failed: %w", err)
	}

	return nil
}

// ForcePushLease performs a safe force push.
func (c *GitClient) ForcePushLease(ctx context.Context, branch string) error {
	remote := c.RemoteName()

	err := c.runGit(ctx, "push", "--force-with-lease", remote, branch)
	if err != nil {
		return fmt.Errorf("git push --force-with-lease failed: %w", err)
	}

	return nil
}

// GetCommitCount returns the number of commits between two references (e.g., "main..HEAD").
func (c *GitClient) GetCommitCount(ctx context.Context, rangeSpec string) (int, error) {
	out, _, err := c.captureGitOutput(ctx, "rev-list", "--count", rangeSpec)
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(out), "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count, nil
}

func (c *GitClient) runGit(ctx context.Context, args ...string) error {
	// Logger().Debug(...) is already part of executor.Execute
	//nolint:wrapcheck // Wrapping is handled by caller or executor.
	return c.executor.Execute(ctx, c.repoPath, c.config.GitExecutable, args...)
}

func (c *GitClient) captureGitOutput(ctx context.Context, args ...string) (string, string, error) {
	// Logger().Debug(...) is already part of executor.CaptureOutput
	//nolint:wrapcheck // Wrapping is handled by caller or executor.
	return c.executor.CaptureOutput(ctx, c.repoPath, c.config.GitExecutable, args...)
}
