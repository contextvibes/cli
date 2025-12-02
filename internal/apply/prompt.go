package apply

import _ "embed"

//go:embed assets/change_plan_prompt.md
var changePlanPrompt string

// GetChangePlanPrompt returns the content of the embedded AI prompt.
func GetChangePlanPrompt() string {
	return changePlanPrompt
}
