package apply

import "github.com/contextvibes/cli/internal/codemod"

// ChangePlan defines the top-level structure for a declarative plan.
type ChangePlan struct {
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}

// Step represents a single action in the ChangePlan.
type Step struct {
	Type        string `json:"type"`
	Description string `json:"description"`

	// Fields for "file_modification" type
	Changes codemod.ChangeScript `json:"changes,omitempty"`

	// Fields for "command_execution" type
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}
