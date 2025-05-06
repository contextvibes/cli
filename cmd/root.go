// cmd/root.go

package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// AppLogger is the central logger ONLY for the AI log file.
var AppLogger *slog.Logger

// Flags
var (
	// AI Logging Flags
	logLevelAI string
	aiLogFile  string
	// Non-interactive Flag
	AppVersion string

	//Non-Interactive Flag
	assumeYes bool // Variable to hold the value of the --yes flag
)

const defaultAILogFilename = "contextvibes.log"

var rootCmd = &cobra.Command{
	Use:   "contextvibes",
	Short: "Manages project tasks: AI context generation, Git workflow, IaC, etc.",
	Long: `ContextVibes: Your Project Co-Pilot CLI
(Full description here)

Use the --yes flag to skip interactive confirmation prompts.`, // Added help text
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// --- Initialize AI File Logging ONLY ---
		aiLevel := parseLogLevel(logLevelAI, slog.LevelDebug)

		targetAILogFile := defaultAILogFilename
		aiLogTargetDesc := fmt.Sprintf("default file '%s'", defaultAILogFilename)
		if aiLogFile != "" {
			targetAILogFile = aiLogFile
			aiLogTargetDesc = fmt.Sprintf("specified file '%s'", aiLogFile)
		}

		var aiOut io.Writer = io.Discard
		// var aiLogOpenErr error // No need to store this error currently

		file, err := os.OpenFile(targetAILogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			// aiLogOpenErr = err // No need to store this error currently
			// Log the error directly to stderr as the main logger might not be ready
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to open AI log file (%s): %v. AI logs will be discarded.\n", aiLogTargetDesc, err)
		} else {
			aiOut = file
			// TODO: Consider adding a defer file.Close() here if the file handle needs to be closed cleanly on exit.
			// However, since the logger might be used throughout the application's lifetime,
			// closing it here might be premature. Managing the file handle lifecycle might require
			// a more sophisticated shutdown mechanism if strict resource cleanup is needed.
		}

		// --- NO human logger setup anymore ---

		// Create AI JSON handler (or discard handler)
		var aiHandler slog.Handler
		if aiOut != io.Discard {
			aiHandler = slog.NewJSONHandler(aiOut, &slog.HandlerOptions{Level: aiLevel})
		} else {
			// Use a discard handler if the file couldn't be opened
			aiHandler = slog.NewJSONHandler(io.Discard, nil)
		}
		AppLogger = slog.New(aiHandler)

		AppLogger.Debug("AI Logger initialized",
			slog.String("ai_log_level_set", aiLevel.String()),
			slog.String("ai_log_target_used", targetAILogFile),
			slog.Bool("ai_log_active", aiOut != io.Discard),
		)
		// Log if non-interactive mode is active (useful for AI trace)
		if assumeYes {
			AppLogger.Info("Running in non-interactive mode (--yes specified)")
		}
		return nil
	},
}

// Execute function remains the same
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Attempt to use the AppLogger if initialized, otherwise fallback to stderr
		if AppLogger != nil {
			AppLogger.Error("CLI execution finished with error", slog.String("error", err.Error()))
		} else {
			fmt.Fprintf(os.Stderr, "[ERROR] CLI execution failed: %v (logger not initialized)\n", err)
		}
		// Cobra/application might handle exit code, or os.Exit(1) could be added here if needed.
		// os.Exit(1) // Uncomment if explicit non-zero exit is desired on error
	}
}

func init() {
	AppVersion = "0.0.3"
	// AI Logging Flags
	rootCmd.PersistentFlags().StringVar(&logLevelAI, "log-level-ai", "debug", "AI (JSON) file log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&aiLogFile, "ai-log-file", "", fmt.Sprintf("AI (JSON) log file path (default: %s)", defaultAILogFilename))

	// *** Add the non-interactive flag ***
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume 'yes' to all confirmation prompts")

	// Subcommands add themselves to rootCmd
}
func init(){
	rootCmd.AddCommand(versionCmd)}

// parseLogLevel function remains the same
func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	// Convert input to lower case for case-insensitive comparison
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
		// Only warn if a level was provided but it wasn't valid *and* it's not the default level string representation
		if levelStr != "" && !strings.EqualFold(levelStr, defaultLevel.String()) {
			// Use Fprintf to stderr for consistency, as logger might not be fully set up yet.
			fmt.Fprintf(os.Stderr, "[WARNING] Invalid AI log level '%s', using default '%s'.\n", levelStr, defaultLevel.String())
		}
		return defaultLevel
	}
}
