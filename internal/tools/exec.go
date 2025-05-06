package tools

import (
	"bytes"
	"context" // Import context
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandExists checks if a command executable name is found in the system's PATH.
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ExecuteCommand runs an arbitrary command, piping its stdout, stderr, and stdin.
// It's a utility for commands where output capture isn't needed, just execution.
// TODO: This function pipes the command's stdio directly. Callers (e.g., commands in cmd/)
//
//	should use their ui.Presenter instance to announce the command *before* calling this function
//	if user-facing status information is desired.
func ExecuteCommand(cwd, commandName string, args ...string) error {
	// Using context.Background() as a default; callers could potentially pass one in
	// if cancellation needed propagating, but the function signature doesn't support it currently.
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Removed direct fmt.Printf("-> Running: ...") call to separate UI concerns.
	// Callers should handle announcements via the Presenter.

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Stderr was already piped to os.Stderr
			// Return an error that includes the exit code but avoids duplicating stderr.
			return fmt.Errorf("command '%s %s' failed with exit code %d", commandName, strings.Join(args, " "), exitErr.ExitCode())
		}
		// Other execution errors (e.g., command not found, context issues)
		return fmt.Errorf("failed to execute command '%s %s': %w", commandName, strings.Join(args, " "), err)
	}
	return nil
}

// CaptureCommandOutput executes a command and captures its stdout and stderr.
// Returns stdout string, stderr string, and any error (including *exec.ExitError if command ran but exited non-zero).
func CaptureCommandOutput(cwd string, commandName string, args ...string) (stdoutStr string, stderrStr string, err error) {
	// Using context.Background() as a default.
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Dir = cwd
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdoutStr = stdoutBuf.String()
	stderrStr = stderrBuf.String()

	// Don't wrap error if it's nil
	if err != nil {
		// Construct a more informative error message including stderr content if available
		baseErrMsg := fmt.Sprintf("command '%s %s' failed", commandName, strings.Join(args, " "))
		if exitErr, ok := err.(*exec.ExitError); ok {
			baseErrMsg = fmt.Sprintf("%s with exit code %d", baseErrMsg, exitErr.ExitCode())
		} else {
			baseErrMsg = fmt.Sprintf("%s: %v", baseErrMsg, err) // Include original error type/msg
		}

		// Append stderr if it provides additional context not already in the main error
		trimmedStderr := strings.TrimSpace(stderrStr)
		if trimmedStderr != "" && !strings.Contains(err.Error(), trimmedStderr) {
			// Use %w for proper error wrapping if Go version supports it well enough
			// For broader compatibility or simpler messages, just append.
			err = fmt.Errorf("%s. Stderr: %s", baseErrMsg, trimmedStderr)
			// Or using wrapping: err = fmt.Errorf("%s. Stderr: %s: %w", baseErrMsg, trimmedStderr, err)
		} else {
			// If stderr is empty or already in the error message, just wrap the base message.
			err = fmt.Errorf("%s: %w", baseErrMsg, err)
		}
	}
	return stdoutStr, stderrStr, err // Return potentially wrapped error
}
