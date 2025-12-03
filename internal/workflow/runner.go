package workflow

import (
	"context"
	"fmt"
)

// Runner manages the execution of a series of workflow steps.
type Runner struct {
	presenter PresenterInterface
	assumeYes bool
}

// NewRunner creates a new workflow runner.
func NewRunner(presenter PresenterInterface, assumeYes bool) *Runner {
	return &Runner{presenter: presenter, assumeYes: assumeYes}
}

// Run executes the entire workflow.
func (r *Runner) Run(ctx context.Context, title string, steps ...Step) error {
	r.presenter.Summary(title)

	err := r.runPreChecks(ctx, steps)
	if err != nil {
		return err
	}

	r.presenter.Success("✓ All prerequisite checks passed.")
	r.presenter.Newline()

	r.presentPlan(steps)

	err = r.confirmExecution()
	if err != nil {
		return err
	}

	err = r.executeSteps(ctx, steps)
	if err != nil {
		return err
	}

	r.presenter.Newline()
	r.presenter.Success("Workflow completed successfully.")

	r.checkStashAdvice(steps)

	return nil
}

func (r *Runner) runPreChecks(ctx context.Context, steps []Step) error {
	for _, step := range steps {
		err := step.PreCheck(ctx)
		if err != nil {
			// The step's PreCheck is responsible for its own user-facing error message.
			// We wrap it to satisfy linter, though the message might be redundant if printed by step.
			return fmt.Errorf("pre-check failed: %w", err)
		}
	}

	return nil
}

func (r *Runner) presentPlan(steps []Step) {
	r.presenter.Info("Proposed Workflow Plan:")

	for i, step := range steps {
		r.presenter.Detail("%d. %s", i+1, step.Description())
	}

	r.presenter.Newline()
}

func (r *Runner) confirmExecution() error {
	if r.assumeYes {
		r.presenter.Info("Confirmation bypassed via --yes flag.")
		r.presenter.Newline()

		return nil
	}

	confirmed, err := r.presenter.PromptForConfirmation("Proceed with this workflow?")
	if err != nil {
		return fmt.Errorf("confirmation prompt failed: %w", err)
	}

	if !confirmed {
		r.presenter.Info("Workflow aborted by user.")

		return nil // Return nil to stop without error
	}

	r.presenter.Newline()

	return nil
}

func (r *Runner) executeSteps(ctx context.Context, steps []Step) error {
	for _, step := range steps {
		r.presenter.Step(step.Description())

		err := step.Execute(ctx)
		if err != nil {
			// The step's Execute method is responsible for its own user-facing error message.
			return fmt.Errorf("step execution failed: %w", err) // Abort on first failure.
		}
	}

	return nil
}

func (r *Runner) checkStashAdvice(steps []Step) {
	// ADDED: After success, check if a stash was made and provide advice.
	for _, step := range steps {
		if stashStep, ok := step.(*CheckAndPromptStashStep); ok && stashStep.DidStash {
			r.presenter.Advice(
				"Your uncommitted changes were stashed. Run `git stash pop` to restore them.",
			)

			break // Found it, no need to check further.
		}
	}
}
