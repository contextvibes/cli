package config_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// MockExecutor for FindRepoRootConfigPath tests.
type mockExecutor struct {
	CaptureOutputFunc func(ctx context.Context, dir string, commandName string, args ...string) (string, string, error)
}

func (m *mockExecutor) CaptureOutput(
	ctx context.Context,
	dir string,
	commandName string,
	args ...string,
) (string, string, error) {
	if m.CaptureOutputFunc != nil {
		return m.CaptureOutputFunc(ctx, dir, commandName, args...)
	}

	return "", "", errors.New("CaptureOutputFunc not implemented in mock")
}

func (m *mockExecutor) Execute(
	_ context.Context,
	_ string,
	_ string,
	_ ...string,
) error {
	return errors.New("Execute not implemented in mock")
}

func (m *mockExecutor) CommandExists(_ string) bool {
	return false
}

func (m *mockExecutor) Logger() *slog.Logger {
	return nil
}

func TestGetDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := config.GetDefaultConfig()
	require.NotNil(t, cfg)

	assert.Equal(t, config.DefaultGitRemote, cfg.Git.DefaultRemote)
	assert.Equal(t, config.DefaultGitMainBranch, cfg.Git.DefaultMainBranch)
	assert.Equal(t, config.UltimateDefaultAILogFilename, cfg.Logging.DefaultAILogFile)

	require.NotNil(t, cfg.Validation.BranchName.Enable)
	assert.True(t, *cfg.Validation.BranchName.Enable)
	assert.Equal(t, config.DefaultBranchNamePattern, cfg.Validation.BranchName.Pattern)

	require.NotNil(t, cfg.Validation.CommitMessage.Enable)
	assert.True(t, *cfg.Validation.CommitMessage.Enable)
	assert.Equal(t, config.DefaultCommitMessagePattern, cfg.Validation.CommitMessage.Pattern)

	require.NotNil(t, cfg.ProjectState.StrategicKickoffCompleted)
	assert.False(t, *cfg.ProjectState.StrategicKickoffCompleted)
	assert.Empty(t, cfg.ProjectState.LastStrategicKickoffDate)

	// Check AI Collaboration Preferences defaults
	assert.Equal(t, "bash_cat_eof", cfg.AI.CollaborationPreferences.CodeProvisioningStyle)
	assert.Equal(t, "raw_markdown", cfg.AI.CollaborationPreferences.MarkdownDocsStyle)
	assert.Equal(t, "mode_b", cfg.AI.CollaborationPreferences.DetailedTaskMode)
	assert.Equal(t, "detailed_explanations", cfg.AI.CollaborationPreferences.ProactiveDetailLevel)
	assert.Equal(t, "proactive_suggestions", cfg.AI.CollaborationPreferences.AIProactivity)
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	t.Run("file does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.LoadConfig(filepath.Join(tempDir, "nonexistent.yaml"))
		require.NoError(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		emptyFilePath := filepath.Join(tempDir, "empty.yaml")
		require.NoError(t, os.WriteFile(emptyFilePath, []byte{}, 0o600))
		cfg, err := config.LoadConfig(emptyFilePath)
		require.NoError(t, err)
		assert.Nil(t, cfg, "LoadConfig with empty file should return nil config and no error")
	})

	t.Run("malformed YAML", func(t *testing.T) {
		t.Parallel()

		malformedFilePath := filepath.Join(tempDir, "malformed.yaml")
		malformedYAMLContent := "git: { defaultRemote: origin\n  defaultMainBranch: \"not_closed_string"
		require.NoError(t, os.WriteFile(malformedFilePath, []byte(malformedYAMLContent), 0o600))

		cfg, err := config.LoadConfig(malformedFilePath)
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(
			t,
			err.Error(),
			fmt.Sprintf("failed to parse YAML config file '%s'", malformedFilePath),
		)

		var yamlErr *yaml.TypeError

		isYamlError := strings.Contains(err.Error(), "yaml:")
		assert.True(
			t,
			isYamlError || errors.As(err, &yamlErr),
			"Error should be or wrap a YAML parsing error",
		)
	})

	t.Run("valid YAML", func(t *testing.T) {
		t.Parallel()

		validFilePath := filepath.Join(tempDir, "valid.yaml")
		validYAML := `
git:
  defaultRemote: "upstream"
  defaultMainBranch: "develop"
logging:
  defaultAILogFile: "custom_ai.log"
validation:
  branchName:
    enable: false
    pattern: "custom_branch_pattern"
  commitMessage:
    enable: true
    pattern: "custom_commit_pattern"
projectState:
  strategicKickoffCompleted: true
  lastStrategicKickoffDate: "2024-01-01T10:00:00Z"
ai:
  collaborationPreferences:
    codeProvisioningStyle: "raw_markdown"
    detailedTaskMode: "mode_a"
`
		require.NoError(t, os.WriteFile(validFilePath, []byte(validYAML), 0o600))
		cfg, err := config.LoadConfig(validFilePath)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "upstream", cfg.Git.DefaultRemote)
		assert.Equal(t, "develop", cfg.Git.DefaultMainBranch)
		assert.Equal(t, "custom_ai.log", cfg.Logging.DefaultAILogFile)
		require.NotNil(t, cfg.Validation.BranchName.Enable)
		assert.False(t, *cfg.Validation.BranchName.Enable)
		assert.Equal(t, "custom_branch_pattern", cfg.Validation.BranchName.Pattern)
		require.NotNil(t, cfg.Validation.CommitMessage.Enable)
		assert.True(t, *cfg.Validation.CommitMessage.Enable)
		assert.Equal(t, "custom_commit_pattern", cfg.Validation.CommitMessage.Pattern)
		require.NotNil(t, cfg.ProjectState.StrategicKickoffCompleted)
		assert.True(t, *cfg.ProjectState.StrategicKickoffCompleted)
		assert.Equal(t, "2024-01-01T10:00:00Z", cfg.ProjectState.LastStrategicKickoffDate)
		assert.Equal(t, "raw_markdown", cfg.AI.CollaborationPreferences.CodeProvisioningStyle)
		assert.Equal(t, "mode_a", cfg.AI.CollaborationPreferences.DetailedTaskMode)
	})
}

//nolint:funlen // Test function length is acceptable.
func TestMergeWithDefaults(t *testing.T) {
	t.Parallel()

	defaults := config.GetDefaultConfig()
	require.NotNil(t, defaults, "GetDefaultConfig returned nil")

	t.Run("loaded config is nil", func(t *testing.T) {
		t.Parallel()

		merged := config.MergeWithDefaults(nil, defaults)
		assert.Equal(t, defaults, merged)
	})

	t.Run("loaded config with no overrides", func(t *testing.T) {
		t.Parallel()
		//nolint:exhaustruct // Testing partial config.
		loaded := &config.Config{}
		merged := config.MergeWithDefaults(loaded, defaults)
		assert.Equal(t, defaults.Git.DefaultRemote, merged.Git.DefaultRemote)
		assert.Equal(t, defaults.Logging.DefaultAILogFile, merged.Logging.DefaultAILogFile)
		require.NotNil(t, merged.Validation.BranchName.Enable)
		assert.True(t, *merged.Validation.BranchName.Enable)
		assert.Equal(
			t,
			defaults.Validation.BranchName.Pattern,
			merged.Validation.BranchName.Pattern,
		)
		require.NotNil(t, merged.ProjectState.StrategicKickoffCompleted)
		assert.False(t, *merged.ProjectState.StrategicKickoffCompleted)
		assert.Equal(
			t,
			defaults.AI.CollaborationPreferences.CodeProvisioningStyle,
			merged.AI.CollaborationPreferences.CodeProvisioningStyle,
		)
	})

	t.Run("partial git override", func(t *testing.T) {
		t.Parallel()
		//nolint:exhaustruct // Testing partial config.
		loaded := &config.Config{Git: config.GitSettings{DefaultRemote: "myfork"}}
		merged := config.MergeWithDefaults(loaded, defaults)
		assert.Equal(t, "myfork", merged.Git.DefaultRemote)
		assert.Equal(t, defaults.Git.DefaultMainBranch, merged.Git.DefaultMainBranch)
	})

	t.Run("disable branch validation", func(t *testing.T) {
		t.Parallel()

		disableValidation := false
		//nolint:exhaustruct // Testing partial config.
		loaded := &config.Config{Validation: struct {
			BranchName    config.ValidationRule `yaml:"branchName,omitempty"`
			CommitMessage config.ValidationRule `yaml:"commitMessage,omitempty"`
		}{
			BranchName: config.ValidationRule{
				Enable:  &disableValidation,
				Pattern: "should_be_ignored_if_disabled",
			},
		}}
		merged := config.MergeWithDefaults(loaded, defaults)
		require.NotNil(t, merged.Validation.BranchName.Enable)
		assert.False(t, *merged.Validation.BranchName.Enable)
		assert.Empty(
			t,
			merged.Validation.BranchName.Pattern,
			"Pattern should be cleared if validation is disabled",
		)
	})

	t.Run("override one AI collaboration preference", func(t *testing.T) {
		t.Parallel()
		//nolint:exhaustruct // Testing partial config.
		loaded := &config.Config{
			AI: config.AISettings{
				CollaborationPreferences: config.AICollaborationPreferences{
					DetailedTaskMode: "mode_a",
				},
			},
		}
		merged := config.MergeWithDefaults(loaded, defaults)
		assert.Equal(t, "mode_a", merged.AI.CollaborationPreferences.DetailedTaskMode)
		assert.Equal(
			t,
			defaults.AI.CollaborationPreferences.CodeProvisioningStyle,
			merged.AI.CollaborationPreferences.CodeProvisioningStyle,
		)
		assert.Equal(
			t,
			defaults.AI.CollaborationPreferences.AIProactivity,
			merged.AI.CollaborationPreferences.AIProactivity,
		)
	})

	t.Run("override strategicKickoffCompleted to true", func(t *testing.T) {
		t.Parallel()

		trueVal := true
		//nolint:exhaustruct // Testing partial config.
		loaded := &config.Config{
			ProjectState: config.ProjectState{StrategicKickoffCompleted: &trueVal},
		}
		merged := config.MergeWithDefaults(loaded, defaults)
		require.NotNil(t, merged.ProjectState.StrategicKickoffCompleted)
		assert.True(t, *merged.ProjectState.StrategicKickoffCompleted)
		assert.Equal(
			t,
			defaults.ProjectState.LastStrategicKickoffDate,
			merged.ProjectState.LastStrategicKickoffDate,
		)
	})
}

func TestUpdateAndSaveConfig(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	validConfig := config.GetDefaultConfig()
	validConfig.Git.DefaultRemote = "test_remote"

	t.Run("save nil config", func(t *testing.T) {
		t.Parallel()

		err := config.UpdateAndSaveConfig(nil, filepath.Join(tempDir, "nil_config.yaml"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot save a nil config to file")
	})

	t.Run("successful save new file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(tempDir, "new_config.yaml")
		err := config.UpdateAndSaveConfig(validConfig, filePath)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		assert.NoError(t, err, "Config file should exist after saving")

		loaded, loadErr := config.LoadConfig(filePath)
		require.NoError(t, loadErr)
		require.NotNil(t, loaded)
		assert.Equal(t, "test_remote", loaded.Git.DefaultRemote)
	})

	t.Run("successful overwrite existing file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(tempDir, "overwrite_config.yaml")
		initialCfg := config.GetDefaultConfig()
		initialData, _ := yaml.Marshal(initialCfg)
		require.NoError(t, os.WriteFile(filePath, initialData, 0o600))

		updatedCfg := config.GetDefaultConfig()
		updatedCfg.Logging.DefaultAILogFile = "overwrite.log"
		err := config.UpdateAndSaveConfig(updatedCfg, filePath)
		require.NoError(t, err)

		loaded, loadErr := config.LoadConfig(filePath)
		require.NoError(t, loadErr)
		require.NotNil(t, loaded)
		assert.Equal(t, "overwrite.log", loaded.Logging.DefaultAILogFile)
		assert.Equal(t, initialCfg.Git.DefaultRemote, loaded.Git.DefaultRemote)
	})
}

func TestFindRepoRootConfigPath(t *testing.T) {
	t.Parallel()

	t.Run("git rev-parse fails", func(t *testing.T) {
		t.Parallel()

		mockExec := &mockExecutor{
			CaptureOutputFunc: func(ctxIn context.Context, _ string, commandName string, args ...string) (string, string, error) {
				// Note: ctxIn might be different due to context wrapping, so we don't assert equality.
				if commandName == "git" && args[0] == "rev-parse" {
					return "", "git error", errors.New("git rev-parse failed")
				}

				return "", "", errors.New("unexpected command")
			},
		}
		client := exec.NewClient(mockExec)
		_, err := config.FindRepoRootConfigPath(client)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to determine git repository root")
	})

	t.Run("git rev-parse returns empty", func(t *testing.T) {
		t.Parallel()

		mockExec := &mockExecutor{
			CaptureOutputFunc: func(_ context.Context, _ string, commandName string, args ...string) (string, string, error) {
				if commandName == "git" && args[0] == "rev-parse" {
					return "  ", "", nil
				}

				return "", "", errors.New("unexpected command")
			},
		}
		client := exec.NewClient(mockExec)
		_, err := config.FindRepoRootConfigPath(client)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "returned an empty or invalid path")
	})

	t.Run("config file not found in repo root", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		mockExec := &mockExecutor{
			CaptureOutputFunc: func(_ context.Context, _ string, commandName string, args ...string) (string, string, error) {
				if commandName == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
					return tempDir + "\n", "", nil
				}

				return "", "", errors.New("unexpected command")
			},
		}
		client := exec.NewClient(mockExec)

		configPath, err := config.FindRepoRootConfigPath(client)
		require.NoError(t, err)
		assert.Empty(t, configPath, "Should return empty path if config file not found")
	})

	t.Run("config file found in repo root", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		expectedConfigPath := filepath.Join(tempDir, config.DefaultConfigFileName)
		require.NoError(t, os.WriteFile(expectedConfigPath, []byte("git: {}"), 0o600))

		mockExec := &mockExecutor{
			CaptureOutputFunc: func(_ context.Context, _ string, commandName string, args ...string) (string, string, error) {
				if commandName == "git" && args[0] == "rev-parse" && args[1] == "--show-toplevel" {
					return tempDir + "\n", "", nil
				}

				return "", "", errors.New("unexpected command")
			},
		}
		client := exec.NewClient(mockExec)

		configPath, err := config.FindRepoRootConfigPath(client)
		require.NoError(t, err)
		assert.Equal(t, expectedConfigPath, configPath)
	})
}
