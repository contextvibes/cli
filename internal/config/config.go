// Package config manages the application configuration.
package config

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigFileName is the name of the configuration file.
	DefaultConfigFileName = ".contextvibes.yaml"
	// DefaultCodemodFilename is the default name for codemod scripts.
	DefaultCodemodFilename = "codemod.json"
	// DefaultDescribeOutputFile is the default output for the describe command.
	DefaultDescribeOutputFile = "contextvibes.md"
	// StrategicKickoffFilename is the path to the generated kickoff protocol.
	StrategicKickoffFilename = "docs/strategic_kickoff_protocol.md"
	// DefaultBranchNamePattern is the default regex for branch validation.
	DefaultBranchNamePattern = `^((feature|fix|docs|format)/.+)$`
	// DefaultCommitMessagePattern is the default regex for commit message validation.
	//nolint:lll // Regex pattern is long by necessity.
	DefaultCommitMessagePattern = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+`
	// DefaultGitRemote is the default git remote name.
	DefaultGitRemote = "origin"
	// DefaultGitMainBranch is the default main branch name.
	DefaultGitMainBranch = "main"
	// UltimateDefaultAILogFilename is the fallback log file name.
	UltimateDefaultAILogFilename = "contextvibes_ai_trace.log"

	dirPermUserRWX = 0o750
)

var (
	// ErrEmptyConfig is returned when the configuration file is empty.
	ErrEmptyConfig = errors.New("config file is empty")
	// ErrNoExecutor is returned when the executor client is nil.
	ErrNoExecutor = errors.New("executor client is nil, cannot find repo root")
	// ErrNotGitRepo is returned when the current directory is not a git repository.
	ErrNotGitRepo = errors.New("git rev-parse --show-toplevel returned an empty or invalid path, not in a git repository")
	// ErrNilConfigSave is returned when attempting to save a nil config.
	ErrNilConfigSave = errors.New("cannot save a nil config to file")
)

// GitSettings configures git behavior.
type GitSettings struct {
	DefaultRemote     string `yaml:"defaultRemote,omitempty"`
	DefaultMainBranch string `yaml:"defaultMainBranch,omitempty"`
}

// ValidationRule defines a validation rule with an enable flag and a regex pattern.
type ValidationRule struct {
	Enable  *bool  `yaml:"enable,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`
}

// LoggingSettings configures application logging.
type LoggingSettings struct {
	Enable *bool `yaml:"enable,omitempty"`
	//nolint:tagliatelle // Keep camelCase for backward compatibility.
	DefaultAILogFile string `yaml:"defaultAILogFile,omitempty"`
}

// SystemPromptSettings configures system prompt generation.
type SystemPromptSettings struct {
	DefaultOutputFiles map[string]string `yaml:"defaultOutputFiles,omitempty"`
}

// ProjectState tracks the state of project workflows.
type ProjectState struct {
	StrategicKickoffCompleted *bool  `yaml:"strategicKickoffCompleted,omitempty"`
	LastStrategicKickoffDate  string `yaml:"lastStrategicKickoffDate,omitempty"`
}

// AICollaborationPreferences defines how the AI should interact with the user.
type AICollaborationPreferences struct {
	CodeProvisioningStyle string `yaml:"codeProvisioningStyle,omitempty"`
	MarkdownDocsStyle     string `yaml:"markdownDocsStyle,omitempty"`
	DetailedTaskMode      string `yaml:"detailedTaskMode,omitempty"`
	ProactiveDetailLevel  string `yaml:"proactiveDetailLevel,omitempty"`
	AIProactivity         string `yaml:"aiProactivity,omitempty"`
}

// AISettings groups AI-related settings.
type AISettings struct {
	CollaborationPreferences AICollaborationPreferences `yaml:"collaborationPreferences,omitempty"`
}

// VerificationCheck defines a command to run to verify an example.
type VerificationCheck struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Command     string   `yaml:"command"`
	Args        []string `yaml:"args,omitempty"`
}

// ExampleSettings configures settings for a specific example.
type ExampleSettings struct {
	Verify []VerificationCheck `yaml:"verify,omitempty"`
}

// RunSettings configures the 'run' command.
type RunSettings struct {
	Examples map[string]ExampleSettings `yaml:"examples,omitempty"`
}

// ExportSettings configures the 'export' command.
type ExportSettings struct {
	ExcludePatterns []string `yaml:"excludePatterns,omitempty"`
}

// DescribeSettings configures the 'describe' command.
type DescribeSettings struct {
	IncludePatterns []string `yaml:"includePatterns,omitempty"`
	ExcludePatterns []string `yaml:"excludePatterns,omitempty"`
}

// ProjectSettings configures project-wide settings.
type ProjectSettings struct {
	Provider        string   `yaml:"provider,omitempty"`
	UpstreamModules []string `yaml:"upstreamModules,omitempty"`
}

// BehaviorSettings configures general CLI behavior.
type BehaviorSettings struct {
	DualOutput bool `yaml:"dualOutput,omitempty"`
}

// FeedbackSettings configures the 'feedback' command.
type FeedbackSettings struct {
	DefaultRepository string            `yaml:"defaultRepository,omitempty"`
	Repositories      map[string]string `yaml:"repositories,omitempty"`
}

// Config is the top-level configuration structure.
type Config struct {
	Git          GitSettings          `yaml:"git,omitempty"`
	Logging      LoggingSettings      `yaml:"logging,omitempty"`
	SystemPrompt SystemPromptSettings `yaml:"systemPrompt,omitempty"`
	Validation   struct {
		BranchName    ValidationRule `yaml:"branchName,omitempty"`
		CommitMessage ValidationRule `yaml:"commitMessage,omitempty"`
	} `yaml:"validation,omitempty"`
	ProjectState ProjectState     `yaml:"projectState,omitempty"`
	AI           AISettings       `yaml:"ai,omitempty"`
	Run          RunSettings      `yaml:"run,omitempty"`
	Export       ExportSettings   `yaml:"export,omitempty"`
	Describe     DescribeSettings `yaml:"describe,omitempty"`
	Project      ProjectSettings  `yaml:"project,omitempty"`
	Behavior     BehaviorSettings `yaml:"behavior,omitempty"`
	Feedback     FeedbackSettings `yaml:"feedback,omitempty"`
}

// GetDefaultConfig returns a Config struct populated with default values.
//
//nolint:funlen // Initialization of large struct requires length.
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
		Describe: DescribeSettings{
			IncludePatterns: []string{
				//nolint:lll // Regex patterns are long.
				`\.(go|mod|sum|tf|py|yaml|yml|json|md|gitignore|txt|hcl|nix|js|html|css|sql|sqlx|sh|rb|java|c|cpp|h|hpp|rs|toml|xml|proto)$`,
				`^(Dockerfile|Makefile|Taskfile\.yml|requirements\.txt|README\.md|\.idx/dev\.nix|\.idx/airules\.md)$`,
			},
			ExcludePatterns: []string{
				//nolint:lll // Regex patterns are long.
				`(^vendor/|^\.git/|^\.terraform/|^\.venv/|^__pycache__/|^\.DS_Store|^\.pytest_cache/|^\.vscode/|node_modules/|dist/|build/)`,
				`(\.tfstate|\.tfplan|^secrets?/|\.auto\.tfvars|ai_context\.txt|crash.*\.log|contextvibes\.md)$`,
				`\.(exe|bin|dll|so|jar|class|o|a|zip|tar\.gz|rar|7z|jpg|jpeg|png|gif|svg|ico|woff|woff2|ttf|eot)$`,
			},
		},
		Project: ProjectSettings{
			Provider:        "github",
			UpstreamModules: nil,
		},
		Behavior: BehaviorSettings{
			DualOutput: true,
		},
		Feedback: FeedbackSettings{
			DefaultRepository: "cli",
			Repositories: map[string]string{
				"cli":  "contextvibes/cli",
				"thea": "contextvibes/thea",
			},
		},
	}

	return cfg
}

// LoadConfig attempts to load configuration from the specified file path.
func LoadConfig(filePath string) (*Config, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		//nolint:nilnil // Returning nil, nil is valid for optional config.
		return nil, nil
	}

	//nolint:gosec // Reading config file is intended behavior.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}

	if len(data) == 0 {
		return nil, ErrEmptyConfig
	}

	var cfg Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML config file '%s': %w", filePath, err)
	}

	return &cfg, nil
}

// FindRepoRootConfigPath attempts to locate the config file in the git repo root.
func FindRepoRootConfigPath(execClient *exec.ExecutorClient) (string, error) {
	if execClient == nil {
		return "", ErrNoExecutor
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
		return "", ErrNotGitRepo
	}

	configPath := filepath.Join(repoRoot, DefaultConfigFileName)

	_, statErr := os.Stat(configPath)
	if os.IsNotExist(statErr) {
		return "", nil
	} else if statErr != nil {
		return "", fmt.Errorf("error checking for config file at '%s': %w", configPath, statErr)
	}

	return configPath, nil
}

// MergeWithDefaults merges a loaded config with the default config.
//
//nolint:gocognit,gocyclo,cyclop,funlen // Complexity and length are due to many fields to check.
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

	if loadedCfg.Describe.IncludePatterns != nil {
		finalCfg.Describe.IncludePatterns = loadedCfg.Describe.IncludePatterns
	}

	if loadedCfg.Describe.ExcludePatterns != nil {
		finalCfg.Describe.ExcludePatterns = loadedCfg.Describe.ExcludePatterns
	}

	if loadedCfg.Project.Provider != "" {
		finalCfg.Project.Provider = loadedCfg.Project.Provider
	}

	if loadedCfg.Project.UpstreamModules != nil {
		finalCfg.Project.UpstreamModules = loadedCfg.Project.UpstreamModules
	}

	if loadedCfg.Behavior.DualOutput != defaultConfig.Behavior.DualOutput {
		finalCfg.Behavior.DualOutput = loadedCfg.Behavior.DualOutput
	}

	if loadedCfg.Feedback.DefaultRepository != "" {
		finalCfg.Feedback.DefaultRepository = loadedCfg.Feedback.DefaultRepository
	}

	if loadedCfg.Feedback.Repositories != nil {
		maps.Copy(finalCfg.Feedback.Repositories, loadedCfg.Feedback.Repositories)
	}

	return &finalCfg
}

// UpdateAndSaveConfig writes the configuration to the specified file path.
func UpdateAndSaveConfig(cfgToSave *Config, filePath string) error {
	if cfgToSave == nil {
		return ErrNilConfigSave
	}

	yamlData, err := yaml.Marshal(cfgToSave)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML for saving: %w", err)
	}

	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, dirPermUserRWX)
		if err != nil {
			return fmt.Errorf("failed to create directory for config file '%s': %w", dir, err)
		}
	}

	tempFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file in '%s': %w", dir, err)
	}

	defer func() { _ = os.Remove(tempFile.Name()) }()

	_, err = tempFile.Write(yamlData)
	if err != nil {
		_ = tempFile.Close()

		return fmt.Errorf("failed to write to temporary config file '%s': %w", tempFile.Name(), err)
	}

	err = tempFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close temporary config file '%s': %w", tempFile.Name(), err)
	}

	err = os.Rename(tempFile.Name(), filePath)
	if err != nil {
		return fmt.Errorf("failed to rename temporary config file to '%s': %w", filePath, err)
	}

	return nil
}
