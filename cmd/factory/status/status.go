// cmd/factory/status/status.go
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
var StatusCmd = &cobra.Command{
	Use:     "status",
	Example: `  contextvibes factory status`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		gitCfg := git.GitClientConfig{
			Logger:   globals.AppLogger,
			Executor: globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			return err
		}

		stdout, _, err := client.GetStatusShort(ctx)
		if err != nil {
			return err
		}

		trimmedStdout := strings.TrimSpace(stdout)
		if trimmedStdout == "" {
			presenter.Info("Working tree is clean.")
		} else {
			presenter.InfoPrefixOnly()
			_, _ = fmt.Fprintln(presenter.Out(), "  Current Changes (--short format):")
			scanner := bufio.NewScanner(strings.NewReader(trimmedStdout))
			for scanner.Scan() {
				_, _ = fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
			presenter.Newline()
		}

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(statusLongDescription, nil)
	if err != nil {
		panic(err)
	}

	StatusCmd.Short = desc.Short
	StatusCmd.Long = desc.Long
}
