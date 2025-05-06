// internal/git/executor.go
// (Defines how commands are run, allowing mocking)

package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// executor defines the interface for running external commands, specifically Git.
type executor interface {
	// Execute runs a git command, connecting stdio, and returns an error if execution fails.
	// dir: the working directory for the command.
	// command: should always be the path/name of the git executable.
	Execute(ctx context.Context, dir string, command string, args ...string) error

	// CaptureOutput runs a git command, capturing stdout and stderr.
	// dir: the working directory for the command.
	// command: should always be the path/name of the git executable.
	CaptureOutput(ctx context.Context, dir string, command string, args ...string) (stdout, stderr string, err error)

	// CommandExists checks if a command is available in the PATH or at the specified path.
	CommandExists(command string) bool
}

// --- Default Implementation (using os/exec) ---

type commandExecutor struct{}

func NewCommandExecutor() executor {
	return &commandExecutor{}
}

func (e *commandExecutor) Execute(ctx context.Context, dir string, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	// Pipe standard streams directly - useful for commands like 'git status' or interactive ones
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		// Command ran but exited non-zero
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Access exit code if needed
			_ = exitErr.ExitCode()
			// Stderr already piped, return the original error which includes exit status
			return err
		}
		// Other errors (e.g., command not found, context cancelled before start)
		return fmt.Errorf("failed to execute command '%s %v': %w", command, args, err)
	}
	return nil
}

func (e *commandExecutor) CaptureOutput(ctx context.Context, dir string, command string, args ...string) (string, string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run() // Run the command and wait for completion

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if err != nil {
		// Enhance error message with stderr content if available
		errMsg := fmt.Sprintf("command '%s %v' failed in dir '%s'", command, args, dir)
		if exitErr, ok := err.(*exec.ExitError); ok {
			errMsg = fmt.Sprintf("%s with exit code %d", errMsg, exitErr.ExitCode())
		} else {
			errMsg = fmt.Sprintf("%s: %s", errMsg, err.Error()) // Add underlying error kind
		}
		if stderrStr != "" {
			errMsg = fmt.Sprintf("%s. Stderr: %s", errMsg, strings.TrimSpace(stderrStr))
		}
		// Return the combined error message, preserving the original error type if possible
		// (though wrapping often loses the original type unless using Go 1.13+ error wrapping carefully)
		// For simplicity, return a formatted error string built upon the original error.
		return stdoutStr, stderrStr, fmt.Errorf("%s: %w", errMsg, err)
	}

	return stdoutStr, stderrStr, nil // err is nil here
}

func (e *commandExecutor) CommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
