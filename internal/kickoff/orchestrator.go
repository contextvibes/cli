// Package kickoff provides the orchestration logic for the ContextVibes CLI's
// 'kickoff' command, handling both strategic project kickoffs and daily
// Git workflow startups.
package kickoff

import (
	"bytes" // For text/template
	"context"
	_ "embed" // Required for //go:embed
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template" // Required for template processing
	"time"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
)

var assumeYes bool

const strategicKickoffPromptFilename = "STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md"

//go:embed assets/strategic_kickoff_protocol_template.md
var masterKickoffProtocolTemplateContent string // Content will be embedded here

type Orchestrator struct {
	logger         *slog.Logger
	config         *config.Config
	presenter      *ui.Presenter
	gitClient      *git.GitClient
	configFilePath string
}

func NewOrchestrator(
	logger *slog.Logger,
	cfg *config.Config,
	presenter *ui.Presenter,
	gitClient *git.GitClient,
	configFilePath string,
) *Orchestrator {
	if logger == nil {
		fmt.Fprintln(os.Stderr, "[WARN] Kickoff Orchestrator initialized with a nil logger. Using discard logger.")
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Orchestrator{
		logger:         logger.With("component", "kickoff.Orchestrator"),
		config:         cfg,
		presenter:      presenter,
		gitClient:      gitClient,
		configFilePath: configFilePath,
	}
}

func (o *Orchestrator) ExecuteKickoff(ctx context.Context, isStrategicFlag bool, branchNameFlag string, globalAssumeYes bool) error {
	assumeYes = globalAssumeYes
	o.logger.DebugContext(ctx, "ExecuteKickoff called",
		slog.Bool("isStrategicFlag", isStrategicFlag),
		slog.String("branchNameFlag", branchNameFlag),
		slog.Bool("assumeYesGlobal", assumeYes))

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
		errMsg := fmt.Sprintf("failed to save updated configuration to %s", o.configFilePath)
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
		return fmt.Errorf("phase IV (Collaboration Setup for CLI) failed: %w", err)
	}
	initialProjectInfo, errGathering := o.runInitialInfoGathering(ctx)
	if errGathering != nil {
		return fmt.Errorf("phase I (Initial Info Gathering for CLI) failed: %w", errGathering)
	}
	if err := o.runTechnicalReadinessInquiry(ctx); err != nil {
		return fmt.Errorf("phase II (Technical Readiness Inquiry for CLI) failed: %w", err)
	}

	o.presenter.Newline()
	o.presenter.Step("Generating the Master Kickoff Prompt for your AI assistant...")

	promptText, err := o.generateMasterKickoffPromptText(initialProjectInfo)
	if err != nil {
		// Error already logged by generateMasterKickoffPromptText
		o.presenter.Error("Failed to generate the master kickoff prompt content: %v", err)
		return err
	}

	promptFilePath := filepath.Join(".", strategicKickoffPromptFilename)
	err = os.WriteFile(promptFilePath, []byte(promptText), 0644)
	if err != nil {
		o.presenter.Error("Failed to save the master kickoff prompt to '%s': %v", promptFilePath, err)
		o.logger.ErrorContext(ctx, "Failed to write master kickoff prompt file", "path", promptFilePath, "error", err)
		return fmt.Errorf("failed to save master kickoff prompt: %w", err)
	}
	o.presenter.Success("Master Kickoff Prompt successfully generated and saved to: %s", o.presenter.Highlight(promptFilePath))
	o.logger.InfoContext(ctx, "Master kickoff prompt generated and saved.", "path", promptFilePath)

	o.presenter.Newline()
	o.presenter.Header("Next Steps for Your Strategic Kickoff:")
	o.presenter.Info("1. Open the generated file: %s", o.presenter.Highlight(promptFilePath))
	o.presenter.Info("2. Copy its entire content.")
	o.presenter.Info("3. Paste it as the initial prompt to your preferred AI assistant (e.g., Gemini, Claude, ChatGPT).")
	o.presenter.Info("4. The AI will then guide you through the detailed strategic kickoff process.")
	o.presenter.Advice("   During your session with the AI, it may ask you to run specific `contextvibes` commands")
	o.presenter.Advice("   (like `describe` or `status`) in your terminal. Provide the output back to the AI.")
	o.presenter.Newline()
	o.presenter.Info("After completing the session with your AI and creating a project summary:")
	o.presenter.Advice(" - Consider saving any AI-generated YAML configurations to your '.contextvibes.yaml' or other project files.")
	o.presenter.Advice(" - You can then inform ContextVibes by running: `contextvibes kickoff --mark-strategic-complete`.")

	o.logger.InfoContext(ctx, "Strategic Project Kickoff prompt generation finished successfully.")
	return nil
}

func (o *Orchestrator) runCollaborationSetup(ctx context.Context) error {
	o.logger.DebugContext(ctx, "Executing Phase IV: Collaboration Model Setup (for ContextVibes CLI interaction)")
	o.presenter.Header("IV. ContextVibes CLI Interaction Preferences")
	o.presenter.Info("Let's set preferences for how I, ContextVibes CLI, will provide outputs like code or docs if generated based on your AI session summary later.")
	o.presenter.Detail("Your choices can be stored in '%s' if you complete a strategic kickoff and this CLI is enhanced to save them.", o.configFilePath)
	o.presenter.Newline()

	if o.config == nil { return errors.New("config is nil in runCollaborationSetup") }
	prefs := &o.config.AI.CollaborationPreferences

	askAndSetPreference := func(promptKey, currentValue string, validOptions []string, exampleValue string, updateFunc func(string)) error {
		var currentDisplay string
		if currentValue == "" { currentDisplay = "(default or not set)"
		} else { currentDisplay = fmt.Sprintf("current: %s", currentValue) }
		optionsStr := strings.Join(validOptions, " | ")
		promptMsgStr := fmt.Sprintf("%s\n  Options: [ %s ] (e.g., '%s')\n  (%s)\nYour choice (or press Enter to keep current/default): ",
			promptKey, optionsStr, exampleValue, currentDisplay)

		userInput, err := o.presenter.PromptForInput(promptMsgStr)
		if err != nil { return fmt.Errorf("failed to get user input for '%s': %w", promptKey, err) }
		userInput = strings.TrimSpace(userInput)

		if userInput == "" {
			o.presenter.Info("  -> No change for '%s'. Using: '%s'", promptKey, currentValue)
			o.logger.DebugContext(ctx, "Collaboration preference unchanged by user", "key", promptKey, "value", currentValue)
			return nil
		}
		isValidOption := false
		for _, opt := range validOptions { if userInput == opt { isValidOption = true; break } }
		if !isValidOption {
			o.presenter.Warning("  -> Invalid option '%s' for '%s'. Kept: '%s'.", userInput, promptKey, currentValue)
			o.logger.WarnContext(ctx, "Invalid collaboration preference option entered", "key", promptKey, "entered_value", userInput, "kept_value", currentValue)
			return nil 
		}
		updateFunc(userInput)
		o.presenter.Info("  -> Preference for '%s' set to: '%s'", promptKey, userInput)
		o.logger.InfoContext(ctx, "Collaboration preference updated", "key", promptKey, "new_value", userInput)
		return nil
	}
	
	var err error
	err = askAndSetPreference("1. CLI Code/Command Provisioning Style:", prefs.CodeProvisioningStyle, []string{"bash_cat_eof", "raw_markdown"}, "bash_cat_eof", func(s string) { prefs.CodeProvisioningStyle = s })
	if err != nil { return err }
	err = askAndSetPreference("2. CLI Markdown Doc Presentation Style:", prefs.MarkdownDocsStyle, []string{"raw_markdown"}, "raw_markdown", func(s string) { prefs.MarkdownDocsStyle = s })
	if err != nil { return err }
	
	o.presenter.Newline()
	o.presenter.Success("ContextVibes CLI interaction preferences noted (will be saved if strategic kickoff is marked complete).")
	o.logger.InfoContext(ctx, "ContextVibes CLI collaboration model setup phase complete.", slog.Any("final_preferences_in_config_struct", *prefs))
	return nil
}

func (o *Orchestrator) runInitialInfoGathering(ctx context.Context) (map[string]string, error) {
	o.logger.DebugContext(ctx, "Executing Simplified Initial Info Gathering for Prompt Parameterization")
	o.presenter.Header("Project Details for Kickoff Prompt Customization")
	o.presenter.Info("I need a few details to customize the master kickoff prompt for your AI.")
	o.presenter.Newline()

	gatheredInfo := make(map[string]string)
	var err error
	var tempStr string

	tempStr, err = o.presenter.PromptForInput("  What is the official name for this project? (e.g., MyNewService): ")
	if err != nil { return nil, err }
	if tempStr == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil { 
			o.logger.ErrorContext(ctx, "Failed to get current working directory for default project name", slog.Any("error", cwdErr))
			return nil, fmt.Errorf("failed to get current directory for default project name: %w", cwdErr)
		}
		gatheredInfo["projectName"] = filepath.Base(cwd)
	} else {
		gatheredInfo["projectName"] = tempStr
	}
	
	tempStr, err = o.presenter.PromptForInput(fmt.Sprintf("  What is the primary application type being developed for '%s'? (e.g., Go API, Python CLI): ", gatheredInfo["projectName"]))
	if err != nil { return nil, err }
	if tempStr == "" { return nil, errors.New("project application type cannot be empty") }
	gatheredInfo["projectAppType"] = tempStr

	tempStr, err = o.presenter.PromptForInput(fmt.Sprintf("  Is '%s' a brand new project, or an existing one entering a new phase? [new/existing]: ", gatheredInfo["projectName"]))
	if err != nil { return nil, err }
	if tempStr == "" { return nil, errors.New("project stage (new/existing) cannot be empty") }
	gatheredInfo["projectStage"] = tempStr
	
	o.presenter.Newline()
	o.logger.InfoContext(ctx, "Simplified initial info gathered.", slog.Any("info", gatheredInfo))
	return gatheredInfo, nil
}

func (o *Orchestrator) runTechnicalReadinessInquiry(ctx context.Context) error {
	o.logger.DebugContext(ctx, "Executing Phase II: Technical Readiness Inquiry (for ContextVibes CLI user)")
	o.presenter.Header("II. ContextVibes CLI Readiness Check")
	o.presenter.Info("Just a couple of quick checks regarding your `contextvibes` CLI environment:")
	o.presenter.Newline()

	var err error
	_, err = o.presenter.PromptForConfirmation("  1. Is the `contextvibes` CLI installed and accessible in your system PATH? [Y/n]: ")
	if err != nil { return err }

	_, err = o.presenter.PromptForConfirmation("  2. Are any environment variables required by `contextvibes` for its general operation set correctly? [Y/n]: ")
	if err != nil { return err }
	o.presenter.Newline()
	
	o.logger.InfoContext(ctx, "Technical readiness inquiry phase (for CLI user) finished.")
	return nil
}

// generateCollaborationPrefsYAML is a helper to create the YAML string for AI collaboration preferences.
func (o *Orchestrator) generateCollaborationPrefsYAML() string {
    // Ensure defaults if some preferences are empty, to make the YAML more complete for the AI prompt.
    prefs := o.config.AI.CollaborationPreferences
    
    style := prefs.CodeProvisioningStyle
    if style == "" { style = "bash_cat_eof" } // Example default
    
    mdStyle := prefs.MarkdownDocsStyle
    if mdStyle == "" { mdStyle = "raw_markdown" }

    taskMode := prefs.DetailedTaskMode
    if taskMode == "" { taskMode = "mode_b" }

    detailLevel := prefs.ProactiveDetailLevel
    if detailLevel == "" && taskMode == "mode_b" { detailLevel = "detailed_explanations" }
	 if detailLevel == "" { detailLevel = "concise_unless_asked" }


    proactivity := prefs.AIProactivity
    if proactivity == "" { proactivity = "proactive_suggestions" }

    var sb strings.Builder
    sb.WriteString("ai:\n")
    sb.WriteString("  collaborationPreferences:\n")
    sb.WriteString(fmt.Sprintf("    codeProvisioningStyle: \"%s\"\n", style))
    sb.WriteString(fmt.Sprintf("    markdownDocsStyle: \"%s\"\n", mdStyle))
    sb.WriteString(fmt.Sprintf("    detailedTaskMode: \"%s\"\n", taskMode))
    sb.WriteString(fmt.Sprintf("    proactiveDetailLevel: \"%s\"\n", detailLevel))
    sb.WriteString(fmt.Sprintf("    aiProactivity: \"%s\"\n", proactivity))
    return sb.String()
}

func (o *Orchestrator) generateMasterKickoffPromptText(initialInfo map[string]string) (string, error) {
	projectName := initialInfo["projectName"]
	if projectName == "" { projectName = "[User to Specify Project Name]" }
	projectAppType := initialInfo["projectAppType"]
	if projectAppType == "" { projectAppType = "[User to Specify Project Type]" }

	collaborationPrefsYAML := o.generateCollaborationPrefsYAML()

	templateData := struct {
		ProjectName            string
		ProjectAppType         string
		CollaborationPrefsYAML string
	}{
		ProjectName:            projectName,
		ProjectAppType:         projectAppType,
		CollaborationPrefsYAML: collaborationPrefsYAML,
	}

	// Ensure masterKickoffProtocolTemplateContent is not empty
	if strings.TrimSpace(masterKickoffProtocolTemplateContent) == "" {
		errMsg := "embedded master kickoff protocol template is empty"
		o.logger.ErrorContext(context.Background(), errMsg) // Use a background context if ctx not available
		return "", errors.New(errMsg)
	}

	tmpl, err := template.New("kickoffProtocol").Parse(masterKickoffProtocolTemplateContent)
	if err != nil {
		o.logger.Error("Failed to parse master kickoff protocol template", slog.Any("error", err))
		return "", fmt.Errorf("failed to parse master kickoff protocol template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.Execute(&outputBuffer, templateData); err != nil {
		o.logger.Error("Failed to execute master kickoff protocol template", slog.Any("error", err))
		return "", fmt.Errorf("failed to execute master kickoff protocol template: %w", err)
	}
	
	return outputBuffer.String(), nil
}

func (o *Orchestrator) executeDailyKickoff(ctx context.Context, branchNameFlag string) error {
	o.presenter.Summary("Starting Daily Git Kickoff workflow.")
	o.logger.InfoContext(ctx, "Starting Daily Git Kickoff workflow.")

	if o.gitClient == nil {
		errMsg := "git client not available for daily kickoff. Ensure you are in a Git repository" 
		o.presenter.Error(errMsg)
		o.logger.ErrorContext(ctx, errMsg)
		return errors.New(errMsg)
	}

	mainBranchName := o.config.Git.DefaultMainBranch
	if mainBranchName == "" { mainBranchName = config.DefaultGitMainBranch } 
	remoteName := o.config.Git.DefaultRemote
	if remoteName == "" { remoteName = config.DefaultGitRemote } 
	
	o.presenter.Info("Using remote '%s' and main branch '%s'.", remoteName, mainBranchName)
	o.presenter.Info("Checking Git prerequisites for daily kickoff...")

	currentBranch, err := o.gitClient.GetCurrentBranchName(ctx)
	if err != nil { o.presenter.Error("Failed to get current Git branch: %v", err); return err }
	if currentBranch != mainBranchName {
		errMsg := fmt.Sprintf("not on the main branch ('%s'). Current branch: '%s'", mainBranchName, currentBranch)
		o.presenter.Error(errMsg)
		o.presenter.Advice("Switch to the main branch first using `git switch %s`.", mainBranchName)
		return errors.New(errMsg)
	}
	o.presenter.Info("Confirmed on main branch '%s'.", mainBranchName)

	isClean, err := o.gitClient.IsWorkingDirClean(ctx)
	if err != nil { o.presenter.Error("Failed checking working directory status: %v", err); return err }
	if !isClean {
		errMsg := "working directory is not clean. Daily kickoff requires a clean state on main"
		o.presenter.Error(errMsg)
		o.presenter.Advice("Commit or stash changes first. Try `contextvibes commit -m \"...\"` or `git stash`.")
		return errors.New(errMsg)
	}
	o.presenter.Info("Working directory is clean.")

	targetBranchName := strings.TrimSpace(branchNameFlag)
	branchValidationRule := o.config.Validation.BranchName
	validationIsEnabled := branchValidationRule.Enable == nil || *branchValidationRule.Enable
	effectivePattern := branchValidationRule.Pattern
	
	if validationIsEnabled && effectivePattern == "" {
		effectivePattern = config.DefaultBranchNamePattern 
		o.logger.DebugContext(ctx, "Branch name validation enabled, using default pattern due to empty config pattern", "pattern", effectivePattern)
	}

	if targetBranchName == "" {
		if assumeYes { 
			errMsg := "branch name is required via --branch flag when using --yes for daily kickoff"
			o.presenter.Error(errMsg)
			o.logger.ErrorContext(ctx, errMsg)
			return errors.New(errMsg)
		}
		o.presenter.Newline()
		o.presenter.Info("Please provide the name for the new daily/feature branch.")
		if validationIsEnabled {
			o.presenter.Advice("Pattern to match: %s", effectivePattern)
		}
		for {
			targetBranchName, err = o.presenter.PromptForInput("New branch name: ")
			if err != nil { return err }
			targetBranchName = strings.TrimSpace(targetBranchName)
			if targetBranchName != "" { break }
			o.presenter.Warning("Branch name cannot be empty.")
		}
	}

	if validationIsEnabled {
		branchNameRe, errRe := regexp.Compile(effectivePattern)
		if errRe != nil {
			o.presenter.Error("Internal error: Invalid branch name validation pattern ('%s') in configuration: %v", effectivePattern, errRe)
			return fmt.Errorf("invalid branch name regex in config: %w", errRe)
		}
		if !branchNameRe.MatchString(targetBranchName) {
			errMsg := fmt.Sprintf("invalid branch name: '%s'", targetBranchName) 
			o.presenter.Error(errMsg)
			o.presenter.Advice("Branch name must match the configured pattern: %s", effectivePattern)
			return errors.New(errMsg)
		}
		o.logger.DebugContext(ctx, "Branch name format validated successfully", "pattern", effectivePattern, "branch", targetBranchName)
	} else {
		o.logger.InfoContext(ctx, "Branch name validation is disabled by configuration.")
	}
	
	existsLocally, err := o.gitClient.LocalBranchExists(ctx, targetBranchName)
	if err != nil { o.presenter.Error("Failed checking if branch '%s' exists locally: %v", targetBranchName, err); return err }
	if existsLocally {
		o.presenter.Error("Branch '%s' already exists locally.", targetBranchName)
		return errors.New("branch already exists locally for daily kickoff")
	}
	
	o.presenter.Newline()
	o.presenter.Info("Proposed Daily Kickoff Actions:")
	o.presenter.Detail("1. Update main branch '%s' from remote '%s' (git pull --rebase).", mainBranchName, remoteName)
	o.presenter.Detail("2. Create and switch to new local branch '%s' from '%s'.", targetBranchName, mainBranchName)
	o.presenter.Detail("3. Push new branch '%s' to '%s' and set upstream tracking.", targetBranchName, remoteName)
	o.presenter.Newline()

	if !assumeYes {
		confirmed, promptErr := o.presenter.PromptForConfirmation("Proceed with Daily Kickoff Git workflow?")
		if promptErr != nil { return promptErr }
		if !confirmed {
			o.presenter.Info("Daily kickoff aborted by user.")
			return nil
		}
	} else {
		o.presenter.Info("Confirmation for Daily Kickoff Git workflow bypassed via --yes flag.")
	}

	o.presenter.Newline()
	o.presenter.Step("Updating main branch '%s' from '%s'...", mainBranchName, remoteName)
	if err := o.gitClient.PullRebase(ctx, mainBranchName); err != nil { o.presenter.Error("Failed to update main branch: %v", err); return err }
	o.presenter.Info("Main branch update successful.")

	o.presenter.Newline()
	o.presenter.Step("Creating and switching to new branch '%s' from '%s'...", targetBranchName, mainBranchName)
	if err := o.gitClient.CreateAndSwitchBranch(ctx, targetBranchName, mainBranchName); err != nil { o.presenter.Error("Failed to create/switch branch: %v", err); return err }
	o.presenter.Info("Successfully on new branch '%s'.", targetBranchName)

	o.presenter.Newline()
	o.presenter.Step("Pushing new branch '%s' to '%s' and setting upstream...", targetBranchName, remoteName)
	if err := o.gitClient.PushAndSetUpstream(ctx, targetBranchName); err != nil { o.presenter.Error("Failed to push new branch: %v", err); return err }
	o.presenter.Info("New branch '%s' pushed and upstream tracking set.", targetBranchName)
	
	o.presenter.Newline()
	o.presenter.Success("Daily Git Kickoff workflow completed. You are now on branch '%s'.", targetBranchName)
	o.logger.InfoContext(ctx, "Daily Git Kickoff workflow finished successfully.", slog.String("new_branch", targetBranchName))
	return nil
}

