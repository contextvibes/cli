package workflow

import (
	"context"
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

	// 1. Run all pre-checks first.
	for _, step := range steps {
		if err := step.PreCheck(ctx); err != nil {
			// The step's PreCheck is responsible for its own user-facing error message.
			return err
		}
	}
	r.presenter.Success("✓ All prerequisite checks passed.")
	r.presenter.Newline()

	// 2. Explain the plan.
	r.presenter.Info("Proposed Workflow Plan:")
	for i, step := range steps {
		r.presenter.Detail("%d. %s", i+1, step.Description())
	}
	r.presenter.Newline()

	// 3. Confirm with the user.
	if !r.assumeYes {
		confirmed, err := r.presenter.PromptForConfirmation("Proceed with this workflow?")
		if err != nil {
			return err
		}
		if !confirmed {
			r.presenter.Info("Workflow aborted by user.")
			return nil
		}
	} else {
		r.presenter.Info("Confirmation bypassed via --yes flag.")
	}
	r.presenter.Newline()

	// 4. Execute all steps.
	for _, step := range steps {
		r.presenter.Step(step.Description())
		if err := step.Execute(ctx); err != nil {
			// The step's Execute method is responsible for its own user-facing error message.
			return err // Abort on first failure.
		}
	}

	r.presenter.Newline()
	r.presenter.Success("Workflow completed successfully.")

	// ADDED: After success, check if a stash was made and provide advice.
	for _, step := range steps {
		if stashStep, ok := step.(*CheckAndPromptStashStep); ok && stashStep.DidStash {
			r.presenter.Advice(
				"Your uncommitted changes were stashed. Run `git stash pop` to restore them.",
			)
			break // Found it, no need to check further.
		}
	}

	return nil
}
