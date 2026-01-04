package config

import (
	"embed"
	"fmt"
	"path"
)

//go:embed assets/*
var assetsFS embed.FS

// AssetType defines the category of the asset.
type AssetType string

const (
	// AssetLintStrict represents the full strict configuration.
	AssetLintStrict AssetType = "lint-strict"
	// AssetLintSecurity represents security-focused configuration.
	AssetLintSecurity AssetType = "lint-security"
	// AssetLintComplexity represents complexity-focused configuration.
	AssetLintComplexity AssetType = "lint-complexity"
	// AssetLintStyle represents style-focused configuration.
	AssetLintStyle AssetType = "lint-style"
	// AIContextPrompt is the asset for the AI context header prompt.
	AIContextPrompt AssetType = "ai-context-prompt"
)

// assetMap holds the mapping from (language, assetType) to filename.
var assetMap = map[string]map[AssetType]string{
	"go": {
		AssetLintStrict:     "golangci-strict.yml",
		AssetLintSecurity:   "lint-security.yml",
		AssetLintComplexity: "lint-complexity.yml",
		AssetLintStyle:      "lint-style.yml",
		AIContextPrompt:     "templates/ai_context_header.md.tmpl",
	},
}

// GetLanguageAsset retrieves an embedded configuration file.
func GetLanguageAsset(language string, assetType AssetType) ([]byte, error) {
	langAssets, ok := assetMap[language]
	if !ok {
		return nil, fmt.Errorf("no assets defined for language '%s'", language)
	}

	filename, ok := langAssets[assetType]
	if !ok {
		return nil, fmt.Errorf("unknown asset type '%s' for language '%s'", assetType, language)
	}

	var basePath string
	if assetType == AIContextPrompt {
		basePath = path.Join("assets", "go")
	} else {
		basePath = path.Join("assets", language)
	}

	fullPath := path.Join(basePath, filename)

	content, err := assetsFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded asset '%s': %w", fullPath, err)
	}

	return content, nil
}
