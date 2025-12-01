// Package git provides a high-level client for interacting with Git repositories.
package git

import (
	"log/slog"

	"github.com/contextvibes/cli/internal/exec" // Import the new exec package
)

// GitClientConfig configures the GitClient.
type GitClientConfig struct {
	GitExecutable         string
	DefaultRemoteName     string               // Will be set by cmd layer from LoadedAppConfig
	DefaultMainBranchName string               // Will be set by cmd layer from LoadedAppConfig
	Logger                *slog.Logger         // Logger for GitClient's own operations
	Executor              exec.CommandExecutor // Use the new CommandExecutor interface from internal/exec
}

// validateAndSetDefaults now expects that if an Executor is needed, it's either provided,
// or can be created using a logger that should also be provided (or defaulted).
// The primary logger for the application (AppLogger) can be passed to create a default executor.
//
//nolint:unparam // Error return is currently unused but kept for future validation.
func (c GitClientConfig) validateAndSetDefaults() (GitClientConfig, error) {
	validated := c

	// DefaultRemoteName and DefaultMainBranchName are expected to be set by the caller (cmd/root.go)
	// using the application's LoadedAppConfig. We still provide fallbacks here for safety.
	if validated.DefaultRemoteName == "" {
		validated.DefaultRemoteName = "origin"
	}

	if validated.DefaultMainBranchName == "" {
		validated.DefaultMainBranchName = "main"
	}

	if validated.GitExecutable == "" {
		validated.GitExecutable = "git" // Default to looking for 'git' in PATH
	}

	// Logger for the GitClient itself. If not provided, it might inherit from the Executor's logger,
	// or use a specific one. For simplicity, let's ensure it has one.
	// The AppLogger from cmd/root.go is a good candidate to pass into here.
	if validated.Logger == nil {
		// This state (nil Logger in GitClientConfig) should ideally be avoided by the caller.
		// If the Executor is also nil, its creation below would also lack a logger.
		// For robustness, if Executor is also nil, its default creation will use its own discard/default logger.
		// If Executor is provided, GitClient can use its logger.
		if validated.Executor != nil {
			validated.Logger = validated.Executor.Logger()
		} else {
			// Fallback: create a new OS executor which will have its own default/discard logger
			// and use that logger for the GitClient. This is less ideal than injecting AppLogger.
			tempExecutor := exec.NewOSCommandExecutor(nil) // Creates OS executor with a discard logger
			validated.Logger = tempExecutor.Logger()
		}
	}

	// Executor for running git commands.
	if validated.Executor == nil {
		// If no executor is provided, create a default OSCommandExecutor.
		// It's crucial that this OSCommandExecutor uses a proper logger.
		// Pass the GitClient's logger (which should have been resolved above, possibly from AppLogger).
		validated.Executor = exec.NewOSCommandExecutor(validated.Logger)
	}

	return validated, nil
}
