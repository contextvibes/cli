package tools

import (
	"bytes"
	"fmt"
	"strings"
)

// AppendSectionHeader adds a standard Markdown H3 section header to the buffer.
func AppendSectionHeader(buf *bytes.Buffer, title string) {
	buf.WriteString("### ")
	buf.WriteString(title)
	buf.WriteString("\n\n")
}

// AppendFencedCodeBlock adds a standard Markdown fenced code block to the buffer.
func AppendFencedCodeBlock(buf *bytes.Buffer, content string, languageHint string) {
	buf.WriteString("```")

	if languageHint != "" {
		buf.WriteString(languageHint)
	}

	buf.WriteString("\n")
	// Ensure content ends with a newline before the closing fence
	// But avoid adding a double newline if one already exists
	trimmedContent := strings.TrimSuffix(content, "\n")
	buf.WriteString(trimmedContent)
	buf.WriteByte('\n') // Ensure at least one newline

	buf.WriteString("```\n\n")
}

// AppendFileMarkerHeader adds the explicit file start marker.
func AppendFileMarkerHeader(buf *bytes.Buffer, filePath string) {
	// Ensure preceding content has adequate spacing, but avoid excessive newlines
	trimmedBytes := bytes.TrimRight(buf.Bytes(), "\n")
	buf.Reset()
	buf.Write(trimmedBytes)
	// Add consistent spacing before the header
	fmt.Fprintf(buf, "\n\n======== FILE: %s ========\n\n", filePath)
}

// AppendFileMarkerFooter adds the explicit file end marker.
func AppendFileMarkerFooter(buf *bytes.Buffer, filePath string) {
	// Simpler approach: Trim all trailing whitespace, then add exactly two newlines before the footer.
	trimmedBytes := bytes.TrimSpace(buf.Bytes())
	buf.Reset()             // Clear the buffer
	buf.Write(trimmedBytes) // Write back the trimmed content
	// Add exactly two newlines before the footer marker
	fmt.Fprintf(buf, "\n\n======== END FILE: %s ========\n\n", filePath)
}
