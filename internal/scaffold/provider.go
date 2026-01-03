package scaffold

import (
	"embed"
	"fmt"
	"path"
)

//go:embed assets
var assetsFS embed.FS

// Provider handles retrieving scaffold templates.
type Provider struct{}

// NewProvider creates a new scaffold provider.
func NewProvider() *Provider {
	return &Provider{}
}

// GetFiles returns a map of filename -> content for a specific target (e.g., "idx", "vscode").
func (p *Provider) GetFiles(target string) (map[string]string, error) {
	targetDir := path.Join("assets", target)

	entries, err := assetsFS.ReadDir(targetDir)
	if err != nil {
		return nil, fmt.Errorf("target '%s' not found in assets: %w", target, err)
	}

	files := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := assetsFS.ReadFile(path.Join(targetDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read asset %s: %w", entry.Name(), err)
		}

		files[entry.Name()] = string(content)
	}

	return files, nil
}
