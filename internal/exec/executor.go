// internal/exec/executor.go
package exec

import (
	"context"
	"log/slog"
)

// CommandExecutor defines the interface for running external commands.
// Implementations of this interface handle the actual execution logic.
type CommandExecutor interface {
	// Execute runs a command, connecting stdio to the parent process's stdio.
	// dir: the working directory for the command.
	// commandName: the name or path of the command to run.
	// args: arguments for the command.
	// Returns an error if execution fails.
	Execute(ctx context.Context, dir string, commandName string, args ...string) error

	// CaptureOutput runs a command, capturing its stdout and stderr.
	// dir: the working directory for the command.
	// commandName: the name or path of the command to run.
	// args: arguments for the command.
	// Returns stdout, stderr, and any error (including *exec.PkgExitError).
	CaptureOutput(
		ctx context.Context,
		dir string,
		commandName string,
		args ...string,
	) (stdout, stderr string, err error)

	// CommandExists checks if a command is available in the PATH or at the specified path.
	CommandExists(commandName string) bool

	// Logger returns the logger associated with this executor.
	Logger() *slog.Logger
}
