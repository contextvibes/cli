// cmd/root.go
package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/spf13/cobra"
)

// These are package-level variables, intended to be accessible by all files
// within this 'cmd' package.
var (
	AppLogger       *slog.Logger
	LoadedAppConfig *config.Config
	ExecClient      *exec.ExecutorClient // For general command execution
	assumeYes       bool                 // For --yes flag
	AppVersion      string               // Set in init() or by ldflags
)

// rootCmd is the base for all commands. It's also a package-level variable.
var rootCmd = &cobra.Command{
	Use:   "contextvibes",
	Short: "Manages project tasks: AI context generation, Git workflow, IaC, etc.",
	Long: `ContextVibes: Your Project Co-Pilot CLI.
This tool helps streamline common development tasks by providing consistent wrappers
for Git workflows, Infrastructure as Code (IaC) operations, code quality checks,
and more. It aims for clear, structured terminal output and detailed background
logging suitable for AI consumption.

Use the --yes flag to skip interactive confirmation prompts.
Customizations for branch naming, commit message validation, default Git
settings, and default AI log file name can be placed in a .contextvibes.yaml
file in the project root.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		tempLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

		bootstrapOSExecutor := exec.NewOSCommandExecutor(slog.New(slog.NewTextHandler(io.Discard, nil)))
		bootstrapExecClient := exec.NewClient(bootstrapOSExecutor)

		defaultCfg := config.GetDefaultConfig()
		var loadedUserConfig *config.Config
		var configLoadErr error
		var foundConfigPath string

		repoConfigPath, findPathErr := config.FindRepoRootConfigPath(bootstrapExecClient)
		if findPathErr != nil {
			tempLogger.Debug("Could not find git repo root to look for .contextvibes.yaml, using defaults.", slog.String("error", findPathErr.Error()))
			LoadedAppConfig = defaultCfg
		} else if repoConfigPath == "" {
			tempLogger.Debug(".contextvibes.yaml not found in repository root, using default configuration.")
			LoadedAppConfig = defaultCfg
		} else {
			foundConfigPath = repoConfigPath
			tempLogger.Debug("Attempting to load config file", slog.String("path", foundConfigPath))
			loadedUserConfig, configLoadErr = config.LoadConfig(foundConfigPath)

			if configLoadErr != nil {
				fmt.Fprintf(os.Stderr, "[WARNING] Error loading config file '%s': %v. Using default settings.\n", foundConfigPath, configLoadErr)
				tempLogger.Error("Failed to load or parse .contextvibes.yaml, using defaults.", slog.String("path", foundConfigPath), slog.String("error", configLoadErr.Error()))
				LoadedAppConfig = defaultCfg
			} else if loadedUserConfig == nil {
				tempLogger.Info(".contextvibes.yaml was checked but not found or effectively empty, using default configuration.", slog.String("path_checked", foundConfigPath))
				LoadedAppConfig = defaultCfg
			} else {
				tempLogger.Info("Successfully loaded .contextvibes.yaml.", slog.String("path", foundConfigPath))
				LoadedAppConfig = config.MergeWithDefaults(loadedUserConfig, defaultCfg)
			}
		}

		aiLevel := parseLogLevel(logLevelAIValue, slog.LevelDebug) // Renamed flag variable
		targetAILogFile := LoadedAppConfig.Logging.DefaultAILogFile
		if aiLogFileFlagValue != "" { // Renamed flag variable
			targetAILogFile = aiLogFileFlagValue
		}

		var aiOut io.Writer = io.Discard
		logFileHandle, errLogFile := os.OpenFile(targetAILogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if errLogFile != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to open AI log file '%s': %v. AI logs will be discarded.\n", targetAILogFile, errLogFile)
		} else {
			aiOut = logFileHandle
		}
		aiHandlerOptions := &slog.HandlerOptions{Level: aiLevel}
		aiHandler := slog.NewJSONHandler(aiOut, aiHandlerOptions)
		AppLogger = slog.New(aiHandler)

		mainOSExecutor := exec.NewOSCommandExecutor(AppLogger)
		ExecClient = exec.NewClient(mainOSExecutor)

		AppLogger.Debug("AI Logger and main ExecutorClient initialized",
			slog.String("log_level_set_for_ai_file", aiLevel.String()),
			slog.String("ai_log_file_target", targetAILogFile),
			slog.Bool("ai_log_file_active", aiOut != io.Discard),
		)
		if assumeYes {
			AppLogger.Info("Running in non-interactive mode (--yes specified)")
		}

		if LoadedAppConfig != nil {
			AppLogger.Debug("Effective application configuration resolved",
				slog.Group("config",
					slog.Group("git",
						slog.String("defaultRemote", LoadedAppConfig.Git.DefaultRemote),
						slog.String("defaultMainBranch", LoadedAppConfig.Git.DefaultMainBranch),
					),
					slog.Group("logging",
						slog.String("defaultAILogFile", LoadedAppConfig.Logging.DefaultAILogFile),
					),
					slog.Group("validation",
						slog.Group("branchName",
							slog.Bool("enable", (LoadedAppConfig.Validation.BranchName.Enable != nil && *LoadedAppConfig.Validation.BranchName.Enable) || LoadedAppConfig.Validation.BranchName.Enable == nil),
							slog.String("pattern", LoadedAppConfig.Validation.BranchName.Pattern),
						),
						slog.Group("commitMessage",
							slog.Bool("enable", (LoadedAppConfig.Validation.CommitMessage.Enable != nil && *LoadedAppConfig.Validation.CommitMessage.Enable) || LoadedAppConfig.Validation.CommitMessage.Enable == nil),
							slog.String("pattern", LoadedAppConfig.Validation.CommitMessage.Pattern),
						),
					),
				),
			)
		} else {
			AppLogger.Error("CRITICAL: LoadedAppConfig is unexpectedly nil after initialization attempt.")
		}
		return nil
	},
}

// Execute is the main entry point for the CLI. It's made public so main.go can call it.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if AppLogger != nil {
			AppLogger.Error("CLI execution finished with error", slog.String("error", err.Error()))
		} else {
			fmt.Fprintf(os.Stderr, "[ERROR] CLI execution failed before logger initialization: %v\n", err)
		}
		os.Exit(1)
	}
}

// Flag variables should have distinct names from package-level vars if they are only for binding.
var (
	logLevelAIValue    string
	aiLogFileFlagValue string
)

func init() {
	if AppVersion == "" {
		AppVersion = "v0.0.4"
	}
	// Use different names for flag-bound variables to avoid confusion with package vars
	// that might be intended for direct use.
	rootCmd.PersistentFlags().StringVar(&logLevelAIValue, "log-level-ai", "debug", "AI (JSON) file log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&aiLogFileFlagValue, "ai-log-file", "",
		fmt.Sprintf("AI (JSON) log file path (overrides config default: see .contextvibes.yaml, fallback: %s)", config.UltimateDefaultAILogFilename))
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume 'yes' to all confirmation prompts, enabling non-interactive mode")

	// Subcommands (like versionCmd, kickoffCmd, codemodCmd) add themselves to rootCmd
	// via their own init() functions. This is a standard Cobra pattern.
}

func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	levelStrLower := strings.ToLower(levelStr)
	switch levelStrLower {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		if levelStr != "" && !strings.EqualFold(levelStr, defaultLevel.String()) {
			fmt.Fprintf(os.Stderr, "[WARNING] Invalid AI log level '%s' provided, using default '%s'.\n", levelStr, defaultLevel.String())
		}
		return defaultLevel
	}
}
