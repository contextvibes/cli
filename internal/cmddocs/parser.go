package cmddocs

import (
	"bytes"
	"strings"
	"text/template"
)

// CommandDesc contains the short and long descriptions for a cobra command.
type CommandDesc struct {
	Short string
	Long  string
}

// ParseAndExecute parses a Markdown template, executes it with the provided data,
// and extracts the Short and Long descriptions.
// It expects the first non-empty line after templating to be an H1 title (e.g., "# Command Short Desc").
func ParseAndExecute(templateContent string, data any) (CommandDesc, error) {
	tmpl, err := template.New("cmddoc").Parse(templateContent)
	if err != nil {
		return CommandDesc{}, err
	}

	var executedTpl bytes.Buffer
	if err := tmpl.Execute(&executedTpl, data); err != nil {
		return CommandDesc{}, err
	}

	var shortDesc, longDesc string

	lines := strings.Split(executedTpl.String(), "\n")
	foundTitle := false

	var longLines []string

	for _, line := range lines {
		if !foundTitle && strings.HasPrefix(line, "# ") {
			shortDesc = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			foundTitle = true

			continue
		}

		if foundTitle {
			longLines = append(longLines, line)
		}
	}

	if !foundTitle && len(lines) > 0 {
		// Fallback if no H1 is found
		shortDesc = lines[0]
		longDesc = strings.Join(lines[1:], "\n")
	} else {
		longDesc = strings.TrimSpace(strings.Join(longLines, "\n"))
	}

	return CommandDesc{Short: shortDesc, Long: longDesc}, nil
}
