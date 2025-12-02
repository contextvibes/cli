package tools

import (
	"bytes"
	"fmt"
	"os"
)

// ReadFileContent reads the entire content of the file at the specified path.
// Returns the content as a byte slice or an error if reading fails.
func ReadFileContent(filePath string) ([]byte, error) {
	//nolint:gosec // Generic file read helper.
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Wrap the error with more context.
		return nil, fmt.Errorf("error reading file '%s': %w", filePath, err)
	}

	return content, nil
}

// WriteBufferToFile writes the content of a bytes.Buffer to the specified file path.
// It uses default file permissions (0600).
// It prints informational messages about writing to os.Stdout.
//nolint:godox // TODO is acceptable here.
// TODO: Refactor to remove direct fmt.Printf calls.
//
//	Calling commands should use their Presenter for user-facing messages
//	or a Logger for debug/trace information related to file writing.
//	This function should focus solely on writing the file.
func WriteBufferToFile(filePath string, buf *bytes.Buffer) error {
	// These fmt.Printf calls directly write to os.Stdout.
	// They are currently used by cmd/diff and cmd/describe.
	// Future refactoring might replace these with presenter calls from the cmd layer.
	//nolint:forbidigo // Legacy logging to stdout.
	fmt.Printf("INFO: Writing output to %s...\n", filePath)

	//nolint:mnd // 0600 is standard file permission.
	err := os.WriteFile(filePath, buf.Bytes(), 0o600) // Use standard file permissions
	if err != nil {
		// Wrap the error with more context.
		return fmt.Errorf("failed to write output file '%s': %w", filePath, err)
	}

	//nolint:forbidigo // Legacy logging to stdout.
	fmt.Printf("INFO: Successfully wrote %s.\n", filePath)

	return nil
}
