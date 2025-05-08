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
	Long: `ContextVibes: Your Project Development Assistant CLI.

Designed to enhance your development workflow, ContextVibes offers a suite of
commands that bring consistency, automation, and AI-readiness to your daily tasks.

Key Capabilities:
  * Git Workflow Automation: Streamlined commands like 'kickoff', 'commit',
    'sync', 'wrapup', and 'status'. Features configurable validation for
    branch names and commit messages.
  * AI Context Generation: The 'describe' and 'diff' commands produce
    AI-friendly markdown ('contextvibes.md') detailing project state or
    changes, perfect for integrating with large language models.
  * Infrastructure & Code Management: Consistent wrappers for 'plan', 'deploy',
    'init' (IaC for Terraform/Pulumi), 'quality' checks, 'format' (code
    formatting for Go, Python, Terraform), and 'test' (project testing).
  * Programmatic Refactoring: The 'codemod' command allows applying
    structured code modifications from a JSON script.

Output & Logging for Clarity and AI:
  * User-Focused Terminal Output: Employs clear, structured messages with
    semantic prefixes (SUMMARY, INFO, ERROR, ADVICE, +, ~, !) and colors,
    all managed by an internal UI presenter.
  * Detailed AI Trace Log: Generates a separate, comprehensive JSON log
    (default: 'contextvibes_ai_trace.log', configurable) capturing in-depth
    execution details, ideal for AI analysis or advanced debugging.

Global Features for Control & Customization:
  * Non-Interactive Mode: Use the global '--yes' (or '-y') flag to
    automatically confirm prompts, enabling use in scripts and automation.
  * Project-Specific Configuration: Tailor default behaviors such as Git
    remote/main branch names, validation rule patterns (for branches and
    commits), and the default AI log file path using a '.contextvibes.yaml'
    file in your project's root directory.

For detailed information on any command, use 'contextvibes [command] --help'.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Temporary logger for bootstrap phase, before full config is loaded.
		// This logger writes to stderr for messages during the configuration loading process.
		tempLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

		// Minimal executor for finding config file, uses a discard logger
		// to avoid polluting logs before the main logger is set up.
		bootstrapOSExecutor := exec.NewOSCommandExecutor(slog.New(slog.NewTextHandler(io.Discard, nil)))
		bootstrapExecClient := exec.NewClient(bootstrapOSExecutor)

		defaultCfg := config.GetDefaultConfig()
		var loadedUserConfig *config.Config
		var configLoadErr error
		var foundConfigPath string

		// Attempt to find and load config file. The config is loaded from the repository root
		// or defaults to the CLI's internal configuration if none is found or there is an error during loading.
		repoConfigPath, findPathErr := config.FindRepoRootConfigPath(bootstrapExecClient)
		if findPathErr != nil {
			// Log to tempLogger (stderr) if finding path fails. The application will continue with default settings.
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
				// User-facing warning to stderr if config loading fails
				fmt.Fprintf(os.Stderr, "[WARNING] Error loading config file '%s': %v. Using default settings.\n", foundConfigPath, configLoadErr)
				// Log to tempLogger (stderr) for more detail
				tempLogger.Error("Failed to load or parse .contextvibes.yaml, using defaults.", slog.String("path", foundConfigPath), slog.String("error", configLoadErr.Error()))
				LoadedAppConfig = defaultCfg
			} else if loadedUserConfig == nil {
				// Config file path was found, but file was empty or didn't parse to anything
				tempLogger.Info(".contextvibes.yaml was checked but not found or effectively empty, using default configuration.", slog.String("path_checked", foundConfigPath))
				LoadedAppConfig = defaultCfg
			} else {
				// Successfully loaded user config, now merge with defaults
				tempLogger.Info("Successfully loaded .contextvibes.yaml.", slog.String("path", foundConfigPath))
				LoadedAppConfig = config.MergeWithDefaults(loadedUserConfig, defaultCfg)
			}
		}

		// Determine AI log level and file path.
		// 1. Command-line flags take precedence over everything.
		// 2. If no flag is provided, the configuration file settings take precedence over the defaults.
		// 3. If neither a flag nor a config file setting is present, the built-in default values are used.
		aiLevel := parseLogLevel(logLevelAIValue, slog.LevelDebug) // logLevelAIValue is from the flag

		targetAILogFile := LoadedAppConfig.Logging.DefaultAILogFile // From merged config (or default if no user config)
		if aiLogFileFlagValue != "" {                               // aiLogFileFlagValue is from the flag
			targetAILogFile = aiLogFileFlagValue // Flag overrides config
		}

		// Initialize AppLogger (the main AI trace logger).
		var aiOut io.Writer = io.Discard // Default to discard if file opening fails
		logFileHandle, errLogFile := os.OpenFile(targetAILogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if errLogFile != nil {
			// User-facing error if AI log file cannot be opened
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to open AI log file '%s': %v. AI logs will be discarded.\n", targetAILogFile, errLogFile)
		} else {
			aiOut = logFileHandle
			// defer logFileHandle.Close() // This would close too early, needs to be closed on app exit if managed here. Usually handled by OS.
		}
		aiHandlerOptions := &slog.HandlerOptions{Level: aiLevel}
		aiHandler := slog.NewJSONHandler(aiOut, aiHandlerOptions)
		AppLogger = slog.New(aiHandler)

		// Initialize the main ExecutorClient with the now-configured AppLogger.
		mainOSExecutor := exec.NewOSCommandExecutor(AppLogger)
		ExecClient = exec.NewClient(mainOSExecutor)

		// Log initial setup to the now active AppLogger.
		AppLogger.Debug("AI Logger and main ExecutorClient initialized",
			slog.String("log_level_set_for_ai_file", aiLevel.String()),
			slog.String("ai_log_file_target", targetAILogFile),
			slog.Bool("ai_log_file_active", aiOut != io.Discard),
		)
		if assumeYes {
			AppLogger.Info("Running in non-interactive mode (--yes specified)")
		}

		if LoadedAppConfig != nil {
			// Log the effective configuration that the application will use.
			// Correctly determine the effective boolean value for validation settings:
			// Enabled if nil (not set by user, use default=true) OR if set to true by user.
			branchNameValidationEnabled := LoadedAppConfig.Validation.BranchName.Enable == nil || *LoadedAppConfig.Validation.BranchName.Enable
			commitMsgValidationEnabled := LoadedAppConfig.Validation.CommitMessage.Enable == nil || *LoadedAppConfig.Validation.CommitMessage.Enable

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
							slog.Bool("enable", branchNameValidationEnabled),
							slog.String("pattern", LoadedAppConfig.Validation.BranchName.Pattern),
						),
						slog.Group("commitMessage",
							slog.Bool("enable", commitMsgValidationEnabled),
							slog.String("pattern", LoadedAppConfig.Validation.CommitMessage.Pattern),
						),
					),
				),
			)
		} else {
			// This should ideally not happen if logic above is correct
			AppLogger.Error("CRITICAL: LoadedAppConfig is unexpectedly nil after initialization attempt.")
			// Potentially return an error here to prevent CLI from running with no config
			// return errors.New("critical error: application configuration failed to load")
		}
		return nil
	},
}

// Execute is the main entry point for the CLI. It's made public so main.go can call it.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Ensure AppLogger is checked for nil before use, in case Execute() fails very early
		if AppLogger != nil {
			AppLogger.Error("CLI execution finished with error", slog.String("error", err.Error()))
		} else {
			// Fallback to stderr if logger isn't initialized
			fmt.Fprintf(os.Stderr, "[ERROR] CLI execution failed before logger initialization: %v\n", err)
		}
		os.Exit(1)
	}
}

// Flag variables should have distinct names from package-level vars if they are only for binding.
var (
	logLevelAIValue    string // Bound to --log-level-ai flag
	aiLogFileFlagValue string // Bound to --ai-log-file flag
)

func init() {
	// Set the application version. This can be overridden by ldflags during build.
	if AppVersion == "" {
		AppVersion = "v0.0.6" // Default version if not set by build flags
	}

	// Define persistent flags available to all commands.
	// Use different names for flag-bound variables (logLevelAIValue, aiLogFileFlagValue)
	// to avoid confusion with package-level variables that might be intended for direct use or derived values.
	// These flags bind to the variables logLevelAIValue and aiLogFileFlagValue, and take precedence over config file settings.
	rootCmd.PersistentFlags().StringVar(&logLevelAIValue, "log-level-ai", "debug", "AI (JSON) file log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&aiLogFileFlagValue, "ai-log-file", "",
		fmt.Sprintf("AI (JSON) log file path (overrides config default: see .contextvibes.yaml, fallback: %s)", config.UltimateDefaultAILogFilename))
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume 'yes' to all confirmation prompts, enabling non-interactive mode")

	// Subcommands (like versionCmd, kickoffCmd, codemodCmd, etc.) add themselves to rootCmd
	// via their own init() functions. This is a standard Cobra pattern and keeps this file cleaner.
}

// parseLogLevel converts a string log level to an slog.Level.
func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	levelStrLower := strings.ToLower(strings.TrimSpace(levelStr))
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
		// If an invalid level string is provided (and it's not empty/default),
		// print a warning to stderr.
		if levelStr != "" && !strings.EqualFold(levelStr, defaultLevel.String()) {
			fmt.Fprintf(os.Stderr, "[WARNING] Invalid AI log level '%s' provided. Using default level '%s'.\n", levelStr, defaultLevel.String())
		}
		return defaultLevel
	}
}
