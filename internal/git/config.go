// internal/git/config.go
// (Configuration for the GitClient)

package git

import (
	"io"
	"log/slog"
)

// GitClientConfig holds configuration for the GitClient.
type GitClientConfig struct {
	// GitExecutable is the path to the Git executable.
	// If empty, "git" will be looked up in PATH.
	GitExecutable string

	// DefaultRemoteName specifies the default remote to use (e.g., "origin").
	DefaultRemoteName string

	// DefaultMainBranchName specifies the default main branch name (e.g., "main", "master").
	DefaultMainBranchName string

	// Logger is the structured logger instance for the client's operations.
	// If nil, a default discard logger will be used.
	Logger *slog.Logger

	// Executor is an optional custom executor for testing or specialized execution.
	// If nil, a default commandExecutor will be used.
	Executor executor
}

// validateAndSetDefaults checks the config and applies defaults.
// Returns a validated config or an error.
func (c GitClientConfig) validateAndSetDefaults() (GitClientConfig, error) {
	validated := c // Start with a copy

	if validated.DefaultRemoteName == "" {
		validated.DefaultRemoteName = "origin"
	}
	if validated.DefaultMainBranchName == "" {
		validated.DefaultMainBranchName = "main"
	}
	if validated.GitExecutable == "" {
		validated.GitExecutable = "git" // Will be resolved by PATH
	}
	if validated.Logger == nil {
		// Provide a discard logger if none is given.
		validated.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if validated.Executor == nil {
		validated.Executor = NewCommandExecutor() // Default command executor
	}
	// Add any other validation logic here (e.g., check if names are valid)
	return validated, nil
}
