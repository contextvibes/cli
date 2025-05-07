// internal/exec/doc.go

/*
Package exec provides a client and interface for executing external commands.
It abstracts the underlying os/exec calls, allowing for easier testing and
consistent command execution logic throughout the application.

The primary components are:
  - CommandExecutor: An interface defining methods to run commands and capture output.
  - OSCommandExecutor: The default implementation of CommandExecutor using os/exec.
  - ExecutorClient: A client that uses a CommandExecutor to provide higher-level
    methods for command execution.

Usage:

	// In your application setup (e.g., cmd/root.go or per command)
	osExecutor := exec.NewOSCommandExecutor(someLogger) // Pass an *slog.Logger
	execClient := exec.NewClient(osExecutor)

	// Later, to run a command:
	err := execClient.Execute(ctx, "/tmp", "ls", "-l")
	if err != nil {
		// handle error
	}

	// To capture output:
	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", "version")
	if err != nil {
		// handle error
	}
	fmt.Printf("Go version: %s", stdout)
*/
package exec
