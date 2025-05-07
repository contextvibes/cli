// internal/config/config.go
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec" // Import the new exec client
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFileName             = ".contextvibes.yaml"
	DefaultBranchNamePattern          = `^((feature|fix|docs|format)/.+)$`
	DefaultCommitMessagePattern     = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_]+\))?:\s.+`
	DefaultGitRemote                  = "origin"
	DefaultGitMainBranch              = "main"
	UltimateDefaultAILogFilename = "contextvibes_ai_trace.log"
)

type GitSettings struct {
	DefaultRemote     string `yaml:"defaultRemote,omitempty"`
	DefaultMainBranch string `yaml:"defaultMainBranch,omitempty"`
}

type ValidationRule struct {
	Enable  *bool  `yaml:"enable,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`
}

type LoggingSettings struct {
	DefaultAILogFile string `yaml:"defaultAILogFile,omitempty"`
}

type Config struct {
	Git        GitSettings     `yaml:"git,omitempty"`
	Logging    LoggingSettings `yaml:"logging,omitempty"`
	Validation struct {
		BranchName    ValidationRule `yaml:"branchName,omitempty"`
		CommitMessage ValidationRule `yaml:"commitMessage,omitempty"`
	} `yaml:"validation,omitempty"`
}

func GetDefaultConfig() *Config {
	enableTrue := true
	return &Config{
		Git: GitSettings{
			DefaultRemote:     DefaultGitRemote,
			DefaultMainBranch: DefaultGitMainBranch,
		},
		Logging: LoggingSettings{
			DefaultAILogFile: UltimateDefaultAILogFilename,
		},
		Validation: struct {
			BranchName    ValidationRule `yaml:"branchName,omitempty"`
			CommitMessage ValidationRule `yaml:"commitMessage,omitempty"`
		}{
			BranchName: ValidationRule{
				Enable:  &enableTrue,
				Pattern: DefaultBranchNamePattern,
			},
			CommitMessage: ValidationRule{
				Enable:  &enableTrue,
				Pattern: DefaultCommitMessagePattern,
			},
		},
	}
}

func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", filePath, err)
	}
	return &cfg, nil
}

// FindRepoRootConfigPath now takes an ExecutorClient to find the repo root.
func FindRepoRootConfigPath(execClient *exec.ExecutorClient) (string, error) {
	if execClient == nil {
		return "", fmt.Errorf("executor client is nil, cannot find repo root")
	}
	// Use the passed ExecutorClient to find the repo root.
	// The CWD for CaptureOutput can be "." as rev-parse works from any subdirectory.
	// Using context.Background() for this setup-time, non-cancellable operation.
	ctx := context.Background()
	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		// Error from CaptureOutput likely includes context (command, exit code, stderr)
		return "", fmt.Errorf("failed to determine git repository root (is this a git repo, or is 'git' not in PATH? details: %s): %w", strings.TrimSpace(stderr), err)
	}
	repoRoot := filepath.Clean(strings.TrimSpace(stdout))
	if repoRoot == "" || repoRoot == "." { // Should not happen if rev-parse succeeded
		return "", fmt.Errorf("git rev-parse --show-toplevel returned an empty or invalid path: '%s'", repoRoot)
	}

	configPath := filepath.Join(repoRoot, DefaultConfigFileName)
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return "", nil // Config file not found at the discovered root path, not an error for this func
	} else if statErr != nil {
		return "", fmt.Errorf("error checking for config file at '%s': %w", configPath, statErr)
	}

	return configPath, nil
}

func MergeWithDefaults(loadedCfg *Config, defaultConfig *Config) *Config {
	if loadedCfg == nil {
		return defaultConfig
	}
	finalCfg := *defaultConfig

	if loadedCfg.Git.DefaultRemote != "" {
		finalCfg.Git.DefaultRemote = loadedCfg.Git.DefaultRemote
	}
	if loadedCfg.Git.DefaultMainBranch != "" {
		finalCfg.Git.DefaultMainBranch = loadedCfg.Git.DefaultMainBranch
	}
	if loadedCfg.Logging.DefaultAILogFile != "" {
		finalCfg.Logging.DefaultAILogFile = loadedCfg.Logging.DefaultAILogFile
	}
	if loadedCfg.Validation.BranchName.Enable != nil {
		finalCfg.Validation.BranchName.Enable = loadedCfg.Validation.BranchName.Enable
	}
	if (finalCfg.Validation.BranchName.Enable == nil || *finalCfg.Validation.BranchName.Enable) && loadedCfg.Validation.BranchName.Pattern != "" {
		// This logic had a slight flaw, it should use loadedCfg pattern if present
		if loadedCfg.Validation.BranchName.Pattern != "" {
			finalCfg.Validation.BranchName.Pattern = loadedCfg.Validation.BranchName.Pattern
		} // else keep the default pattern already in finalCfg
	} else if finalCfg.Validation.BranchName.Enable != nil && !*finalCfg.Validation.BranchName.Enable {
		finalCfg.Validation.BranchName.Pattern = ""
	}

	if loadedCfg.Validation.CommitMessage.Enable != nil {
		finalCfg.Validation.CommitMessage.Enable = loadedCfg.Validation.CommitMessage.Enable
	}
	if (finalCfg.Validation.CommitMessage.Enable == nil || *finalCfg.Validation.CommitMessage.Enable) && loadedCfg.Validation.CommitMessage.Pattern != "" {
		// Same potential flaw correction here
		if loadedCfg.Validation.CommitMessage.Pattern != "" {
			finalCfg.Validation.CommitMessage.Pattern = loadedCfg.Validation.CommitMessage.Pattern
		} // else keep default
	} else if finalCfg.Validation.CommitMessage.Enable != nil && !*finalCfg.Validation.CommitMessage.Enable {
		finalCfg.Validation.CommitMessage.Pattern = ""
	}
	return &finalCfg
}
