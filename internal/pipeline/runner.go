package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/ui"
)

// Runner executes a list of checks and aggregates the results.
type Runner struct {
	presenter  *ui.Presenter
	execClient *exec.ExecutorClient
}

// NewRunner creates a new pipeline runner.
func NewRunner(p *ui.Presenter, e *exec.ExecutorClient) *Runner {
	return &Runner{
		presenter:  p,
		execClient: e,
	}
}

// Run executes all provided checks sequentially.
// It returns the list of results and an error if any check failed.
func (r *Runner) Run(ctx context.Context, checks []Check) ([]Result, error) {
	//nolint:prealloc // Pre-allocating is good practice but not critical here.
	var results []Result

	var failureCount int

	for _, check := range checks {
		r.presenter.Step("Running check: %s...", check.Name())

		result := check.Run(ctx, r.execClient)
		results = append(results, result)

		r.reportResult(result)

		if result.Status == StatusFail {
			failureCount++
		}

		r.presenter.Newline()
	}

	if failureCount > 0 {
		//nolint:err113 // Dynamic error is appropriate here.
		return results, fmt.Errorf("%d quality check(s) failed", failureCount)
	}

	return results, nil
}

func (r *Runner) reportResult(result Result) {
	switch result.Status {
	case StatusPass:
		msg := result.Message
		if msg == "" {
			msg = "Passed"
		}

		r.presenter.Success("✓ %s", msg)
	case StatusWarn:
		r.presenter.Warning("~ %s", result.Message)
		r.printDetails(result)

		if result.Advice != "" {
			r.presenter.Advice(result.Advice)
		}
	case StatusFail:
		r.presenter.Error("! Check failed: %s", result.Message)
		r.printDetails(result)

		if result.Error != nil {
			r.presenter.Detail("Error: %v", result.Error)
		}

		if result.Advice != "" {
			r.presenter.Advice(result.Advice)
		}
	}
}

func (r *Runner) printDetails(result Result) {
	if result.Details != "" {
		//nolint:modernize // SplitSeq is too new for some environments, stick to Split.
		lines := strings.Split(strings.TrimSpace(result.Details), "\n")
		for _, line := range lines {
			r.presenter.Detail(line)
		}
	}
}
