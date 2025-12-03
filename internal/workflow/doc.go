/*
Package workflow provides a structured engine for executing sequential, multi-step
CLI operations. It separates the "what" (Steps) from the "how" (Runner), allowing
for reusable logic, consistent error handling, and standardized user interaction.

# Core Concepts

  - Step: An interface defining a single unit of work. It has a Description,
    a PreCheck (validation), and an Execute (action) phase.
  - Runner: The engine that orchestrates the execution of Steps. It handles
    UI presentation, confirmation prompts, and error propagation.
  - PresenterInterface: An abstraction for UI output, allowing the workflow
    to be tested without writing to actual stdout/stderr.

# Usage Pattern

	runner := workflow.NewRunner(presenter, assumeYes)
	err := runner.Run(ctx, "My Workflow Title",
	    &MyStep1{},
	    &MyStep2{},
	)

# PreChecks vs Execution

The Runner executes all PreChecks first. If any PreCheck fails, the workflow
aborts immediately. This ensures the environment is safe before any state
changes occur. If all PreChecks pass, the Runner (optionally) prompts the
user for confirmation before proceeding to the Execute phase of each step.
*/
package workflow
