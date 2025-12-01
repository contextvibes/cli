// cmd/project/issues/create/create.go
package create

import (
	"context"
	_ "embed"
	"errors"
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

//go:embed create.md.tpl
var createLongDescription string

var (
	issueType  string
	issueTitle string
	issueBody  string
)

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

// CreateCmd represents the project issues create command.
var CreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new", "add"},
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		if issueTitle == "" { // Interactive Mode
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().Title("What kind of issue is this?").
						Options(huh.NewOption("Task", "Task"), huh.NewOption("Story", "Story"), huh.NewOption("Bug", "Bug"), huh.NewOption("Chore", "Chore")).
						Value(&issueType),
					huh.NewInput().Title("Title?").Value(&issueTitle),
					huh.NewText().Title("Body?").Value(&issueBody),
				),
			)
			err := form.Run()
			if err != nil {
				return err
			}
		}

		if issueTitle == "" {
			return errors.New("title cannot be empty")
		}

		newItem := workitem.WorkItem{
			Title: issueTitle,
			Body:  issueBody,
			Type:  workitem.Type(issueType),
		}

		presenter.Summary("Creating work item...")
		createdItem, err := provider.CreateItem(ctx, newItem)
		if err != nil {
			presenter.Error("Failed to create work item: %v", err)

			return err
		}

		presenter.Success("Successfully created work item: %s", createdItem.URL)

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(createLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CreateCmd.Short = desc.Short
	CreateCmd.Long = desc.Long

	CreateCmd.Flags().
		StringVarP(&issueType, "type", "t", "Task", "Type of the issue (Task, Story, Bug, Chore)")
	CreateCmd.Flags().StringVarP(&issueTitle, "title", "T", "", "Title of the issue")
	CreateCmd.Flags().StringVarP(&issueBody, "body", "b", "", "Body of the issue")
}
