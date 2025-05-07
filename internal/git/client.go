// internal/git/client.go
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
		err := fmt.Errorf("git executable '%s' not found in PATH or specified path", validatedConfig.GitExecutable)
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
			logger.ErrorContext(ctx, "GitClient setup failed: getwd", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Use the executor to find repo top-level
	topLevelCmdArgs := []string{"rev-parse", "--show-toplevel"}
	topLevel, stderr, err := executor.CaptureOutput(ctx, effectiveWorkDir, validatedConfig.GitExecutable, topLevelCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --show-toplevel",
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("initial_workdir", effectiveWorkDir))
		return nil, fmt.Errorf("path '%s' is not within a Git repository (or git command '%s' failed)", effectiveWorkDir, validatedConfig.GitExecutable)
	}
	repoPath := strings.TrimSpace(topLevel)

	// Use the executor to find .git directory
	gitDirCmdArgs := []string{"rev-parse", "--git-dir"}
	gitDirOutput, stderr, err := executor.CaptureOutput(ctx, repoPath, validatedConfig.GitExecutable, gitDirCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --git-dir",
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("repo_path", repoPath))
		return nil, fmt.Errorf("could not determine .git directory for repo at '%s': %w", repoPath, err)
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

func (c *GitClient) Path() string           { return c.repoPath }
func (c *GitClient) GitDir() string         { return c.gitDir }
func (c *GitClient) MainBranchName() string { return c.config.DefaultMainBranchName }
func (c *GitClient) RemoteName() string     { return c.config.DefaultRemoteName }
func (c *GitClient) Logger() *slog.Logger   { return c.logger }

func (c *GitClient) runGit(ctx context.Context, args ...string) error {
	// Logger().Debug(...) is already part of executor.Execute
	return c.executor.Execute(ctx, c.repoPath, c.config.GitExecutable, args...)
}

func (c *GitClient) captureGitOutput(ctx context.Context, args ...string) (string, string, error) {
	// Logger().Debug(...) is already part of executor.CaptureOutput
	return c.executor.CaptureOutput(ctx, c.repoPath, c.config.GitExecutable, args...)
}

// --- Public Git Operation Methods ---
// (These methods remain largely the same but now internally call c.executor methods
//  which are of type exec.CommandExecutor)

func (c *GitClient) GetCurrentBranchName(ctx context.Context) (string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to get current branch name", "error", err, "stderr", strings.TrimSpace(stderr))
		return "", fmt.Errorf("failed to get current branch name: %w", err)
	}
	branch := strings.TrimSpace(stdout)
	if branch == "HEAD" {
		return "", fmt.Errorf("currently in detached HEAD state")
	}
	if branch == "" {
		return "", fmt.Errorf("could not determine current branch name (empty output)")
	}
	return branch, nil
}

func (c *GitClient) AddAll(ctx context.Context) error {
	err := c.runGit(ctx, "add", ".")
	if err != nil {
		return fmt.Errorf("git add . failed: %w", err)
	}
	return nil
}

func (c *GitClient) Commit(ctx context.Context, message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}
	err := c.runGit(ctx, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("commit command failed: %w", err)
	}
	return nil
}

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

func (c *GitClient) GetStatusShort(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "status", "--short")
}

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

func (c *GitClient) ListUntrackedFiles(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "ls-files", "--others", "--exclude-standard")
}

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

func (c *GitClient) PullRebase(ctx context.Context, branch string) error {
	remote := c.RemoteName()
	err := c.runGit(ctx, "pull", "--rebase", remote, branch)
	if err != nil {
		return fmt.Errorf("git pull --rebase %s %s failed: %w", remote, branch, err)
	}
	return nil
}

func (c *GitClient) IsBranchAhead(ctx context.Context) (bool, error) {
	stdout, _, err := c.captureGitOutput(ctx, "status", "-sb")
	if err != nil {
		return false, fmt.Errorf("failed to get status to check if branch is ahead: %w", err)
	}
	return strings.Contains(stdout, "[ahead "), nil
}

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
		if strings.Contains(errMsg, "everything up-to-date") || strings.Contains(errMsg, "already up-to-date") {
			c.logger.InfoContext(ctx, "'git push' reported everything up-to-date.", "remote", remote, "branch_arg", branch)
			return nil // Not a failure
		}
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}

func (c *GitClient) LocalBranchExists(ctx context.Context, branchName string) (bool, error) {
	ref := fmt.Sprintf("refs/heads/%s", branchName)
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

func (c *GitClient) SwitchBranch(ctx context.Context, branchName string) error {
	err := c.runGit(ctx, "switch", branchName)
	if err != nil {
		return fmt.Errorf("git switch %s failed: %w", branchName, err)
	}
	return nil
}

func (c *GitClient) CreateAndSwitchBranch(ctx context.Context, newBranchName string, baseBranch string) error {
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

func (c *GitClient) PushAndSetUpstream(ctx context.Context, branchName string) error {
	remote := c.RemoteName()
	err := c.runGit(ctx, "push", "--set-upstream", remote, branchName)
	if err != nil {
		return fmt.Errorf("git push --set-upstream %s %s failed: %w", remote, branchName, err)
	}
	return nil
}

func (c *GitClient) ListTrackedAndCachedFiles(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "ls-files", "-co", "--exclude-standard")
}

// TruncateString helper can be removed if not used, or kept if useful elsewhere.
// For now, keeping it as it was in the provided file.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		if maxLen < 0 {
			return ""
		}
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
