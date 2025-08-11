// FILE: internal/kickoff/orchestrator.go
package kickoff

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
)

// PresenterInterface defines the set of methods the orchestrator needs from a presenter.
// This allows for mocking in tests.
//nolint:interfacebloat
type PresenterInterface interface {
	Header(format string, a ...any)
	Summary(format string, a ...any)
	Step(format string, a ...any)
	Info(format string, a ...any)
	Success(format string, a ...any)
	Error(format string, a ...any)
	Warning(format string, a ...any)
	Advice(format string, a ...any)
	Detail(format string, a ...any)
	Highlight(text string) string
	Newline()
	PromptForInput(prompt string) (string, error)
	PromptForConfirmation(prompt string) (bool, error)
}

const strategicKickoffPromptFilename = "STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md"

//go:embed assets/strategic_kickoff_protocol_template.md
var masterKickoffProtocolTemplateContent string

type Orchestrator struct {
	logger         *slog.Logger
	config         *config.Config
	presenter      PresenterInterface
	gitClient      *git.GitClient
	configFilePath string
	assumeYes      bool
}

func NewOrchestrator(
	logger *slog.Logger,
	cfg *config.Config,
	presenter PresenterInterface,
	gitClient *git.GitClient,
	configFilePath string,
	globalAssumeYes bool,
) *Orchestrator {
	if logger == nil {
		fmt.Fprintln(os.Stderr, "[WARN] Kickoff Orchestrator initialized with a nil logger. Using discard logger.")
		logger = slog.New(slog.DiscardHandler)
	}
	return &Orchestrator{
		logger:         logger.With("component", "kickoff.Orchestrator"),
		config:         cfg,
		presenter:      presenter,
		gitClient:      gitClient,
		configFilePath: configFilePath,
		assumeYes:      globalAssumeYes,
	}
}

func (o *Orchestrator) ExecuteKickoff(ctx context.Context, isStrategicFlag bool, branchNameFlag string) error {
	o.logger.DebugContext(ctx, "ExecuteKickoff called",
		slog.Bool("isStrategicFlag", isStrategicFlag),
		slog.String("branchNameFlag", branchNameFlag),
		slog.Bool("assumeYes", o.assumeYes))

	if o.config == nil {
		err := errors.New("orchestrator config is nil, cannot proceed")
		o.logger.ErrorContext(ctx, "Configuration error in ExecuteKickoff", slog.Any("error", err))
		o.presenter.Error("Internal error: Kickoff orchestrator missing essential configuration.")
		return err
	}

	runStrategic := isStrategicFlag
	if !runStrategic {
		if o.config.ProjectState.StrategicKickoffCompleted == nil || !*o.config.ProjectState.StrategicKickoffCompleted {
			o.presenter.Info("No prior strategic kickoff detected for this project.")
			o.presenter.Info("This command will now guide you to generate a master prompt for an AI to facilitate a full Strategic Project Kickoff.")
			o.presenter.Newline()
			runStrategic = true
		}
	}

	if runStrategic {
		return o.executeStrategicKickoffGeneration(ctx)
	}

	return o.executeDailyKickoff(ctx, branchNameFlag)
}

func (o *Orchestrator) MarkStrategicKickoffComplete(ctx context.Context) error {
	o.presenter.Summary("Marking Strategic Project Kickoff as Complete...")
	o.logger.InfoContext(ctx, "Attempting to mark strategic kickoff as complete.")

	if o.config == nil {
		err := errors.New("orchestrator config is nil, cannot mark complete")
		o.logger.ErrorContext(ctx, "Configuration error in MarkStrategicKickoffComplete", slog.Any("error", err))
		o.presenter.Error("Internal error: Configuration not loaded.")
		return err
	}

	trueVal := true
	o.config.ProjectState.StrategicKickoffCompleted = &trueVal
	o.config.ProjectState.LastStrategicKickoffDate = time.Now().UTC().Format(time.RFC3339)

	if err := config.UpdateAndSaveConfig(o.config, o.configFilePath); err != nil {
		errMsg := "failed to save updated configuration to " + o.configFilePath
		o.presenter.Error("%s: %v", errMsg, err)
		o.logger.ErrorContext(ctx, "Failed to save config for marking strategic kickoff complete", "error", err, "path", o.configFilePath)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	o.presenter.Success("Strategic kickoff has been marked as complete in: %s", o.presenter.Highlight(o.configFilePath))
	o.presenter.Info("Subsequent `contextvibes kickoff` runs (without --strategic) will now perform the daily Git workflow.")
	o.logger.InfoContext(ctx, "Strategic kickoff marked as complete.", "path", o.configFilePath)
	return nil
}

func (o *Orchestrator) executeStrategicKickoffGeneration(ctx context.Context) error {
	o.presenter.Summary("ContextVibes: Strategic Project Kickoff - Prompt Generation")
	o.logger.InfoContext(ctx, "Starting Strategic Project Kickoff prompt generation.")

	if err := o.runCollaborationSetup(ctx); err != nil {
		o.presenter.Error("Failed during CLI collaboration setup phase.")
		return fmt.Errorf("phase IV (Collaboration Setup for CLI) failed: %w", err)
	}

	initialProjectInfo, errGathering := o.runInitialInfoGathering(ctx)
	if errGathering != nil {
		o.presenter.Error("Failed during initial project information gathering.")
		return fmt.Errorf("phase I (Initial Info Gathering for CLI) failed: %w", errGathering)
	}

	if err := o.runTechnicalReadinessInquiry(ctx); err != nil {
		o.presenter.Error("Failed during technical readiness inquiry.")
		return fmt.Errorf("phase II (Technical Readiness Inquiry for CLI) failed: %w", err)
	}

	if err := config.UpdateAndSaveConfig(o.config, o.configFilePath); err != nil {
		o.presenter.Warning("Could not immediately save AI collaboration preferences to '%s': %v", o.configFilePath, err)
		o.logger.WarnContext(ctx, "Failed to save config after collaboration setup", "error", err, "path", o.configFilePath)
	} else {
		o.logger.InfoContext(ctx, "AI collaboration preferences saved to config file.", "path", o.configFilePath)
	}

	o.presenter.Newline()
	o.presenter.Step("Generating the Master Kickoff Prompt for your AI assistant...")

	promptText, err := o.generateMasterKickoffPromptText(initialProjectInfo)
	if err != nil {
		o.presenter.Error("Failed to generate the master kickoff prompt content: %v", err)
		return err
	}

	promptFilePath := filepath.Join(".", strategicKickoffPromptFilename)
	err = os.WriteFile(promptFilePath, []byte(promptText), 0644)
	if err != nil {
		o.presenter.Error("Failed to save the master kickoff prompt to '%s': %v", promptFilePath, err)
		return fmt.Errorf("failed to save master kickoff prompt: %w", err)
	}

	o.presenter.Success("Master Kickoff Prompt successfully generated and saved to: %s", o.presenter.Highlight(promptFilePath))
	o.presenter.Newline()
	o.presenter.Header("Next Steps for Your Strategic Kickoff:")
	o.presenter.Info("1. Open the generated file: %s", o.presenter.Highlight(promptFilePath))
	o.presenter.Info("2. Copy its entire content.")
	o.presenter.Info("3. Paste it as the initial prompt to your preferred AI assistant (e.g., Gemini, Claude, ChatGPT).")
	o.presenter.Info("4. The AI will then guide you through the detailed strategic kickoff process.")
	return nil
}

func (o *Orchestrator) runCollaborationSetup(ctx context.Context) error { return nil }
func (o *Orchestrator) runInitialInfoGathering(ctx context.Context) (map[string]string, error) { return nil, nil }
func (o *Orchestrator) runTechnicalReadinessInquiry(ctx context.Context) error { return nil }
func (o *Orchestrator) generateCollaborationPrefsYAML() string { return "" }
func (o *Orchestrator) generateMasterKickoffPromptText(initialInfo map[string]string) (string, error) { return "prompt text", nil }

func (o *Orchestrator) executeDailyKickoff(ctx context.Context, branchNameFlag string) error {
	if o.gitClient == nil {
		o.presenter.Error("git client not available for daily kickoff.")
		return errors.New("git client not available")
	}
	// ... (rest of daily kickoff logic)
	return nil
}