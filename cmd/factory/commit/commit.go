// cmd/factory/commit/commit.go
package commit

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed commit.md.tpl
var commitLongDescription string

var commitMessageFlag string

// CommitCmd represents the commit command.
var CommitCmd = &cobra.Command{
	Use:     "commit -m <message>",
	Example: `  contextvibes factory commit -m "feat(auth): Implement OTP login"`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if strings.TrimSpace(commitMessageFlag) == "" {
			return errors.New("commit message is required via -m flag")
		}

		validationRule := globals.LoadedAppConfig.Validation.CommitMessage
		validationEnabled := validationRule.Enable == nil || *validationRule.Enable

		if validationEnabled {
			pattern := validationRule.Pattern
			if pattern == "" {
				pattern = config.DefaultCommitMessagePattern
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return errors.New("invalid commit message validation regex")
			}
			if !re.MatchString(commitMessageFlag) {
				presenter.Error("Invalid commit message format.")
				presenter.Advice("Message must match pattern: %s", pattern)

				return errors.New("invalid commit message format")
			}
		}

		gitCfg := git.GitClientConfig{
			Logger:                globals.AppLogger,
			DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, ".", gitCfg)
		if err != nil {
			return err
		}

		if err := client.AddAll(ctx); err != nil {
			return err
		}

		hasStaged, err := client.HasStagedChanges(ctx)
		if err != nil {
			return err
		}
		if !hasStaged {
			presenter.Info("No changes were staged for commit.")

			return nil
		}

		currentBranch, _ := client.GetCurrentBranchName(ctx)
		statusOutput, _, _ := client.GetStatusShort(ctx)
		presenter.InfoPrefixOnly()
		_, _ = fmt.Fprintf(presenter.Out(), "  Branch: %s\n", currentBranch)
		_, _ = fmt.Fprintf(presenter.Out(), "  Commit Message: %s\n", commitMessageFlag)
		_, _ = fmt.Fprintf(presenter.Out(), "  Staged Changes:\n%s\n", statusOutput)

		if !globals.AssumeYes {
			confirmed, err := presenter.PromptForConfirmation("Proceed?")
			if err != nil || !confirmed {
				return errors.New("commit aborted")
			}
		}

		return client.Commit(ctx, commitMessageFlag)
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(commitLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CommitCmd.Short = desc.Short
	CommitCmd.Long = desc.Long
	CommitCmd.Flags().
		StringVarP(&commitMessageFlag, "message", "m", "", "Commit message (required)")
}
