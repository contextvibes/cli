package feedback

import (
	"log/slog"

	"github.com/contextvibes/cli/internal/config"
)

// This file is only compiled during testing.
// It bridges internal logic to the external test package.

// Export the private struct type so tests can reference it.
type FeedbackParams = feedbackParams

// Export internal functions by assigning them to public variables.
var (
	ParseFeedbackArgs = parseFeedbackArgs
	ResolveTarget     = resolveTarget
	ConstructWorkItem = constructWorkItem
)

// NewTestParams is a helper to construct the private feedbackParams struct
// from the external test package. This avoids making fields public in production code.
func NewTestParams(cfg *config.FeedbackSettings, title, body string) *FeedbackParams {
	return &feedbackParams{
		cfg:   cfg,
		title: title,
		body:  body,
		// Initialize other fields with safe defaults if needed for tests
		logger: slog.Default(),
	}
}
