// internal/config/config.go
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFileName        = ".contextvibes.yaml"
	DefaultCodemodFilename       = "codemod.json"
	DefaultDescribeOutputFile    = "contextvibes.md"
	StrategicKickoffFilename     = "docs/strategic_kickoff_protocol.md"
	DefaultBranchNamePattern     = `^((feature|fix|docs|format)/.+)$`
	DefaultCommitMessagePattern  = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+`
	DefaultGitRemote             = "origin"
	DefaultGitMainBranch         = "main"
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
	Enable           *bool  `yaml:"enable,omitempty"`
	DefaultAILogFile string `yaml:"defaultAILogFile,omitempty"`
}

type SystemPromptSettings struct {
	DefaultOutputFiles map[string]string `yaml:"defaultOutputFiles,omitempty"`
}

type ProjectState struct {
	StrategicKickoffCompleted *bool  `yaml:"strategicKickoffCompleted,omitempty"`
	LastStrategicKickoffDate  string `yaml:"lastStrategicKickoffDate,omitempty"`
}

type AICollaborationPreferences struct {
	CodeProvisioningStyle string `yaml:"codeProvisioningStyle,omitempty"`
	MarkdownDocsStyle     string `yaml:"markdownDocsStyle,omitempty"`
	DetailedTaskMode      string `yaml:"detailedTaskMode,omitempty"`
	ProactiveDetailLevel  string `yaml:"proactiveDetailLevel,omitempty"`
	AIProactivity         string `yaml:"aiProactivity,omitempty"`
}

type AISettings struct {
	CollaborationPreferences AICollaborationPreferences `yaml:"collaborationPreferences,omitempty"`
}

type VerificationCheck struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Command     string   `yaml:"command"`
	Args        []string `yaml:"args,omitempty"`
}

type ExampleSettings struct {
	Verify []VerificationCheck `yaml:"verify,omitempty"`
}

type RunSettings struct {
	Examples map[string]ExampleSettings `yaml:"examples,omitempty"`
}

type ExportSettings struct {
	ExcludePatterns []string `yaml:"excludePatterns,omitempty"`
}

type Config struct {
	Git          GitSettings          `yaml:"git,omitempty"`
	Logging      LoggingSettings      `yaml:"logging,omitempty"`
	SystemPrompt SystemPromptSettings `yaml:"systemPrompt,omitempty"`
	Validation   struct {
		BranchName    ValidationRule `yaml:"branchName,omitempty"`
		CommitMessage ValidationRule `yaml:"commitMessage,omitempty"`
	} `yaml:"validation,omitempty"`
	ProjectState ProjectState   `yaml:"projectState,omitempty"`
	AI           AISettings     `yaml:"ai,omitempty"`
	Run          RunSettings    `yaml:"run,omitempty"`
	Export       ExportSettings `yaml:"export,omitempty"`
}

func GetDefaultConfig() *Config {
	enableTrue := true
	defaultFalse := false

	cfg := &Config{
		Git: GitSettings{
			DefaultRemote:     DefaultGitRemote,
			DefaultMainBranch: DefaultGitMainBranch,
		},
		Logging: LoggingSettings{
			Enable:           &defaultFalse,
			DefaultAILogFile: UltimateDefaultAILogFilename,
		},
		SystemPrompt: SystemPromptSettings{
			DefaultOutputFiles: map[string]string{
				"idx":      ".idx/airules.md",
				"aistudio": "contextvibes_aistudio_prompt.md",
			},
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
		ProjectState: ProjectState{
			StrategicKickoffCompleted: &defaultFalse,
			LastStrategicKickoffDate:  "",
		},
		AI: AISettings{
			CollaborationPreferences: AICollaborationPreferences{
				CodeProvisioningStyle: "bash_cat_eof",
				MarkdownDocsStyle:     "raw_markdown",
				DetailedTaskMode:      "mode_b",
				ProactiveDetailLevel:  "detailed_explanations",
				AIProactivity:         "proactive_suggestions",
			},
		},
		Run: RunSettings{
			Examples: make(map[string]ExampleSettings),
		},
		Export: ExportSettings{
			ExcludePatterns: []string{"vendor/"},
		},
	}
	return cfg
}

func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML config file '%s': %w", filePath, err)
	}
	return &cfg, nil
}

func FindRepoRootConfigPath(execClient *exec.ExecutorClient) (string, error) {
	if execClient == nil {
		return "", errors.New("executor client is nil, cannot find repo root")
	}
	ctx := context.Background()
	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf(
			"failed to determine git repository root (is this a git repo, or is 'git' not in PATH? details: %s): %w",
			strings.TrimSpace(stderr),
			err,
		)
	}
	repoRoot := filepath.Clean(strings.TrimSpace(stdout))
	if repoRoot == "" || repoRoot == "." {
		return "", errors.New(
			"git rev-parse --show-toplevel returned an empty or invalid path, not in a git repository",
		)
	}

	configPath := filepath.Join(repoRoot, DefaultConfigFileName)
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return "", nil
	} else if statErr != nil {
		return "", fmt.Errorf("error checking for config file at '%s': %w", configPath, statErr)
	}
	return configPath, nil
}

func MergeWithDefaults(loadedCfg *Config, defaultConfig *Config) *Config {
	if defaultConfig == nil {
		panic("MergeWithDefaults called with nil defaultConfig")
	}
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
	if loadedCfg.Logging.Enable != nil {
		finalCfg.Logging.Enable = loadedCfg.Logging.Enable
	}
	if loadedCfg.Logging.DefaultAILogFile != "" {
		finalCfg.Logging.DefaultAILogFile = loadedCfg.Logging.DefaultAILogFile
	}

	// Merge the system prompt config. User's map replaces the default map.
	if loadedCfg.SystemPrompt.DefaultOutputFiles != nil {
		finalCfg.SystemPrompt.DefaultOutputFiles = loadedCfg.SystemPrompt.DefaultOutputFiles
	}

	if loadedCfg.Validation.BranchName.Enable != nil {
		finalCfg.Validation.BranchName.Enable = loadedCfg.Validation.BranchName.Enable
	}
	if finalCfg.Validation.BranchName.Enable == nil || *finalCfg.Validation.BranchName.Enable {
		if loadedCfg.Validation.BranchName.Pattern != "" {
			finalCfg.Validation.BranchName.Pattern = loadedCfg.Validation.BranchName.Pattern
		}
	} else {
		finalCfg.Validation.BranchName.Pattern = ""
	}

	if loadedCfg.Validation.CommitMessage.Enable != nil {
		finalCfg.Validation.CommitMessage.Enable = loadedCfg.Validation.CommitMessage.Enable
	}
	if finalCfg.Validation.CommitMessage.Enable == nil ||
		*finalCfg.Validation.CommitMessage.Enable {
		if loadedCfg.Validation.CommitMessage.Pattern != "" {
			finalCfg.Validation.CommitMessage.Pattern = loadedCfg.Validation.CommitMessage.Pattern
		}
	} else {
		finalCfg.Validation.CommitMessage.Pattern = ""
	}

	if loadedCfg.ProjectState.StrategicKickoffCompleted != nil {
		finalCfg.ProjectState.StrategicKickoffCompleted = loadedCfg.ProjectState.StrategicKickoffCompleted
	}
	if loadedCfg.ProjectState.LastStrategicKickoffDate != "" {
		finalCfg.ProjectState.LastStrategicKickoffDate = loadedCfg.ProjectState.LastStrategicKickoffDate
	}

	userAICollabPrefs := loadedCfg.AI.CollaborationPreferences
	if userAICollabPrefs.CodeProvisioningStyle != "" {
		finalCfg.AI.CollaborationPreferences.CodeProvisioningStyle = userAICollabPrefs.CodeProvisioningStyle
	}
	if userAICollabPrefs.MarkdownDocsStyle != "" {
		finalCfg.AI.CollaborationPreferences.MarkdownDocsStyle = userAICollabPrefs.MarkdownDocsStyle
	}
	if userAICollabPrefs.DetailedTaskMode != "" {
		finalCfg.AI.CollaborationPreferences.DetailedTaskMode = userAICollabPrefs.DetailedTaskMode
	}
	if userAICollabPrefs.ProactiveDetailLevel != "" {
		finalCfg.AI.CollaborationPreferences.ProactiveDetailLevel = loadedCfg.AI.CollaborationPreferences.ProactiveDetailLevel
	}
	if userAICollabPrefs.AIProactivity != "" {
		finalCfg.AI.CollaborationPreferences.AIProactivity = loadedCfg.AI.CollaborationPreferences.AIProactivity
	}

	if loadedCfg.Run.Examples != nil {
		finalCfg.Run.Examples = loadedCfg.Run.Examples
	}
	if loadedCfg.Export.ExcludePatterns != nil {
		finalCfg.Export.ExcludePatterns = loadedCfg.Export.ExcludePatterns
	}

	return &finalCfg
}

func UpdateAndSaveConfig(cfgToSave *Config, filePath string) error {
	if cfgToSave == nil {
		return errors.New("cannot save a nil config to file")
	}

	yamlData, err := yaml.Marshal(cfgToSave)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML for saving: %w", err)
	}

	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory for config file '%s': %w", dir, err)
		}
	}

	tempFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file in '%s': %w", dir, err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if _, err := tempFile.Write(yamlData); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to write to temporary config file '%s': %w", tempFile.Name(), err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary config file '%s': %w", tempFile.Name(), err)
	}
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("failed to rename temporary config file to '%s': %w", filePath, err)
	}
	return nil
}
