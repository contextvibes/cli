// Package create provides the command to create new issues.
package create

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed create.md.tpl
var createLongDescription string

//go:embed assets/pbi.md
var pbiTemplate string

//go:embed assets/prompt.md.tpl
var aiPromptTemplateStr string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	issueType  string
	issueTitle string
	issueBody  string
	aiAssist   bool
)

var (
	errEmptyIntent = errors.New("intent cannot be empty")
	errInvalidJSON = errors.New("invalid JSON provided")
)

// aiIssueResponse matches the expected JSON from the AI.
type aiIssueResponse struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Type   string   `json:"type"`
	Labels []string `json:"labels"`
}

// newProvider is a factory function that returns the configured work item provider.
func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		return github.New(ctx, logger, cfg) //nolint:wrapcheck // Factory function.
	case "":
		logger.DebugContext(
			ctx,
			"Work item provider not specified in config, defaulting to 'github'",
		)

		return github.New(ctx, logger, cfg) //nolint:wrapcheck // Factory function.
	default:
		//nolint:err113 // Dynamic error is appropriate here.
		return nil, fmt.Errorf(
			"unsupported work item provider '%s' specified in .contextvibes.yaml",
			cfg.Project.Provider,
		)
	}
}

// CreateCmd represents the project issues create command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new", "add"},
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		// --- AI Assisted Mode ---
		if aiAssist {
			return runAIAssistedCreation(ctx, presenter, provider)
		}

		// --- Standard Interactive/Flag Mode ---
		if issueTitle == "" {
			form := huh.NewForm(
				huh.NewGroup(
					//nolint:lll // Long line for options.
					huh.NewSelect[string]().Title("What kind of issue is this?").
						Options(
							huh.NewOption("Task", "Task"),
							huh.NewOption("Story", "Story"),
							huh.NewOption("Bug", "Bug"),
							huh.NewOption("Chore", "Chore"),
							huh.NewOption("PBI (Rigorous)", "PBI"),
						).
						Value(&issueType),
					huh.NewInput().Title("Title?").Value(&issueTitle),
				),
			)
			err := form.Run()
			if err != nil {
				return fmt.Errorf("input form failed: %w", err)
			}

			// If PBI is selected and body is empty, pre-fill it
			if issueType == "PBI" && issueBody == "" {
				issueBody = pbiTemplate
			}

			// Prompt for body
			bodyForm := huh.NewForm(
				huh.NewGroup(
					huh.NewText().Title("Body?").Value(&issueBody),
				),
			)
			err = bodyForm.Run()
			if err != nil {
				return fmt.Errorf("input form failed: %w", err)
			}
		}

		if issueTitle == "" {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("title cannot be empty")
		}

		// If using flags and type is PBI but no body provided, use template
		if issueType == "PBI" && issueBody == "" {
			issueBody = pbiTemplate
		}

		//nolint:exhaustruct // Partial initialization is valid for creation.
		newItem := workitem.WorkItem{
			Title: issueTitle,
			Body:  issueBody,
			Type:  workitem.Type(issueType),
		}

		return createItem(ctx, presenter, provider, newItem)
	},
}

func runAIAssistedCreation(
	ctx context.Context,
	presenter *ui.Presenter,
	provider workitem.Provider,
) error {
	// 1. Capture Rough Intent
	roughIntent, err := getRoughIntent()
	if err != nil {
		return err
	}

	// 2. Generate and Display Prompt
	if err := displayAIPrompt(presenter, roughIntent); err != nil {
		return err
	}

	// 3. Capture & Parse AI Response
	aiResp, err := getAIResponse(presenter)
	if err != nil {
		return err
	}

	// 4. Review & Confirm
	confirmed, err := confirmCreation(presenter, aiResp)
	if err != nil {
		return err
	}

	if !confirmed {
		presenter.Info("Aborted.")

		return nil
	}

	// 5. Execute
	//nolint:exhaustruct // Partial initialization is valid for creation.
	newItem := workitem.WorkItem{
		Title:  aiResp.Title,
		Body:   aiResp.Body,
		Type:   workitem.Type(aiResp.Type),
		Labels: aiResp.Labels,
	}

	return createItem(ctx, presenter, provider, newItem)
}

func getRoughIntent() (string, error) {
	var roughIntent string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Describe your rough idea or intent").
				Description("The AI will transform this into a rigorous PBI.").
				Value(&roughIntent),
		),
	)
	if err := form.Run(); err != nil {
		return "", fmt.Errorf("input failed: %w", err)
	}

	if strings.TrimSpace(roughIntent) == "" {
		return "", errEmptyIntent
	}

	return roughIntent, nil
}

func displayAIPrompt(presenter *ui.Presenter, intent string) error {
	tmpl, err := template.New("ai-pbi-prompt").Parse(aiPromptTemplateStr)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var promptBuf bytes.Buffer

	data := struct {
		PBI    string
		Intent string
	}{
		PBI:    pbiTemplate,
		Intent: intent,
	}

	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	presenter.Header("--- 📋 Copy the Prompt Below to your AI ---")
	//nolint:forbidigo // Printing prompt to stdout is the core feature.
	fmt.Println(promptBuf.String())
	presenter.Header("--- End of Prompt ---")

	return nil
}

func getAIResponse(presenter *ui.Presenter) (aiIssueResponse, error) {
	var jsonInput string

	presenter.Info("Paste the AI's JSON response below:")

	jsonForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("AI Response (JSON)").
				Value(&jsonInput),
		),
	)
	if err := jsonForm.Run(); err != nil {
		return aiIssueResponse{}, fmt.Errorf("input failed: %w", err)
	}

	var aiResp aiIssueResponse

	cleanedJSON := strings.TrimSpace(jsonInput)
	cleanedJSON = strings.TrimPrefix(cleanedJSON, "```json")
	cleanedJSON = strings.TrimPrefix(cleanedJSON, "```")
	cleanedJSON = strings.TrimSuffix(cleanedJSON, "```")

	if err := json.Unmarshal([]byte(cleanedJSON), &aiResp); err != nil {
		presenter.Error("Failed to parse JSON: %v", err)

		return aiIssueResponse{}, errInvalidJSON
	}

	return aiResp, nil
}

func confirmCreation(presenter *ui.Presenter, aiResp aiIssueResponse) (bool, error) {
	presenter.Summary("AI Drafted Issue")
	presenter.Detail("Title: %s", aiResp.Title)
	presenter.Detail("Type:  %s", aiResp.Type)
	presenter.Detail("Labels: %v", aiResp.Labels)
	presenter.Header("Body Preview:")
	//nolint:forbidigo // Printing body for review.
	fmt.Println(aiResp.Body)
	presenter.Newline()

	//nolint:wrapcheck // Wrapping is handled by caller.
	return presenter.PromptForConfirmation("Create this issue?")
}

func createItem(
	ctx context.Context,
	presenter *ui.Presenter,
	provider workitem.Provider,
	item workitem.WorkItem,
) error {
	presenter.Summary("Creating work item...")

	createdItem, err := provider.CreateItem(ctx, item)
	if err != nil {
		presenter.Error("Failed to create work item: %v", err)

		return fmt.Errorf("failed to create item: %w", err)
	}

	presenter.Success("Successfully created work item: %s", createdItem.URL)

	return nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(createLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CreateCmd.Short = desc.Short
	CreateCmd.Long = desc.Long

	CreateCmd.Flags().
		StringVarP(&issueType, "type", "t", "Task", "Type of the issue (Task, Story, Bug, Chore, PBI)")
	CreateCmd.Flags().StringVarP(&issueTitle, "title", "T", "", "Title of the issue")
	CreateCmd.Flags().StringVarP(&issueBody, "body", "b", "", "Body of the issue")
	CreateCmd.Flags().
		BoolVar(&aiAssist, "ai", false, "Use AI to draft the issue from a rough intent")
}
