package tools

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// === Git Specific Helpers ===
// Note: CommandExists, ExecuteCommand, CaptureCommandOutput are defined in exec.go in the same package 'tools'

// IsGitRepo checks if a directory contains a .git subdirectory or file (for worktrees).
func IsGitRepo(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}

// CheckGitPrereqs verifies 'git' command exists and current directory is a Git repo.
func CheckGitPrereqs() (string, error) {
	// Uses CommandExists from exec.go
	if !CommandExists("git") {
		return "", fmt.Errorf("'git' command not found in PATH, please ensure Git is installed")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	// Uses IsGitRepo from this file
	if !IsGitRepo(cwd) {
		return "", fmt.Errorf("current directory '%s' is not a Git repository", cwd)
	}
	return cwd, nil
}

// ExecuteGitCommand runs a git command, piping stdio.
func ExecuteGitCommand(cwd string, args ...string) error {
	// Uses ExecuteCommand from exec.go
	return ExecuteCommand(cwd, "git", args...)
}

// CaptureGitOutput runs a git command and captures its stdout and stderr.
// Returns stdout string, stderr string, and any error.
func CaptureGitOutput(cwd string, args ...string) (string, string, error) {
	// Uses CaptureCommandOutput from exec.go
	return CaptureCommandOutput(cwd, "git", args...)
}

// IsWorkingDirClean checks for staged, unstaged, or untracked files.
func IsWorkingDirClean(cwd string) (bool, error) {
	// Check for unstaged changes (working tree vs index)
	_, _, errUnstaged := CaptureGitOutput(cwd, "diff", "--quiet")
	if errUnstaged != nil {
		var exitErr *exec.ExitError
		if errors.As(errUnstaged, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil // Differences found (unstaged)
		}
		// Error wrapping already includes stderr context from CaptureGitOutput
		return false, fmt.Errorf("failed to check unstaged changes: %w", errUnstaged)
	}

	// Check for staged changes (index vs HEAD)
	_, _, errStaged := CaptureGitOutput(cwd, "diff", "--cached", "--quiet")
	if errStaged != nil {
		var exitErr *exec.ExitError
		if errors.As(errStaged, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil // Differences found (staged)
		}
		return false, fmt.Errorf("failed to check staged changes: %w", errStaged)
	}

	// Check for untracked files
	stdoutUntracked, _, errUntracked := CaptureGitOutput(cwd, "ls-files", "--others", "--exclude-standard")
	if errUntracked != nil {
		return false, fmt.Errorf("failed to check for untracked files: %w", errUntracked)
	}
	if len(strings.TrimSpace(stdoutUntracked)) > 0 {
		return false, nil // Untracked files found
	}

	return true, nil
}

// GetCurrentBranchName retrieves the current Git branch name.
// Returns an error if in a detached HEAD state or if the branch name cannot be determined.
func GetCurrentBranchName(cwd string) (string, error) {
	// Correctly captures 3 values now
	stdout, _, err := CaptureGitOutput(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		// Error message already includes stderr context from CaptureGitOutput's wrapping
		return "", fmt.Errorf("failed to get current branch name: %w", err)
	}
	branchName := strings.TrimSpace(stdout)
	if branchName == "HEAD" {
		return "", fmt.Errorf("currently in a detached HEAD state, not on a branch")
	}
	if branchName == "" {
		return "", fmt.Errorf("could not determine current branch name (empty output)")
	}
	return branchName, nil
}

// LocalBranchExists checks if a local branch exists by its short name.
func LocalBranchExists(cwd, branchName string) (bool, error) {
	// Using direct exec.Command for precise exit code check remains appropriate here.
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	cmd.Dir = cwd
	err := cmd.Run()

	if err == nil {
		return true, nil // Exit 0 means branch exists
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == 1 {
			return false, nil // Exit 1 specifically means ref does not exist
		}
		return false, fmt.Errorf("git show-ref for branch '%s' failed with exit code %d: %w", branchName, exitErr.ExitCode(), err)
	}
	return false, fmt.Errorf("failed to execute git show-ref for branch '%s': %w", branchName, err)
}

// BranchSyncStatus represents the synchronization state of a local branch relative to its remote.
type BranchSyncStatus int

const (
	StatusUpToDate BranchSyncStatus = iota
	StatusAhead
	StatusBehind
	StatusDiverged
	StatusNoUpstreamOrRemoteMissing
	StatusError
)

// String provides a human-readable representation of BranchSyncStatus.
func (s BranchSyncStatus) String() string {
	switch s {
	case StatusUpToDate:
		return "UpToDate"
	case StatusAhead:
		return "Ahead"
	case StatusBehind:
		return "Behind"
	case StatusDiverged:
		return "Diverged"
	case StatusNoUpstreamOrRemoteMissing:
		return "NoUpstreamOrRemoteMissing"
	case StatusError:
		return "Error"
	default:
		return "UnknownStatus"
	}
}

// GetLocalBranchSyncStatus compares a local branch with its specified remote-tracking branch.
// Assumes 'git fetch' has been run recently enough for the remote-tracking ref to be locally available if it exists remotely.
func GetLocalBranchSyncStatus(cwd, localBranchName, remoteName, remoteBranchName string) (status BranchSyncStatus, aheadCount int, behindCount int, err error) {
	if localBranchName == "" || remoteName == "" || remoteBranchName == "" {
		return StatusError, 0, 0, fmt.Errorf("localBranchName, remoteName, and remoteBranchName must all be provided")
	}

	localRef := fmt.Sprintf("refs/heads/%s", localBranchName)
	remoteTrackingRef := fmt.Sprintf("refs/remotes/%s/%s", remoteName, remoteBranchName)

	// Verify local branch ref exists
	_, _, errLocal := CaptureGitOutput(cwd, "rev-parse", "--verify", "--quiet", localRef)
	if errLocal != nil {
		return StatusError, 0, 0, fmt.Errorf("local branch '%s' not found or invalid: %w", localBranchName, errLocal)
	}

	// Verify remote-tracking branch ref exists locally
	_, _, errRemote := CaptureGitOutput(cwd, "rev-parse", "--verify", "--quiet", remoteTrackingRef)
	if errRemote != nil {
		// Indicate the ref is missing locally, which implies no upstream or fetch needed/failed
		return StatusNoUpstreamOrRemoteMissing, 0, 0, fmt.Errorf("remote-tracking branch '%s/%s' (ref: %s) not found locally. Ensure upstream is configured, branch exists on remote '%s', and fetch was successful: %w", remoteName, remoteBranchName, remoteTrackingRef, remoteName, errRemote)
	}

	// Get ahead/behind counts
	revListRange := fmt.Sprintf("%s...%s", localRef, remoteTrackingRef)
	countsStdout, _, errCounts := CaptureGitOutput(cwd, "rev-list", "--left-right", "--count", revListRange)
	if errCounts != nil {
		return StatusError, 0, 0, fmt.Errorf("failed to get commit counts between '%s' and '%s': %w", localRef, remoteTrackingRef, errCounts)
	}

	countsParts := strings.Split(strings.TrimSpace(countsStdout), "\t") // Use tab as delimiter
	if len(countsParts) != 2 {
		return StatusError, 0, 0, fmt.Errorf("unexpected output format from 'git rev-list --left-right --count': '%s'", countsStdout)
	}

	ahead, errAhead := strconv.Atoi(countsParts[0])
	if errAhead != nil {
		return StatusError, 0, 0, fmt.Errorf("failed to parse 'ahead' count from '%s': %w", countsParts[0], errAhead)
	}

	behind, errBehind := strconv.Atoi(countsParts[1])
	if errBehind != nil {
		return StatusError, 0, 0, fmt.Errorf("failed to parse 'behind' count from '%s': %w", countsParts[1], errBehind)
	}

	if ahead > 0 && behind > 0 {
		return StatusDiverged, ahead, behind, nil
	}
	if ahead > 0 {
		return StatusAhead, ahead, behind, nil
	}
	if behind > 0 {
		return StatusBehind, ahead, behind, nil
	}
	return StatusUpToDate, ahead, behind, nil
}
