// internal/codemod/types.go
package codemod

// Operation defines a single modification to be performed on a file.
type Operation struct {
	// Type indicates the kind of operation (e.g., "regex_replace", "add_import").
	Type string `json:"type"`
	// Description provides a human-readable explanation of the operation.
	Description string `json:"description,omitempty"`

	// --- Fields for "regex_replace" type ---
	// FindRegex is the regular expression to find.
	FindRegex string `json:"find_regex,omitempty"`
	// ReplaceWith is the string to replace matches with.
	ReplaceWith string `json:"replace_with,omitempty"`
	// LineNumber can be used to target a specific line for some operations (not used by basic regex_replace yet).
	LineNumber *int `json:"line_number,omitempty"`

	// --- Fields for "delete_file" type ---
	// No specific fields needed for simple delete, relies on FileChangeSet.FilePath

	// Future operations might include:
	// For "add_import_if_missing":
	// ImportPath string `json:"import_path,omitempty"`

	// For "comment_update":
	// OldCommentRegex string `json:"old_comment_regex,omitempty"`
	// NewComment      string `json:"new_comment,omitempty"`
}

// FileChangeSet groups all operations for a single file.
type FileChangeSet struct {
	FilePath   string      `json:"file_path"`
	Operations []Operation `json:"operations"`
}

// ChangeScript is the top-level structure, representing a list of changes for multiple files.
type ChangeScript []FileChangeSet
