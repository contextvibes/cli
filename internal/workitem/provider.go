// internal/workitem/provider.go
package workitem

import "context"

// Provider defines the interface for a work item management system.
// This allows for abstracting the backend (GitHub, GitLab, etc.).
type Provider interface {
	// ListItems retrieves a collection of work items based on the provided options.
	ListItems(ctx context.Context, options ListOptions) ([]WorkItem, error)

	// GetItem retrieves a single work item by its public number, optionally fetching comments.
	GetItem(ctx context.Context, number int, withComments bool) (*WorkItem, error)

	// CreateItem creates a new work item in the backend system.
	CreateItem(ctx context.Context, item WorkItem) (*WorkItem, error)

	// UpdateItem updates an existing work item in the backend system.
	UpdateItem(ctx context.Context, number int, item WorkItem) (*WorkItem, error)

	// SearchItems uses a provider-specific query string to find work items.
	SearchItems(ctx context.Context, query string) ([]WorkItem, error)

	// CreateLabel creates a new label in the backend system.
	CreateLabel(ctx context.Context, label Label) (*Label, error)
}
