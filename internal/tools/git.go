package tools

import (
	"os"
	"path/filepath"
)

// IsGitRepo checks if a directory contains a .git subdirectory or file (for worktrees).
// This is a simple check and does not use any external command execution client.
func IsGitRepo(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}
