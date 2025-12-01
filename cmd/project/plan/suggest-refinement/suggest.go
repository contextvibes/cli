// cmd/project/plan/suggest-refinement/suggest.go
package suggest

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

var outputFile string

// newProvider is a factory function that returns the configured work item provider.
func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		return github.New(ctx, logger, cfg)
	case "":
		logger.DebugContext(
			ctx,
			"Work item provider not specified in config, defaulting to 'github'",
		)

		return github.New(ctx, logger, cfg)
	default:
		return nil, fmt.Errorf(
			"unsupported work item provider '%s' specified in .contextvibes.yaml",
			cfg.Project.Provider,
		)
	}
}

// SuggestRefinementCmd represents the project plan suggest-refinement command.
var SuggestRefinementCmd = &cobra.Command{
	Use:     "suggest-refinement",
	Short:   "Generate a prompt for an AI to classify untyped issues.",
	Example: `  contextvibes project plan suggest-refinement -o for-ai.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

			return err
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
			err := os.WriteFile(outputFile, []byte(prompt), 0o644)
			if err != nil {
				presenter.Error("Failed to write prompt to file %s: %v", outputFile, err)

				return err
			}
			presenter.Success("AI prompt successfully generated at: %s", outputFile)
		}

		return nil
	},
}

func generateAIPrompt(items []workitem.WorkItem) string {
	var b bytes.Buffer

	fmt.Fprintln(&b, "# AI Prompt: Scrum Master Backlog Refinement")
	fmt.Fprintln(&b, "\n## Your Role & Goal")
	fmt.Fprintln(
		&b,
		"You are an expert Scrum Master. Your goal is to analyze the following list of unclassified GitHub issues. For each issue, you must decide if it is an **Epic**, **Story**, **Task**, **Bug**, or **Chore**. Based on your decision, you will generate a `bash` script that uses the `gh` CLI to apply the correct label to each issue.",
	)

	fmt.Fprintln(&b, "\n## Rules")
	fmt.Fprintln(
		&b,
		"1.  **Analyze Content**: Base your decision on the title and body of each issue.",
	)
	fmt.Fprintln(
		&b,
		"2.  **Use `gh` CLI**: The output script MUST use the format `gh issue edit <number> --add-label <type>` for each issue.",
	)
	fmt.Fprintln(
		&b,
		"3.  **Script Only**: Your final output MUST be a single, runnable `bash` script block and nothing else.",
	)
	fmt.Fprintln(
		&b,
		"4.  **Be Decisive**: Do not skip any issues. Assign a type to every issue provided.",
	)

	fmt.Fprintln(&b, "\n## Unclassified Issues for Review")
	fmt.Fprintln(&b, "---")

	for _, item := range items {
		fmt.Fprintf(&b, "\n### Issue #%d: %s\n", item.Number, item.Title)
		fmt.Fprintln(&b, "```")
		fmt.Fprintln(&b, item.Body)
		fmt.Fprintln(&b, "```")
	}

	return b.String()
}

func init() {
	SuggestRefinementCmd.Flags().
		StringVarP(&outputFile, "output", "o", "", "Output file path for the generated AI prompt. Prints to stdout if empty.")
}
