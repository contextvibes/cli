package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec" // Standard library exec
	"strings"
)

// OSCommandExecutor is the default implementation of CommandExecutor using the os/exec package.
type OSCommandExecutor struct {
	logger *slog.Logger
}

// NewOSCommandExecutor creates a new OSCommandExecutor.
// If logger is nil, a discard logger will be used.
//
//nolint:ireturn // Returning interface is intended.
func NewOSCommandExecutor(logger *slog.Logger) CommandExecutor {
	log := logger
	if log == nil {
		log = slog.New(slog.DiscardHandler) // Default to discard if no logger provided
	}

	return &OSCommandExecutor{logger: log}
}

// Logger returns the logger associated with this executor.
func (e *OSCommandExecutor) Logger() *slog.Logger {
	return e.logger
}

// Execute runs a command, piping stdio.
func (e *OSCommandExecutor) Execute(
	ctx context.Context,
	dir string,
	commandName string,
	args ...string,
) error {
	e.logger.DebugContext(ctx, "Executing command",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args),
		slog.String("dir", dir))

	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout // Pipe directly
	cmd.Stderr = os.Stderr // Pipe directly
	cmd.Stdin = os.Stdin   // Pipe directly

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			e.logger.ErrorContext(ctx, "Command failed with exit code",
				slog.String("component", "OSCommandExecutor"),
				slog.String("command", commandName),
				slog.Any("args", args),
				slog.Int("exit_code", exitErr.ExitCode()),
				slog.String("error", err.Error()))
			// Stderr already piped. Return error that includes exit code info.
			return fmt.Errorf(
				"command '%s %s' failed with exit code %d: %w",
				commandName,
				strings.Join(args, " "),
				exitErr.ExitCode(),
				err,
			)
		}

		e.logger.ErrorContext(ctx, "Failed to execute command",
			slog.String("component", "OSCommandExecutor"),
			slog.String("command", commandName),
			slog.Any("args", args),
			slog.String("error", err.Error()))

		return fmt.Errorf(
			"failed to start or execute command '%s %s': %w",
			commandName,
			strings.Join(args, " "),
			err,
		)
	}

	e.logger.InfoContext(ctx, "Command executed successfully",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args))

	return nil
}

// CaptureOutput runs a command and captures its output.
func (e *OSCommandExecutor) CaptureOutput(
	ctx context.Context,
	dir string,
	commandName string,
	args ...string,
) (string, string, error) {
	e.logger.DebugContext(ctx, "Capturing command output",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args),
		slog.String("dir", dir))

	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if err != nil {
		var exitErr *exec.ExitError
		// Construct a more informative error message
		errMsg := fmt.Sprintf(
			"command '%s %s' in dir '%s' failed",
			commandName,
			strings.Join(args, " "),
			dir,
		)
		if errors.As(err, &exitErr) {
			errMsg = fmt.Sprintf("%s with exit code %d", errMsg, exitErr.ExitCode())
		} else {
			errMsg = fmt.Sprintf("%s: %v", errMsg, err) // Include original error type/msg for non-ExitErrors
		}

		trimmedStderr := strings.TrimSpace(stderrStr)
		if trimmedStderr != "" {
			errMsg = fmt.Sprintf("%s. Stderr: %s", errMsg, trimmedStderr)
		}

		e.logger.ErrorContext(ctx, "Command capture failed",
			slog.String("component", "OSCommandExecutor"),
			slog.String("command", commandName),
			slog.Any("args", args),
			slog.String("stdout_capture_len", fmt.Sprintf("%d bytes", len(stdoutStr))),
			slog.String("stderr_capture_len", fmt.Sprintf("%d bytes", len(stderrStr))),
			slog.String("error", err.Error()),     // Log the original simpler error
			slog.String("detailed_error", errMsg)) // Log the detailed constructed error

		//nolint:err113 // Dynamic error is appropriate here.
		return stdoutStr, stderrStr, fmt.Errorf(errMsg+": %w", err) // Wrap original error
	}

	e.logger.DebugContext(ctx, "Command capture successful",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args),
		slog.String("stdout_capture_len", fmt.Sprintf("%d bytes", len(stdoutStr))),
		slog.String("stderr_capture_len", fmt.Sprintf("%d bytes", len(stderrStr))))

	return stdoutStr, stderrStr, nil
}

// CommandExists checks if a command exists in the path.
func (e *OSCommandExecutor) CommandExists(commandName string) bool {
	_, err := exec.LookPath(commandName)
	if err != nil {
		e.logger.Debug("Command existence check: not found",
			slog.String("component", "OSCommandExecutor"),
			slog.String("command", commandName),
			slog.String("error", err.Error()))

		return false
	}

	e.logger.Debug("Command existence check: found",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName))

	return true
}
