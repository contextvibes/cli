// Package pipeline provides an engine for running analytical workflows (quality checks).
package pipeline

import (
	"context"

	"github.com/contextvibes/cli/internal/exec"
)

// Status represents the outcome of a check.
type Status int

const (
	// StatusPass indicates the check succeeded.
	StatusPass Status = iota
	// StatusFail indicates the check failed (critical).
	StatusFail
	// StatusWarn indicates a non-critical issue or skipped check.
	StatusWarn
)

// Result captures the output of a single check.
type Result struct {
	Name    string
	Status  Status
	Message string
	Advice  string
	Details string // Added to hold raw output (like deadcode list)
	Error   error
}

// Check defines a single unit of analysis.
type Check interface {
	Name() string
	Run(ctx context.Context, exec *exec.ExecutorClient) Result
}
