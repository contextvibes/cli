// cmd/project/plan/refine/refine.go
package refine

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed refine.md.tpl
var refineLongDescription string

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

// RefineCmd represents the project plan refine command.
var RefineCmd = &cobra.Command{
	Use:     "refine",
	Short:   "Interactively classify untyped issues.",
	Example: `  contextvibes project plan refine`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		presenter.Summary("Finding unclassified issues to refine...")

		// Construct a search query to find open issues that are not PRs and are missing all type labels.
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

		presenter.Info("Found %d unclassified issues. Starting refinement session...", len(items))
		presenter.Newline()

		for _, item := range items {
			var issueType string
			prompt := fmt.Sprintf("Classify #%d: %s", item.Number, item.Title)

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewNote().Title(prompt).Description(item.Body),
					huh.NewSelect[string]().
						Title("What is the correct type for this issue?").
						Options(
							huh.NewOption("Epic", "epic"),
							huh.NewOption("Story", "story"),
							huh.NewOption("Task", "task"),
							huh.NewOption("Bug", "bug"),
							huh.NewOption("Chore", "chore"),
							huh.NewOption("Skip", "skip"),
						).
						Value(&issueType),
				),
			)

			err := form.Run()
			if err != nil {
				return err // User likely hit Ctrl+C
			}

			if issueType != "skip" && issueType != "" {
				// Add the new label to the existing labels
				updatedItem := item
				updatedItem.Labels = append(updatedItem.Labels, issueType)
				updatedItem.Type = workitem.Type(
					issueType,
				) // Also update the type field for consistency

				_, err := provider.UpdateItem(ctx, item.Number, updatedItem)
				if err != nil {
					presenter.Error(
						"Failed to apply label '%s' to issue #%d: %v",
						issueType,
						item.Number,
						err,
					)
					// We continue to the next issue even if one fails
				} else {
					presenter.Success("✓ Applied label '%s' to issue #%d.", issueType, item.Number)
				}
			} else {
				presenter.Info("Skipped issue #%d.", item.Number)
			}
			presenter.Newline()
		}

		presenter.Success("Refinement session complete.")

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(refineLongDescription, nil)
	if err != nil {
		panic(err)
	}

	RefineCmd.Long = desc.Long
}
