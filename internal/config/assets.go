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
)

// GetLanguageAsset retrieves an embedded configuration file.
func GetLanguageAsset(language string, assetType AssetType) ([]byte, error) {
	var filename string

	switch assetType {
	case AssetLintStrict:
		if language == "go" {
			filename = "golangci-strict.yml"
		}
	case AssetLintSecurity:
		if language == "go" {
			filename = "lint-security.yml"
		}
	case AssetLintComplexity:
		if language == "go" {
			filename = "lint-complexity.yml"
		}
	case AssetLintStyle:
		if language == "go" {
			filename = "lint-style.yml"
		}
	default:
		return nil, fmt.Errorf("unknown asset type: %s", assetType)
	}

	if filename == "" {
		return nil, fmt.Errorf("no asset defined for language '%s' and type '%s'", language, assetType)
	}

	fullPath := path.Join("assets", language, filename)

	content, err := assetsFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded asset '%s': %w", fullPath, err)
	}

	return content, nil
}
