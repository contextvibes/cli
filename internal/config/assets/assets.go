package assets

import (
	"embed"
	"fmt"
)

//go:embed all:templates
var templates embed.FS

//go:embed go/golangci-strict.yml
var golangCIStrict []byte

//go:embed go/lint-style.yml
var golangCIStyle []byte

//go:embed go/lint-complexity.yml
var golangCIComplexity []byte

//go:embed go/lint-security.yml
var golangCISecurity []byte

// AssetType defines the type of asset to be retrieved.
type AssetType string

const (
	// LinterConfigStrict is the asset name for the strict linter config.
	LinterConfigStrict AssetType = "linter-config-strict"
	// LinterConfigStyle is the asset name for the style linter config.
	LinterConfigStyle AssetType = "linter-config-style"
	// LinterConfigComplexity is the asset name for the complexity linter config.
	LinterConfigComplexity AssetType = "linter-config-complexity"
	// LinterConfigSecurity is the asset name for the security linter config.
	LinterConfigSecurity AssetType = "linter-config-security"
	// AIContextPrompt is the asset name for the AI context prompt.
	AIContextPrompt AssetType = "ai-context-prompt"
)

// languageAssets maps a language to its available assets.
//
//nolint:gochecknoglobals // This is a read-only map of embedded assets.
var languageAssets = map[string]map[AssetType][]byte{
	"go": {
		LinterConfigStrict:     golangCIStrict,
		LinterConfigStyle:      golangCIStyle,
		LinterConfigComplexity: golangCIComplexity,
		LinterConfigSecurity:   golangCISecurity,
	},
}

// GetLanguageAsset retrieves a configuration asset for a specific language.
func GetLanguageAsset(language string, assetType AssetType) ([]byte, error) {
	if assets, ok := languageAssets[language]; ok {
		if asset, ok := assets[assetType]; ok {
			return asset, nil
		}
	}

	// Fallback for assets that are not language-specific
	if assetType == AIContextPrompt {
		data, err := templates.ReadFile("templates/ai_context_header.md.tmpl")
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded template: %w", err)
		}

		return data, nil
	}

	return nil, fmt.Errorf("no asset defined for language '%s' and type '%s'", language, assetType)
}
