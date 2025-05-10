// Package kickoff provides the orchestration logic for the ContextVibes CLI's
// 'kickoff' command, handling both strategic project kickoffs and daily
// Git workflow startups.
package kickoff

import (
	"bytes" 
	"context"
	_ "embed" 
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template" 
	"time"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
)

const strategicKickoffPromptFilename = "STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md"

//go:embed assets/strategic_kickoff_protocol_template.md
var masterKickoffProtocolTemplateContent string 

type Orchestrator struct {
	logger         *slog.Logger
	config         *config.Config
	presenter      *ui.Presenter
	gitClient      *git.GitClient
	configFilePath string
	assumeYes      bool 
}

func NewOrchestrator(
	logger *slog.Logger,
	cfg *config.Config,
	presenter *ui.Presenter,
	gitClient *git.GitClient,
	configFilePath string,
	globalAssumeYes bool, 
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
	
	// Save config *after* collaboration preferences are gathered, so they are included
	// in the generated YAML for the master prompt.
	if err := config.UpdateAndSaveConfig(o.config, o.configFilePath); err != nil {
		o.presenter.Warning("Could not immediately save AI collaboration preferences to '%s': %v", o.configFilePath, err)
		o.logger.WarnContext(ctx, "Failed to save config after collaboration setup in strategic kickoff generation", "error", err, "path", o.configFilePath)
	} else {
		o.logger.InfoContext(ctx, "AI collaboration preferences (in-memory) saved to config file.", "path", o.configFilePath)
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
	o.presenter.Info("Let's set preferences for how I, ContextVibes CLI, will provide outputs like code or docs.")
	o.presenter.Detail("Your choices help tailor my assistance and will be saved in '%s' if you mark this strategic kickoff as complete.", o.configFilePath)
	o.presenter.Newline()

	if o.config == nil { 
		o.logger.ErrorContext(ctx, "Config is nil in runCollaborationSetup")
		return errors.New("internal error: config not loaded for collaboration setup")
	}
	prefs := &o.config.AI.CollaborationPreferences 
	defaultCP := config.GetDefaultConfig().AI.CollaborationPreferences


	askAndSetPreference := func(promptKey, currentValue string, validOptions []string, defaultValue string, updateFunc func(string)) error {
		var currentDisplay string
		if currentValue == "" { 
			currentDisplay = fmt.Sprintf("(current default: %s)", defaultValue)
		} else { 
			currentDisplay = fmt.Sprintf("current: %s", currentValue) 
		}
		optionsStr := strings.Join(validOptions, " | ")
		promptMsgStr := fmt.Sprintf("%s\n  Options: [ %s ]\n  %s\nYour choice (or press Enter to use value shown in current/default): ",
			promptKey, optionsStr, currentDisplay)

		userInput, err := o.presenter.PromptForInput(promptMsgStr)
		if err != nil { return fmt.Errorf("failed to get user input for '%s': %w", promptKey, err) }
		userInput = strings.TrimSpace(userInput)

		effectiveValue := currentValue
		if effectiveValue == "" { effectiveValue = defaultValue}


		if userInput == "" {
			updateFunc(effectiveValue) // Ensure a value (current or default) is set in the config struct
			o.presenter.Info("  -> No change for '%s'. Using: '%s'", promptKey, effectiveValue)
			o.logger.DebugContext(ctx, "Collaboration preference confirmed/defaulted", "key", promptKey, "value", effectiveValue)
			return nil
		}
		isValidOption := false
		for _, opt := range validOptions { if userInput == opt { isValidOption = true; break } }
		
		if !isValidOption {
			o.presenter.Warning("  -> Invalid option '%s' for '%s'. Valid: [%s].", userInput, promptKey, optionsStr)
			o.presenter.Info("  Please try again for '%s'.", promptKey)
			userInputRetry, errRetry := o.presenter.PromptForInput(fmt.Sprintf("Retry for '%s' (Options: %s): ", promptKey, optionsStr))
			if errRetry != nil { return fmt.Errorf("failed to get user input on retry for '%s': %w", promptKey, errRetry) }
			userInput = strings.TrimSpace(userInputRetry)
			isValidOption = false 
			for _, opt := range validOptions { if userInput == opt { isValidOption = true; break } }
			if !isValidOption {
				o.presenter.Error("  -> Still invalid option '%s' for '%s'. Keeping: '%s'.", userInput, promptKey, effectiveValue)
				updateFunc(effectiveValue) 
				o.logger.WarnContext(ctx, "Invalid collaboration preference on retry", "key", promptKey, "attemptedValue", userInput, "kept_value", effectiveValue)
				return nil 
			}
		}
		updateFunc(userInput)
		o.presenter.Info("  -> Preference for '%s' set to: '%s'", promptKey, userInput)
		o.logger.InfoContext(ctx, "Collaboration preference updated", "key", promptKey, "new_value", userInput)
		return nil
	}
	
	var err error
	err = askAndSetPreference("1. My Code/Command Provisioning Style:", prefs.CodeProvisioningStyle, []string{"bash_cat_eof", "raw_markdown"}, defaultCP.CodeProvisioningStyle, func(s string) { prefs.CodeProvisioningStyle = s })
	if err != nil { return err }

	err = askAndSetPreference("2. My Markdown Doc Presentation Style:", prefs.MarkdownDocsStyle, []string{"raw_markdown"}, defaultCP.MarkdownDocsStyle, func(s string) { prefs.MarkdownDocsStyle = s })
	if err != nil { return err }

	err = askAndSetPreference("3. My Detailed Task Interaction Model (for future direct AI):", prefs.DetailedTaskMode, []string{"mode_a", "mode_b"}, defaultCP.DetailedTaskMode, func(s string) { prefs.DetailedTaskMode = s })
	if err != nil { return err }
	
	effectiveTaskMode := prefs.DetailedTaskMode; if effectiveTaskMode == "" {effectiveTaskMode = defaultCP.DetailedTaskMode}
	defaultProactiveDetail := defaultCP.ProactiveDetailLevel
	if effectiveTaskMode == "mode_b" { defaultProactiveDetail = "detailed_explanations"
	} else if defaultProactiveDetail == "" { defaultProactiveDetail = "concise_unless_asked" }


	err = askAndSetPreference("4. My Explanation Depth & Proactive Detail Level:", prefs.ProactiveDetailLevel, []string{"detailed_explanations", "concise_unless_asked"}, defaultProactiveDetail, func(s string) { prefs.ProactiveDetailLevel = s })
	if err != nil { return err }
	
	err = askAndSetPreference("5. My General Proactivity in Suggestions:", prefs.AIProactivity, []string{"proactive_suggestions", "wait_for_request"}, defaultCP.AIProactivity, func(s string) { prefs.AIProactivity = s })
	if err != nil { return err }

	o.presenter.Newline() 
	o.presenter.Info("Informational Points (not saved as preferences):")
	o.presenter.Detail("  - Context Management: For very long *chat sessions* with an AI, you might need to refresh its context.")
	o.presenter.Detail("  - Feedback: Your feedback on ContextVibes (and me!) is always welcome.")
	o.presenter.Detail("  - AI Rules: If `.idx/airules.md` exists, I use its guidance. Any session-specific overrides can be mentioned now.")
	_, _ = o.presenter.PromptForInput("  Any immediate thoughts on AI rules for this session? (Enter or describe): ") 
	o.presenter.Newline()

	o.presenter.Success("ContextVibes CLI interaction preferences configured (in memory).")
	o.logger.InfoContext(ctx, "Collaboration model setup phase complete.", slog.Any("final_preferences_in_config_struct", *prefs))
	return nil
}

func (o *Orchestrator) runInitialInfoGathering(ctx context.Context) (map[string]string, error) {
	o.logger.DebugContext(ctx, "Executing Simplified Initial Info Gathering for Prompt Parameterization")
	o.presenter.Header("I. Initial Project Information for Kickoff Prompt")
	o.presenter.Info("Please provide these initial details about the project you are kicking off.")
	o.presenter.Newline()

	gatheredInfo := make(map[string]string)
	var err error
	var tempStr string

	tempStr, err = o.presenter.PromptForInput("  1. What is the official name for this project?\n     (e.g., MyNewService. Press Enter to use current directory name): ")
	if err != nil { return nil, err }
	if tempStr == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil { 
			o.logger.ErrorContext(ctx, "Failed to get current working directory for default project name", slog.Any("error", cwdErr))
			// Allow user to manually enter if auto-detection fails
			tempStr, err = o.presenter.PromptForInput("  Could not get directory name. Please enter project name: ")
			if err != nil { return nil, err }
			if tempStr == "" { return nil, errors.New("project name cannot be empty if directory name fetch fails") }
		} else {
			gatheredInfo["projectName"] = filepath.Base(cwd)
			o.presenter.Info("     -> Using current directory name: %s", gatheredInfo["projectName"])
		}
	}
	if tempStr != "" { // If user entered a name, or re-entered after CWD fail
		gatheredInfo["projectName"] = tempStr
	}
	o.presenter.Newline()
	
	tempStr, err = o.presenter.PromptForInput(fmt.Sprintf("  2. What is the primary application type for '%s'?\n     (e.g., Go API, Python CLI, Terraform Module): ", gatheredInfo["projectName"]))
	if err != nil { return nil, err }
	if tempStr == "" { return nil, errors.New("project application type cannot be empty") }
	gatheredInfo["projectAppType"] = tempStr
	o.presenter.Newline()

	tempStr, err = o.presenter.PromptForInput(fmt.Sprintf("  3. Is '%s' a [new] project, an [existing] project (new phase), or a [refactor] effort? ", gatheredInfo["projectName"]))
	if err != nil { return nil, err }
	if tempStr == "" { return nil, errors.New("project stage (new/existing/refactor) cannot be empty") }
	gatheredInfo["projectStage"] = tempStr
	o.presenter.Newline()
	
	o.logger.InfoContext(ctx, "Simplified initial info gathered for prompt parameterization.", slog.Any("info", gatheredInfo))
	return gatheredInfo, nil
}

func (o *Orchestrator) runTechnicalReadinessInquiry(ctx context.Context) error {
	o.logger.DebugContext(ctx, "Executing Phase II: Technical Readiness Inquiry (for User's ContextVibes CLI setup)")
	o.presenter.Header("II. User's ContextVibes CLI Readiness Check")
	o.presenter.Info("A quick check regarding your local `contextvibes` CLI environment:")
	o.presenter.Newline()

	var err error
	var confirmed bool

	confirmed, err = o.presenter.PromptForConfirmation("  1. Is your `contextvibes` CLI installed and accessible in your system PATH? [Y/n]: ")
	if err != nil { return err }
	if !confirmed && !o.assumeYes { o.presenter.Warning("     -> Please ensure `contextvibes` is installed and in your PATH for optimal use during the AI-guided session.") }
	o.presenter.Newline()

	confirmed, err = o.presenter.PromptForConfirmation("  2. Are any environment variables required by `contextvibes` for its general operation set correctly? [Y/n]: ")
	if err != nil { return err }
	if !confirmed && !o.assumeYes { o.presenter.Advice("     -> Ensure relevant environment variables are set if ContextVibes features you use depend on them.") }
	o.presenter.Newline()
	
	o.presenter.Success("Technical readiness inquiry complete.")
	o.logger.InfoContext(ctx, "Technical readiness inquiry phase (for CLI user) finished.")
	return nil
}

func (o *Orchestrator) generateCollaborationPrefsYAML() string {
    defaultPrefs := config.GetDefaultConfig().AI.CollaborationPreferences
    currentPrefs := o.config.AI.CollaborationPreferences
    
    getEffective := func(current, def string) string {
        if current != "" { return current }
        return def
    }

    style := getEffective(currentPrefs.CodeProvisioningStyle, defaultPrefs.CodeProvisioningStyle)
    mdStyle := getEffective(currentPrefs.MarkdownDocsStyle, defaultPrefs.MarkdownDocsStyle)
    taskMode := getEffective(currentPrefs.DetailedTaskMode, defaultPrefs.DetailedTaskMode)
    
    detailLevel := currentPrefs.ProactiveDetailLevel
    if detailLevel == "" { // If user didn't explicitly set it
        if taskMode == "mode_b" { // If effective taskMode is mode_b
            detailLevel = "detailed_explanations" // Sensible default for interactive mode
        } else { // For mode_a or if taskMode is somehow empty/defaulted to non-mode_b
            detailLevel = getEffective("", defaultPrefs.ProactiveDetailLevel) // Use general default
			if detailLevel == "" { detailLevel = "concise_unless_asked"; } // Absolute fallback
        }
    }

    proactivity := getEffective(currentPrefs.AIProactivity, defaultPrefs.AIProactivity)

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

	if strings.TrimSpace(masterKickoffProtocolTemplateContent) == "" {
		errMsg := "embedded master kickoff protocol template is empty or not loaded"
		o.logger.ErrorContext(context.Background(), errMsg, slog.String("hint", "Ensure internal/kickoff/assets/strategic_kickoff_protocol_template.md is populated and embed directive is correct."))
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
		if o.assumeYes { 
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

	if !o.assumeYes { 
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

