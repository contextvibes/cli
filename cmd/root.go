package cmd

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/spf13/cobra"
)

var (
	AppLogger       *slog.Logger
	LoadedAppConfig *config.Config
	ExecClient      *exec.ExecutorClient
	assumeYes       bool
	AppVersion      string
)

var rootCmd = &cobra.Command{
	Use:   "contextvibes",
	Short: "Manages project tasks: AI context generation, Git workflow, IaC, etc.",
	Long:  `ContextVibes: Your Project Development Assistant CLI.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Bootstrap logger and exec client
		bootstrapOSExecutor := exec.NewOSCommandExecutor(slog.New(slog.DiscardHandler))
		bootstrapExecClient := exec.NewClient(bootstrapOSExecutor)

		// Load and merge config
		defaultCfg := config.GetDefaultConfig()
		repoConfigPath, findPathErr := config.FindRepoRootConfigPath(bootstrapExecClient)
		if findPathErr != nil || repoConfigPath == "" {
			LoadedAppConfig = defaultCfg
		} else {
			loadedUserConfig, configLoadErr := config.LoadConfig(repoConfigPath)
			if configLoadErr != nil || loadedUserConfig == nil {
				LoadedAppConfig = defaultCfg
			} else {
				LoadedAppConfig = config.MergeWithDefaults(loadedUserConfig, defaultCfg)
			}
		}

		// Initialize final logger
		aiLevel := parseLogLevel(logLevelAIValue, slog.LevelDebug)
		targetAILogFile := LoadedAppConfig.Logging.DefaultAILogFile
		if aiLogFileFlagValue != "" {
			targetAILogFile = aiLogFileFlagValue
		}
		logFileHandle, errLogFile := os.OpenFile(targetAILogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		var aiOut = io.Discard
		if errLogFile == nil {
			aiOut = logFileHandle
		}
		AppLogger = slog.New(slog.NewJSONHandler(aiOut, &slog.HandlerOptions{Level: aiLevel}))

		// Initialize final exec client
		mainOSExecutor := exec.NewOSCommandExecutor(AppLogger)
		ExecClient = exec.NewClient(mainOSExecutor)

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	logLevelAIValue    string
	aiLogFileFlagValue string
)

func init() {
	if AppVersion == "" {
		AppVersion = "v0.2.0-dev" // Updated version
	}

	rootCmd.PersistentFlags().StringVar(&logLevelAIValue, "log-level-ai", "debug", "AI (JSON) file log level")
	rootCmd.PersistentFlags().StringVar(&aiLogFileFlagValue, "ai-log-file", "", "AI (JSON) log file path")
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume 'yes' to all prompts")

	// --- Register ALL Commands Here ---
	rootCmd.AddCommand(codemodCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(kickoffCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(qualityCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(theaCmd)
	rootCmd.AddCommand(versionCmd)
	// Placeholders for commands to be added
	// rootCmd.AddCommand(NewFinishCmd())
	// rootCmd.AddCommand(NewGitTidyCmd())
	// rootCmd.AddCommand(NewRunCmd())
	// rootCmd.AddCommand(NewUpdateCmd())
}

func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return defaultLevel
	}
}
