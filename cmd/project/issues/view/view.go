// Package view provides the command to view issue details.
package view

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/contextvibes/cli/cmd/project/issues/internal"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed view.md.tpl
var viewLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var withComments bool

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

// ViewCmd represents the project issues view command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ViewCmd = &cobra.Command{
	Use:     "view <issue-number>",
	Short:   "Display the details of a specific issue.",
	Example: `  contextvibes project issues view 123 --comments`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		issueNumber, err := strconv.Atoi(args[0])
		if err != nil {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("invalid issue number provided")
		}

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		presenter.Summary("Fetching details for work item #%d...", issueNumber)
		item, err := provider.GetItem(ctx, issueNumber, withComments)
		if err != nil {
			presenter.Error("Failed to fetch work item: %v", err)

			return fmt.Errorf("failed to get item: %w", err)
		}

		// Use the shared display helper for the main body
		internal.DisplayWorkItem(presenter, item)

		// The view command is still responsible for displaying comments
		if withComments {
			//nolint:errcheck // Printing to stdout is best effort.
			fmt.Fprintf(presenter.Out(), "\n--- Comments (%d) ---\n\n", len(item.Comments))
			for _, comment := range item.Comments {
				presenter.Header(
					fmt.Sprintf(
						"Comment by %s on %s",
						comment.Author,
						comment.CreatedAt.Format("2006-01-02"),
					),
				)
				//nolint:errcheck // Printing to stdout is best effort.
				fmt.Fprintln(presenter.Out(), comment.Body)
				presenter.Separator()
			}
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(viewLongDescription, nil)
	if err != nil {
		panic(err)
	}

	ViewCmd.Long = desc.Long

	ViewCmd.Flags().
		BoolVarP(&withComments, "comments", "c", false, "Include issue comments in the output.")
}
