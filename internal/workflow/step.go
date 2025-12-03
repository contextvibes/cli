package workflow

import "context"

// PresenterInterface defines the UI methods the workflow engine needs.
// This decouples the engine from the concrete ui.Presenter.
//
//nolint:interfacebloat // The presenter facade naturally requires many methods.
type PresenterInterface interface {
	Error(format string, a ...any)
	Warning(format string, a ...any)
	Info(format string, a ...any)
	Success(format string, a ...any)
	Detail(format string, a ...any)
	Step(format string, a ...any)
	Header(format string, a ...any)
	Summary(format string, a ...any)
	Newline()
	PromptForConfirmation(prompt string) (bool, error)
	PromptForInput(prompt string) (string, error)
	Advice(format string, a ...any)
}

// Step represents a single, discrete, and reusable action within a larger workflow.
type Step interface {
	// Description returns a user-friendly string explaining what this step will do.
	// This is used to build the "plan" shown to the user before confirmation.
	Description() string

	// PreCheck runs any prerequisite validations before the main execution.
	// If it returns an error, the entire workflow is aborted.
	PreCheck(ctx context.Context) error

	// Execute performs the primary action of the step.
	Execute(ctx context.Context) error
}
