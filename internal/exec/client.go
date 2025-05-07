// internal/exec/client.go
package exec

import (
	"context"
	"log/slog"
)

// ExecutorClient provides a high-level interface for running external commands.
// It uses an underlying CommandExecutor for the actual execution.
type ExecutorClient struct {
	executor CommandExecutor
}

// NewClient creates a new ExecutorClient with the given CommandExecutor.
func NewClient(executor CommandExecutor) *ExecutorClient {
	return &ExecutorClient{executor: executor}
}

// Execute runs a command, typically piping stdio. See CommandExecutor.Execute.
func (c *ExecutorClient) Execute(ctx context.Context, dir string, commandName string, args ...string) error {
	return c.executor.Execute(ctx, dir, commandName, args...)
}

// CaptureOutput runs a command and captures its stdout and stderr. See CommandExecutor.CaptureOutput.
func (c *ExecutorClient) CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error) {
	return c.executor.CaptureOutput(ctx, dir, commandName, args...)
}

// CommandExists checks if a command is available. See CommandExecutor.CommandExists.
func (c *ExecutorClient) CommandExists(commandName string) bool {
	return c.executor.CommandExists(commandName)
}

// Logger returns the logger from the underlying executor.
func (c *ExecutorClient) Logger() *slog.Logger {
	return c.executor.Logger()
}
