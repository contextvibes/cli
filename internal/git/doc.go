// internal/git/doc.go

/*
Package git provides a high-level client for interacting with Git repositories
programmatically from Go applications. It abstracts the direct execution of 'git'
command-line operations, offering a more Go-idiomatic API.

The primary entry point for using this package is the GitClient type, which
is instantiated via the NewClient function. The client requires a working directory
to determine the repository context and can be configured using GitClientConfig.

Key features include:
  - Repository context detection (finding .git and top-level directories).
  - Execution of common Git commands (status, commit, add, branch, etc.)
    through structured methods.
  - Abstraction over command execution, allowing for custom executors (primarily for testing).
  - Integration with structured logging via the slog package.

Usage Example:

	// Assume globalLogger is an initialized *slog.Logger
	ctx := context.Background()
	workDir, _ := os.Getwd()

	gitCfg := git.GitClientConfig{
		Logger: globalLogger,
		// Other configurations can be set here
	}

	client, err := git.NewClient(ctx, workDir, gitCfg)
	if err != nil {
		log.Fatalf("Failed to create Git client: %v", err)
	}

	branch, err := client.GetCurrentBranchName(ctx)
	if err != nil {
		log.Printf("Error getting branch: %v", err)
	} else {
		log.Printf("Current branch: %s", branch)
	}

Error Handling:

Methods on the GitClient typically return an error as their last argument if an
operation fails. Errors can originate from underlying git command failures,
invalid input, or issues with the repository state. It's important for callers
to check these errors.

Logging:

The GitClient uses an slog.Logger instance provided via GitClientConfig.
This allows for consistent, structured logging of its operations, which can be
directed to various outputs (e.g., console, files, AI-consumable streams)
by configuring the logger's handlers at the application level.

Testing:

The GitClient is designed with testability in mind. The GitClientConfig.Executor
field allows injecting a mock 'executor' interface, enabling unit tests for
client methods without relying on an actual 'git' executable or a live repository.
*/
package git
