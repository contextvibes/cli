// Package status provides the command to show git status.
package status

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed status.md.tpl
var statusLongDescription string

// StatusCmd represents the status command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var StatusCmd = &cobra.Command{
	Use:     "status",
	Example: `  contextvibes factory status`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		workDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:   globals.AppLogger,
			Executor: globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		stdout, _, err := client.GetStatusShort(ctx)
		if err != nil {
			return fmt.Errorf("failed to get git status: %w", err)
		}

		trimmedStdout := strings.TrimSpace(stdout)
		if trimmedStdout == "" {
			presenter.Info("Working tree is clean.")
		} else {
			presenter.InfoPrefixOnly()

			fmt.Fprintln(presenter.Out(), "  Current Changes (--short format):")
			scanner := bufio.NewScanner(strings.NewReader(trimmedStdout))
			for scanner.Scan() {

				fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
			presenter.Newline()
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(statusLongDescription, nil)
	if err != nil {
		panic(err)
	}

	StatusCmd.Short = desc.Short
	StatusCmd.Long = desc.Long
}
