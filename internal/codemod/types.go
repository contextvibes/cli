package codemod

// Operation defines a single modification to be performed on a file.
type Operation struct {
	// Type indicates the kind of operation (e.g., "regex_replace", "add_import").
	Type string `json:"type"`
	// Description provides a human-readable explanation of the operation.
	Description string `json:"description,omitempty"`

	// --- Fields for "regex_replace" type ---
	// FindRegex is the regular expression to find.
	//nolint:tagliatelle // JSON keys are fixed by schema.
	FindRegex string `json:"find_regex,omitempty"`
	// ReplaceWith is the string to replace matches with.
	//nolint:tagliatelle // JSON keys are fixed by schema.
	ReplaceWith string `json:"replace_with,omitempty"`

	// --- Fields for "create_or_overwrite" ---
	Content *string `json:"content,omitempty"` // Pointer to distinguish empty from not-set
	// LineNumber can be used to target a specific line for some operations (not used by basic regex_replace yet).
	//nolint:tagliatelle // JSON keys are fixed by schema.
	LineNumber *int `json:"line_number,omitempty"`
}

// FileChangeSet groups all operations for a single file.
type FileChangeSet struct {
	//nolint:tagliatelle // JSON keys are fixed by schema.
	FilePath   string      `json:"file_path"`
	Operations []Operation `json:"operations"`
}

// ChangeScript is the top-level structure, representing a list of changes for multiple files.
type ChangeScript []FileChangeSet
