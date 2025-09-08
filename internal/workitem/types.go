// internal/workitem/types.go
package workitem

import "time"

// Type represents the classification of a work item (e.g., Story, Task).
type Type string

const (
	TypeStory Type = "Story"
	TypeTask  Type = "Task"
	TypeEpic  Type = "Epic"
	TypeBug   Type = "Bug"
	TypeChore Type = "Chore"
)

// State represents the status of a work item (e.g., Open, Closed).
type State string

const (
	StateOpen   State = "Open"
	StateClosed State = "Closed"
)

// Comment represents a single comment on a work item.
type Comment struct {
	Author    string
	Body      string
	CreatedAt time.Time
	URL       string
}

// Label represents a label that can be applied to a work item.
type Label struct {
	Name        string
	Description string
	Color       string
}

// WorkItem is the generic, provider-agnostic representation of a work item.
// It includes a Children slice to allow for building hierarchical trees.
type WorkItem struct {
	// Provider-specific ID (e.g., GitHub's GraphQL node ID).
	ID string
	// Publicly visible number for the item (e.g., GitHub issue #123).
	Number int
	// The title or summary of the work item.
	Title string
	// The detailed description or body of the work item.
	Body string
	// The current state of the item.
	State State
	// The classification of the item.
	Type Type
	// The web URL to view the item in its native system.
	URL string
	// The username of the author.
	Author string
	// A list of associated labels or tags.
	Labels []string
	// A list of usernames assigned to the item.
	Assignees []string
	// When the item was created.
	CreatedAt time.Time
	// When the item was last updated.
	UpdatedAt time.Time
	// A list of comments on the item.
	Comments []Comment
	// A slice of child work items to represent a hierarchy.
	Children []*WorkItem
}

// ListOptions provides filters and pagination for listing work items.
type ListOptions struct {
	State    State
	Labels   []string
	Assignee string
	Limit    int
	Page     int
}
