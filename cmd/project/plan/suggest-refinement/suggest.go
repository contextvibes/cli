// Package suggest provides the command to suggest issue refinements.
package suggest

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed suggest.md.tpl
var suggestLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var outputFile string

// newProvider is a factory function that returns the configured work item provider.
//

func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		//nolint:wrapcheck // Wrapping is handled by caller.
		return github.New(ctx, logger, cfg)
	case "":
		logger.DebugContext(
			ctx,
			"Work item provider not specified in config, defaulting to 'github'",
		)

		//nolint:wrapcheck // Wrapping is handled by caller.
		return github.New(ctx, logger, cfg)
	default:
		//nolint:err113 // Dynamic error is appropriate here.
		return nil, fmt.Errorf(
			"unsupported work item provider '%s' specified in .contextvibes.yaml",
			cfg.Project.Provider,
		)
	}
}

// SuggestRefinementCmd represents the project plan suggest-refinement command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var SuggestRefinementCmd = &cobra.Command{
	Use:     "suggest-refinement",
	Short:   "Generate a prompt for an AI to classify untyped issues.",
	Example: `  contextvibes project plan suggest-refinement -o for-ai.md`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		presenter.Summary("Finding unclassified issues for AI analysis...")
		query := "is:open is:issue -label:epic -label:story -label:bug -label:chore"
		items, err := provider.SearchItems(ctx, query)
		if err != nil {
			presenter.Error("Failed to search for unclassified issues: %v", err)

			return fmt.Errorf("failed to search items: %w", err)
		}

		if len(items) == 0 {
			presenter.Success("No unclassified issues found. The backlog is clean!")

			return nil
		}

		prompt := generateAIPrompt(items)

		if outputFile == "" {
			// If no output file, print to stdout

			fmt.Fprint(presenter.Out(), prompt)
		} else {
			//nolint:gosec,mnd // Writing to user-specified file is intended, 0644 is standard.
			err := os.WriteFile(outputFile, []byte(prompt), 0o644)
			if err != nil {
				presenter.Error("Failed to write prompt to file %s: %v", outputFile, err)

				return fmt.Errorf("failed to write output file: %w", err)
			}
			presenter.Success("AI prompt successfully generated at: %s", outputFile)
		}

		return nil
	},
}

//nolint:lll // Prompt strings are long.
func generateAIPrompt(items []workitem.WorkItem) string {
	var buffer bytes.Buffer

	fmt.Fprintln(&buffer, "# AI Prompt: Scrum Master Backlog Refinement")
	fmt.Fprintln(&buffer, "\n## Your Role & Goal")
	fmt.Fprintln(
		&buffer,
		"You are an expert Scrum Master. Your goal is to analyze the following list of unclassified GitHub issues. For each issue, you must decide if it is an **Epic**, **Story**, **Task**, **Bug**, or **Chore**. Based on your decision, you will generate a `bash` script that uses the `gh` CLI to apply the correct label to each issue.",
	)

	fmt.Fprintln(&buffer, "\n## Rules")
	fmt.Fprintln(
		&buffer,
		"1.  **Analyze Content**: Base your decision on the title and body of each issue.",
	)
	fmt.Fprintln(
		&buffer,
		"2.  **Use `gh` CLI**: The output script MUST use the format `gh issue edit <number> --add-label <type>` for each issue.",
	)
	fmt.Fprintln(
		&buffer,
		"3.  **Script Only**: Your final output MUST be a single, runnable `bash` script block and nothing else.",
	)
	fmt.Fprintln(
		&buffer,
		"4.  **Be Decisive**: Do not skip any issues. Assign a type to every issue provided.",
	)

	fmt.Fprintln(&buffer, "\n## Unclassified Issues for Review")
	fmt.Fprintln(&buffer, "---")

	for _, item := range items {
		fmt.Fprintf(&buffer, "\n### Issue #%d: %s\n", item.Number, item.Title)
		fmt.Fprintln(&buffer, "```")
		fmt.Fprintln(&buffer, item.Body)
		fmt.Fprintln(&buffer, "```")
	}

	return buffer.String()
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(suggestLongDescription, nil)
	if err != nil {
		panic(err)
	}

	SuggestRefinementCmd.Short = desc.Short
	SuggestRefinementCmd.Long = desc.Long

	SuggestRefinementCmd.Flags().
		StringVarP(&outputFile, "output", "o", "", "Output file path for the generated AI prompt. Prints to stdout if empty.")
}
