// internal/git/client.go

package git

import (
	"context"
	"errors" // For errors.As
	"fmt"
	"log/slog"
	"os"      // For os.Getwd()
	"os/exec" // Required for ExitError type
	"path/filepath"
	"strings"
)

// GitClient provides methods for interacting with a Git repository.
// Instances should be created via NewClient.
type GitClient struct {
	repoPath string          // Absolute path to the repository's working directory root.
	gitDir   string          // Absolute path to the .git directory or file.
	config   GitClientConfig // Validated configuration used by this client.
	logger   *slog.Logger    // Logger instance for this client.
	executor executor        // Command executor used by this client.
}

// NewClient creates and initializes a new GitClient for the Git repository
// located at or containing the specified workDir.
// If workDir is an empty string, the current working directory (os.Getwd()) is used
// as the starting point to find the repository root.
// The provided context (ctx) is used for initial setup commands (e.g., 'git rev-parse')
// and can be used for cancellation.
//
// It returns an initialized GitClient or an error if setup fails. Errors can occur if:
// - The workDir (or current directory if workDir is empty) is not within a Git repository.
// - The configured 'git' executable (from GitClientConfig.GitExecutable or system PATH) is not found.
// - The GitClientConfig is invalid after defaults are applied.
// - Underlying 'git rev-parse' commands fail during setup.
func NewClient(ctx context.Context, workDir string, config GitClientConfig) (*GitClient, error) {
	validatedConfig, err := config.validateAndSetDefaults() // This method is in config.go
	if err != nil {
		return nil, fmt.Errorf("invalid GitClientConfig: %w", err)
	}
	logger := validatedConfig.Logger
	executor := validatedConfig.Executor

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
			logger.ErrorContext(ctx, "GitClient setup failed: getwd",
				slog.String("source_component", "GitClient.NewClient"),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}
	logger.DebugContext(ctx, "GitClient setup: using initial workdir",
		slog.String("source_component", "GitClient.NewClient"),
		slog.String("workdir", effectiveWorkDir))

	topLevelCmdArgs := []string{"rev-parse", "--show-toplevel"}
	logger.DebugContext(ctx, "GitClient setup: finding repo top-level",
		slog.String("source_component", "GitClient.NewClient"),
		slog.String("command", validatedConfig.GitExecutable),
		slog.Any("args", topLevelCmdArgs))
	topLevel, stderr, err := executor.CaptureOutput(ctx, effectiveWorkDir, validatedConfig.GitExecutable, topLevelCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --show-toplevel",
			slog.String("source_component", "GitClient.NewClient"),
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("initial_workdir", effectiveWorkDir),
		)
		return nil, fmt.Errorf("path '%s' is not within a Git repository (or git command failed)", effectiveWorkDir)
	}
	repoPath := strings.TrimSpace(topLevel)
	logger.DebugContext(ctx, "GitClient setup: found repo top-level",
		slog.String("source_component", "GitClient.NewClient"),
		slog.String("repo_path", repoPath))

	gitDirCmdArgs := []string{"rev-parse", "--git-dir"}
	logger.DebugContext(ctx, "GitClient setup: finding .git directory",
		slog.String("source_component", "GitClient.NewClient"),
		slog.String("command", validatedConfig.GitExecutable),
		slog.Any("args", gitDirCmdArgs))
	gitDirOutput, stderr, err := executor.CaptureOutput(ctx, repoPath, validatedConfig.GitExecutable, gitDirCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --git-dir",
			slog.String("source_component", "GitClient.NewClient"),
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("repo_path", repoPath),
		)
		return nil, fmt.Errorf("could not determine .git directory for repo at '%s': %w", repoPath, err)
	}
	gitDir := strings.TrimSpace(gitDirOutput)
	if !filepath.IsAbs(gitDir) { // Ensure the gitDir path is absolute for consistency
		gitDir = filepath.Join(repoPath, gitDir)
		logger.DebugContext(ctx, "GitClient setup: resolved relative .git dir",
			slog.String("source_component", "GitClient.NewClient"),
			slog.String("git_dir_abs", gitDir))
	}
	logger.DebugContext(ctx, "GitClient setup: found .git directory",
		slog.String("source_component", "GitClient.NewClient"),
		slog.String("git_dir", gitDir))

	client := &GitClient{
		repoPath: repoPath,
		gitDir:   gitDir,
		config:   validatedConfig,
		logger:   logger,
		executor: executor,
	}

	logger.InfoContext(ctx, "GitClient initialized successfully",
		slog.String("source_component", "GitClient.NewClient"),
		slog.String("repository_path", client.repoPath),
		slog.String("git_dir", client.gitDir),
	)
	return client, nil
}

// Path returns the root path of the repository the client is managing.
func (c *GitClient) Path() string { return c.repoPath }

// GitDir returns the path to the repository's .git directory/file.
func (c *GitClient) GitDir() string { return c.gitDir }

// MainBranchName returns the configured default main branch name.
func (c *GitClient) MainBranchName() string { return c.config.DefaultMainBranchName }

// RemoteName returns the configured default remote name.
func (c *GitClient) RemoteName() string { return c.config.DefaultRemoteName }

// Logger returns the underlying logger instance.
func (c *GitClient) Logger() *slog.Logger { return c.logger }

// --- Internal helpers ---
func (c *GitClient) runGit(ctx context.Context, args ...string) error {
	c.logger.DebugContext(ctx, "Executing git command", slog.String("source_component", "GitClient.runGit_internal"), slog.String("repository_path", c.repoPath), slog.String("command", c.config.GitExecutable), slog.Any("args", args))
	err := c.executor.Execute(ctx, c.repoPath, c.config.GitExecutable, args...)
	if err != nil {
		c.logger.DebugContext(ctx, "Git command execution failed (internal)", slog.String("source_component", "GitClient.runGit_internal"), slog.String("error", err.Error()), slog.Any("args", args))
	}
	return err
}
func (c *GitClient) captureGitOutput(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	c.logger.DebugContext(ctx, "Capturing git command output", slog.String("source_component", "GitClient.captureGitOutput_internal"), slog.String("repository_path", c.repoPath), slog.String("command", c.config.GitExecutable), slog.Any("args", args))
	stdout, stderr, err = c.executor.CaptureOutput(ctx, c.repoPath, c.config.GitExecutable, args...)
	if err != nil {
		c.logger.DebugContext(ctx, "Git command capture finished with error (internal)", slog.String("source_component", "GitClient.captureGitOutput_internal"), slog.String("error", err.Error()), slog.String("stderr_snippet", TruncateString(stderr, 100)), slog.Any("args", args))
	}
	return stdout, stderr, err
}

// --- Public Git Operation Methods ---

func (c *GitClient) GetCurrentBranchName(ctx context.Context) (string, error) {
	const sourceComponent = "GitClient.GetCurrentBranchName"
	c.logger.DebugContext(ctx, "Attempting to get current branch name", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	stdout, stderr, err := c.captureGitOutput(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to get current branch name", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath), slog.String("error", err.Error()), slog.String("stderr", strings.TrimSpace(stderr)))
		return "", fmt.Errorf("failed to get current branch name: %w", err)
	}
	branch := strings.TrimSpace(stdout)
	if branch == "HEAD" {
		warnMsg := "currently in detached HEAD state"
		c.logger.InfoContext(ctx, warnMsg, slog.String("source_component", sourceComponent))
		return "", fmt.Errorf("%s", warnMsg)
	}
	if branch == "" {
		errMsg := "could not determine current branch name (empty output)"
		c.logger.ErrorContext(ctx, errMsg, slog.String("source_component", sourceComponent))
		return "", fmt.Errorf("%s", errMsg)
	}
	c.logger.DebugContext(ctx, "Successfully got current branch name", slog.String("source_component", sourceComponent), slog.String("branch", branch))
	return branch, nil
}

func (c *GitClient) AddAll(ctx context.Context) error {
	const sourceComponent = "GitClient.AddAll"
	c.logger.InfoContext(ctx, "Attempting to stage all changes ('git add .')", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	err := c.runGit(ctx, "add", ".")
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to stage all changes ('git add .')", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath), slog.String("error", err.Error()))
		return fmt.Errorf("git add . failed: %w", err)
	}
	c.logger.InfoContext(ctx, "'git add .' executed successfully.", slog.String("source_component", sourceComponent))
	return nil
}

func (c *GitClient) Commit(ctx context.Context, message string) error {
	const sourceComponent = "GitClient.Commit"
	c.logger.InfoContext(ctx, "Attempting git commit", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath), slog.String("message_preview", TruncateString(message, 50)))
	if strings.TrimSpace(message) == "" {
		err := fmt.Errorf("commit message cannot be empty")
		c.logger.ErrorContext(ctx, "Commit validation failed: empty message", slog.String("source_component", sourceComponent))
		return err
	}
	err := c.runGit(ctx, "commit", "-m", message)
	if err != nil {
		c.logger.ErrorContext(ctx, "Git commit command failed", slog.String("source_component", sourceComponent), slog.String("error", err.Error()))
		return fmt.Errorf("commit command failed: %w", err)
	}
	c.logger.InfoContext(ctx, "Git commit successful", slog.String("source_component", sourceComponent), slog.String("message", message))
	return nil
}

func (c *GitClient) HasStagedChanges(ctx context.Context) (bool, error) {
	const sourceComponent = "GitClient.HasStagedChanges"
	c.logger.DebugContext(ctx, "Checking for staged changes ('git diff --quiet --cached')", slog.String("source_component", sourceComponent))
	_, stderr, err := c.captureGitOutput(ctx, "diff", "--quiet", "--cached")
	if err == nil {
		c.logger.DebugContext(ctx, "No staged changes detected (exit 0).", slog.String("source_component", sourceComponent))
		return false, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == 1 {
			c.logger.DebugContext(ctx, "Staged changes detected (exit 1).", slog.String("source_component", sourceComponent))
			return true, nil
		}
		c.logger.ErrorContext(ctx, "Unexpected exit code from 'git diff --quiet --cached'", slog.String("source_component", sourceComponent), slog.Int("exit_code", exitErr.ExitCode()), slog.String("stderr", strings.TrimSpace(stderr)), slog.String("error_detail", err.Error()))
		return false, fmt.Errorf("failed to determine staged status, git command failed with exit code %d: %w", exitErr.ExitCode(), err)
	}
	c.logger.ErrorContext(ctx, "Error executing 'git diff --quiet --cached'", slog.String("source_component", sourceComponent), slog.String("stderr", strings.TrimSpace(stderr)), slog.String("error_detail", err.Error()))
	return false, fmt.Errorf("failed to check staged status: %w", err)
}

func (c *GitClient) GetStatus(ctx context.Context) (stdout, stderr string, err error) {
	const sourceComponent = "GitClient.GetStatus"
	c.logger.DebugContext(ctx, "Attempting to get git status", slog.String("source_component", sourceComponent))
	stdout, stderr, err = c.captureGitOutput(ctx, "status")
	if err != nil {
		c.logger.WarnContext(ctx, "Git status command finished with error or non-zero exit", slog.String("source_component", sourceComponent), slog.String("error_detail", err.Error()), slog.String("stdout_capture", TruncateString(stdout, 200)), slog.String("stderr_capture", TruncateString(stderr, 200)))
	} else {
		c.logger.DebugContext(ctx, "Git status command executed successfully (exit 0)", slog.String("source_component", sourceComponent))
	}
	return stdout, stderr, err
}

// GetStatusShort retrieves the output of `git status --short`.
// This provides a concise summary of changed files (staged, unstaged, untracked).
// Returns stdout, stderr, and any execution error.
func (c *GitClient) GetStatusShort(ctx context.Context) (stdout, stderr string, err error) {
	const sourceComponent = "GitClient.GetStatusShort"
	c.logger.DebugContext(ctx, "Attempting to get git status --short", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	stdout, stderr, err = c.captureGitOutput(ctx, "status", "--short")
	if err != nil {
		c.logger.WarnContext(ctx, "Git status --short command finished with error or non-zero exit", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath), slog.String("error_detail", err.Error()), slog.String("stdout_capture", TruncateString(stdout, 200)), slog.String("stderr_capture", TruncateString(stderr, 200)))
	} else {
		c.logger.DebugContext(ctx, "Git status --short command executed successfully", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	}
	return stdout, stderr, err
}

// GetDiffCached retrieves the diff between the Git index (staged changes) and the HEAD commit.
// Returns stdout, stderr, and error. Exit code 1 from `git diff` indicates differences were found and is not treated as an execution error by this function itself.
func (c *GitClient) GetDiffCached(ctx context.Context) (stdout, stderr string, err error) {
	const sourceComponent = "GitClient.GetDiffCached"
	c.logger.DebugContext(ctx, "Attempting 'git diff --cached'", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	stdout, stderr, err = c.captureGitOutput(ctx, "diff", "--cached")
	if err != nil {
		var exitErr *exec.ExitError
		isExitCodeOne := errors.As(err, &exitErr) && exitErr.ExitCode() == 1
		if !isExitCodeOne {
			c.logger.WarnContext(ctx, "'git diff --cached' command finished with unexpected error", slog.String("source_component", sourceComponent), slog.String("error_detail", err.Error()), slog.String("stderr_capture", TruncateString(stderr, 200)))
		} else {
			c.logger.DebugContext(ctx, "'git diff --cached' found changes (exit 1).", slog.String("source_component", sourceComponent))
		}
		if !isExitCodeOne {
			return stdout, stderr, err
		} // Return actual error
		return stdout, stderr, nil // Treat exit code 1 as success for this method's contract
	}
	c.logger.DebugContext(ctx, "'git diff --cached' found no changes (exit 0).", slog.String("source_component", sourceComponent))
	return stdout, stderr, nil // No error, no diff
}

// GetDiffUnstaged retrieves the diff between the working directory and the Git index (unstaged changes).
// Returns stdout, stderr, and error. Exit code 1 from `git diff` indicates differences were found and is not treated as an execution error by this function itself.
func (c *GitClient) GetDiffUnstaged(ctx context.Context) (stdout, stderr string, err error) {
	const sourceComponent = "GitClient.GetDiffUnstaged"
	c.logger.DebugContext(ctx, "Attempting 'git diff HEAD'", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	stdout, stderr, err = c.captureGitOutput(ctx, "diff", "HEAD")
	if err != nil {
		var exitErr *exec.ExitError
		isExitCodeOne := errors.As(err, &exitErr) && exitErr.ExitCode() == 1
		if !isExitCodeOne {
			c.logger.WarnContext(ctx, "'git diff HEAD' command finished with unexpected error", slog.String("source_component", sourceComponent), slog.String("error_detail", err.Error()), slog.String("stderr_capture", TruncateString(stderr, 200)))
		} else {
			c.logger.DebugContext(ctx, "'git diff HEAD' found changes (exit 1).", slog.String("source_component", sourceComponent))
		}
		if !isExitCodeOne {
			return stdout, stderr, err
		} // Return actual error
		return stdout, stderr, nil // Treat exit code 1 as success
	}
	c.logger.DebugContext(ctx, "'git diff HEAD' found no changes (exit 0).", slog.String("source_component", sourceComponent))
	return stdout, stderr, nil // No error, no diff
}

// ListUntrackedFiles retrieves a list of untracked files in the repository, respecting .gitignore.
// Returns the list as a string (each file on a newline), stderr output, and any error.
func (c *GitClient) ListUntrackedFiles(ctx context.Context) (stdout, stderr string, err error) {
	const sourceComponent = "GitClient.ListUntrackedFiles"
	c.logger.DebugContext(ctx, "Attempting 'git ls-files --others --exclude-standard'", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath))
	stdout, stderr, err = c.captureGitOutput(ctx, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		c.logger.ErrorContext(ctx, "'git ls-files --others --exclude-standard' command failed", slog.String("source_component", sourceComponent), slog.String("repository_path", c.repoPath), slog.String("error_detail", err.Error()), slog.String("stderr_capture", strings.TrimSpace(stderr)))
		return stdout, stderr, fmt.Errorf("failed to list untracked files: %w", err) // Wrap error
	}
	if strings.TrimSpace(stdout) != "" {
		c.logger.DebugContext(ctx, "Untracked files found.", slog.String("source_component", sourceComponent))
	} else {
		c.logger.DebugContext(ctx, "No untracked files found.", slog.String("source_component", sourceComponent))
	}
	return stdout, stderr, nil
}

// TruncateString helper
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

// internal/git/client.go

// ... (Existing imports, struct, NewClient, other methods) ...

// IsWorkingDirClean checks if the working directory has no staged, unstaged,
// or untracked changes. Returns true if clean, false otherwise.
// Returns an error if any underlying git command fails unexpectedly.
func (c *GitClient) IsWorkingDirClean(ctx context.Context) (bool, error) {
	const sourceComponent = "GitClient.IsWorkingDirClean"
	c.logger.DebugContext(ctx, "Checking working directory cleanliness",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
	)

	// Check unstaged (working tree vs index) using diff --quiet
	_, _, errDiff := c.captureGitOutput(ctx, "diff", "--quiet")
	if errDiff != nil {
		var exitErr *exec.ExitError
		if errors.As(errDiff, &exitErr) && exitErr.ExitCode() == 1 {
			c.logger.DebugContext(ctx, "Working directory not clean: unstaged changes found.", slog.String("source_component", sourceComponent))
			return false, nil // Differences found (exit 1) is not an error here
		}
		// Unexpected error
		c.logger.ErrorContext(ctx, "Failed checking unstaged changes", slog.String("source_component", sourceComponent), slog.String("error", errDiff.Error()))
		return false, fmt.Errorf("failed checking unstaged changes: %w", errDiff)
	}

	// Check staged (index vs HEAD) using diff --cached --quiet
	hasStaged, errStaged := c.HasStagedChanges(ctx) // Reuse existing method
	if errStaged != nil {
		// Unexpected error from HasStagedChanges
		c.logger.ErrorContext(ctx, "Failed checking staged changes", slog.String("source_component", sourceComponent), slog.String("error", errStaged.Error()))
		return false, fmt.Errorf("failed checking staged changes: %w", errStaged)
	}
	if hasStaged {
		c.logger.DebugContext(ctx, "Working directory not clean: staged changes found.", slog.String("source_component", sourceComponent))
		return false, nil // Staged changes found
	}

	// Check untracked using ls-files
	untrackedOut, _, errUntracked := c.ListUntrackedFiles(ctx) // Reuse existing method
	if errUntracked != nil {
		// Unexpected error from ListUntrackedFiles
		c.logger.ErrorContext(ctx, "Failed checking untracked files", slog.String("source_component", sourceComponent), slog.String("error", errUntracked.Error()))
		return false, fmt.Errorf("failed checking untracked files: %w", errUntracked)
	}
	if strings.TrimSpace(untrackedOut) != "" {
		c.logger.DebugContext(ctx, "Working directory not clean: untracked files found.", slog.String("source_component", sourceComponent))
		return false, nil // Untracked files found
	}

	c.logger.DebugContext(ctx, "Working directory is clean.", slog.String("source_component", sourceComponent))
	return true, nil
}

// PullRebase executes `git pull --rebase [remote] [branch]` using the client's
// configured default remote and the specified branch.
// Returns an error if the command fails (e.g., conflicts, network issues).
func (c *GitClient) PullRebase(ctx context.Context, branch string) error {
	const sourceComponent = "GitClient.PullRebase"
	remote := c.RemoteName() // Use configured remote
	c.logger.InfoContext(ctx, "Attempting 'git pull --rebase'",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("remote", remote),
		slog.String("branch", branch),
	)
	// Pull rebase can be interactive on conflict, use runGit which pipes stdio
	err := c.runGit(ctx, "pull", "--rebase", remote, branch)
	if err != nil {
		c.logger.ErrorContext(ctx, "'git pull --rebase' command failed",
			slog.String("source_component", sourceComponent),
			slog.String("repository_path", c.repoPath),
			slog.String("remote", remote),
			slog.String("branch", branch),
			slog.String("error", err.Error()), // Error likely includes exit code/reason
		)
		// Provide a slightly more specific error message
		return fmt.Errorf("git pull --rebase %s %s failed: %w", remote, branch, err)
	}
	c.logger.InfoContext(ctx, "'git pull --rebase' successful",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("remote", remote),
		slog.String("branch", branch),
	)
	return nil
}

// IsBranchAhead checks if the current local branch is ahead of its upstream remote.
// It parses the output of `git status -sb`.
// Returns true if ahead, false otherwise. Returns an error if status cannot be checked.
func (c *GitClient) IsBranchAhead(ctx context.Context) (bool, error) {
	const sourceComponent = "GitClient.IsBranchAhead"
	c.logger.DebugContext(ctx, "Checking if branch is ahead using 'git status -sb'",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
	)
	// Use -sb for concise, parseable output
	// Note: GetStatusShort uses --short, which is different. We need -sb here.
	stdout, stderr, err := c.captureGitOutput(ctx, "status", "-sb")
	if err != nil {
		// Status command failed unexpectedly
		c.logger.ErrorContext(ctx, "Failed to execute 'git status -sb'",
			slog.String("source_component", sourceComponent),
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
		)
		return false, fmt.Errorf("failed to get status to check if branch is ahead: %w", err)
	}

	// Expected output format for ahead: "## branch...remote/branch [ahead N]"
	// We only need to check for the "[ahead " substring in the output.
	isAhead := strings.Contains(stdout, "[ahead ")
	c.logger.DebugContext(ctx, "Branch ahead status determined",
		slog.String("source_component", sourceComponent),
		slog.Bool("is_ahead", isAhead),
		slog.String("status_output_first_line", strings.SplitN(stdout, "\n", 2)[0]), // Log first line for context
	)
	return isAhead, nil
}

// Push executes `git push [remote] [branch]` using the client's configured
// default remote and the specified *local* branch name to push.
// If branch is empty, attempts to push the current branch based on Git's default behavior.
// Returns an error if the push fails. Handles "Everything up-to-date" gracefully.
func (c *GitClient) Push(ctx context.Context, branch string) error {
	const sourceComponent = "GitClient.Push"
	remote := c.RemoteName()
	args := []string{"push", remote}
	if branch != "" {
		args = append(args, branch)
	}
	c.logger.InfoContext(ctx, "Attempting 'git push'",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("remote", remote),
		slog.String("branch_arg", branch), // Log the branch arg passed
	)
	// Push might interact (e.g., credentials), use runGit
	err := c.runGit(ctx, args...)
	if err != nil {
		// Check if the error is simply "up-to-date" which isn't really a failure
		// Need to capture stderr for this, so maybe captureGitOutput is better?
		// Let's switch to capture to check stderr.

		// Re-run with capture (or modify runGit to capture stderr on error?)
		// For simplicity here, let's assume runGit's error string might contain it.
		// A more robust solution captures stderr separately.
		errMsg := err.Error() // Check the error message directly
		if strings.Contains(errMsg, "Everything up-to-date") || strings.Contains(errMsg, "already up-to-date") {
			c.logger.InfoContext(ctx, "'git push' reported everything up-to-date.",
				slog.String("source_component", sourceComponent),
				slog.String("repository_path", c.repoPath),
				slog.String("remote", remote),
				slog.String("branch_arg", branch),
			)
			return nil // Not a failure
		}

		// Real error
		c.logger.ErrorContext(ctx, "'git push' command failed",
			slog.String("source_component", sourceComponent),
			slog.String("repository_path", c.repoPath),
			slog.String("remote", remote),
			slog.String("branch_arg", branch),
			slog.String("error", errMsg),
		)
		return fmt.Errorf("git push failed: %w", err)
	}
	c.logger.InfoContext(ctx, "'git push' successful",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("remote", remote),
		slog.String("branch_arg", branch),
	)
	return nil
}

// internal/git/client.go

// ... (Existing imports, struct, NewClient, other methods) ...

// LocalBranchExists checks if a local branch with the given name exists.
func (c *GitClient) LocalBranchExists(ctx context.Context, branchName string) (bool, error) {
	const sourceComponent = "GitClient.LocalBranchExists"
	c.logger.DebugContext(ctx, "Checking local branch existence",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("branch", branchName),
	)
	// 'git show-ref --verify --quiet refs/heads/<branch>' exits 0 if exists, 1 if not, >1 on error
	ref := fmt.Sprintf("refs/heads/%s", branchName)
	_, stderr, err := c.captureGitOutput(ctx, "show-ref", "--verify", "--quiet", ref)

	if err == nil {
		c.logger.DebugContext(ctx, "Local branch exists.", slog.String("source_component", sourceComponent), slog.String("branch", branchName))
		return true, nil // Exit 0 means it exists
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		c.logger.DebugContext(ctx, "Local branch does not exist.", slog.String("source_component", sourceComponent), slog.String("branch", branchName))
		return false, nil // Exit 1 means it doesn't exist (not an error for this check)
	}

	// Any other error
	c.logger.ErrorContext(ctx, "Failed to check local branch existence",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("branch", branchName),
		slog.String("error", err.Error()),
		slog.String("stderr", strings.TrimSpace(stderr)),
	)
	return false, fmt.Errorf("failed to check existence of local branch '%s': %w", branchName, err)
}

// SwitchBranch switches the working directory to the specified existing local branch.
func (c *GitClient) SwitchBranch(ctx context.Context, branchName string) error {
	const sourceComponent = "GitClient.SwitchBranch"
	c.logger.InfoContext(ctx, "Attempting to switch to branch",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("branch", branchName),
	)
	// Use runGit as switch might print info/errors directly
	err := c.runGit(ctx, "switch", branchName)
	if err != nil {
		c.logger.ErrorContext(ctx, "'git switch' command failed",
			slog.String("source_component", sourceComponent),
			slog.String("branch", branchName),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("git switch %s failed: %w", branchName, err)
	}
	c.logger.InfoContext(ctx, "Successfully switched to branch",
		slog.String("source_component", sourceComponent),
		slog.String("branch", branchName),
	)
	return nil
}

// CreateAndSwitchBranch creates a new local branch from the specified base branch
// (or current HEAD if baseBranch is empty) and switches to it. Equivalent to `git switch -c <newBranch> [baseBranch]`.
func (c *GitClient) CreateAndSwitchBranch(ctx context.Context, newBranchName string, baseBranch string) error {
	const sourceComponent = "GitClient.CreateAndSwitchBranch"
	args := []string{"switch", "-c", newBranchName}
	if baseBranch != "" {
		args = append(args, baseBranch)
	}
	c.logger.InfoContext(ctx, "Attempting to create and switch to new branch",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("new_branch", newBranchName),
		slog.String("base_branch", baseBranch), // Will be empty if using current HEAD
	)
	// Use runGit as switch might print info/errors directly
	err := c.runGit(ctx, args...)
	if err != nil {
		c.logger.ErrorContext(ctx, "'git switch -c' command failed",
			slog.String("source_component", sourceComponent),
			slog.String("new_branch", newBranchName),
			slog.String("base_branch", baseBranch),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("git switch -c %s failed: %w", newBranchName, err)
	}
	c.logger.InfoContext(ctx, "Successfully created and switched to new branch",
		slog.String("source_component", sourceComponent),
		slog.String("new_branch", newBranchName),
		slog.String("base_branch", baseBranch),
	)
	return nil
}

// PushAndSetUpstream pushes the specified local branch to the default remote
// and sets the upstream tracking configuration. Equivalent to `git push --set-upstream <remote> <branch>`.
func (c *GitClient) PushAndSetUpstream(ctx context.Context, branchName string) error {
	const sourceComponent = "GitClient.PushAndSetUpstream"
	remote := c.RemoteName()
	c.logger.InfoContext(ctx, "Attempting 'git push --set-upstream'",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.String("remote", remote),
		slog.String("branch", branchName),
	)
	// Use runGit as push might print info/errors directly
	err := c.runGit(ctx, "push", "--set-upstream", remote, branchName)
	if err != nil {
		c.logger.ErrorContext(ctx, "'git push --set-upstream' command failed",
			slog.String("source_component", sourceComponent),
			slog.String("remote", remote),
			slog.String("branch", branchName),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("git push --set-upstream %s %s failed: %w", remote, branchName, err)
	}
	c.logger.InfoContext(ctx, "'git push --set-upstream' successful",
		slog.String("source_component", sourceComponent),
		slog.String("remote", remote),
		slog.String("branch", branchName),
	)
	return nil
}

// ListTrackedAndCachedFiles retrieves a list of files known to Git (tracked or staged/cached),
// respecting .gitignore rules handled by git itself. Includes 'other' (untracked) files too,
// basically mirroring `git ls-files -co --exclude-standard`.
// Returns the list as a single string (each file on a newline), stderr output, and any error.
func (c *GitClient) ListTrackedAndCachedFiles(ctx context.Context) (stdout, stderr string, err error) {
	const sourceComponent = "GitClient.ListTrackedCached"     // Renamed component slightly
	args := []string{"ls-files", "-co", "--exclude-standard"} // Cached, Others, Exclude standard ignores
	c.logger.DebugContext(ctx, "Listing tracked, cached, and other files",
		slog.String("source_component", sourceComponent),
		slog.String("repository_path", c.repoPath),
		slog.Any("args", args),
	)
	// `ls-files` should exit 0 unless there's a real error.
	stdout, stderr, err = c.captureGitOutput(ctx, args...)
	if err != nil {
		c.logger.ErrorContext(ctx, "'git ls-files -co --exclude-standard' command failed",
			slog.String("source_component", sourceComponent),
			slog.String("repository_path", c.repoPath),
			slog.String("error_detail", err.Error()),
			slog.String("stderr_capture", strings.TrimSpace(stderr)),
		)
		return stdout, stderr, fmt.Errorf("failed to list files known to git: %w", err) // Wrap error
	}
	c.logger.DebugContext(ctx, "'git ls-files -co --exclude-standard' successful.",
		slog.String("source_component", sourceComponent),
		// Optionally log file count if needed, similar to ListUntrackedFiles
		// slog.Int("file_count", strings.Count(strings.TrimSpace(stdout), "\n")+1),
	)
	return stdout, stderr, nil
}
