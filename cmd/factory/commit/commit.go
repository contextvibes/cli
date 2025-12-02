// Package commit provides the command to commit changes.
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

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var commitMessageFlag string

// CommitCmd represents the commit command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CommitCmd = &cobra.Command{
	Use:     "commit -m <message>",
	Example: `  contextvibes factory commit -m "feat(auth): Implement OTP login"`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if strings.TrimSpace(commitMessageFlag) == "" {
			//nolint:err113 // Dynamic error is appropriate here.
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
				//nolint:err113 // Dynamic error is appropriate here.
				return errors.New("invalid commit message validation regex")
			}
			if !re.MatchString(commitMessageFlag) {
				presenter.Error("Invalid commit message format.")
				presenter.Advice("Message must match pattern: %s", pattern)

				//nolint:err113 // Dynamic error is appropriate here.
				return errors.New("invalid commit message format")
			}
		}

		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:                globals.AppLogger,
			DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, ".", gitCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		err = client.AddAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}

		hasStaged, err := client.HasStagedChanges(ctx)
		if err != nil {
			return fmt.Errorf("failed to check staged changes: %w", err)
		}
		if !hasStaged {
			presenter.Info("No changes were staged for commit.")

			return nil
		}

		currentBranch, _ := client.GetCurrentBranchName(ctx)
		statusOutput, _, _ := client.GetStatusShort(ctx)
		presenter.InfoPrefixOnly()
		//nolint:errcheck // Printing to stdout is best effort.
		fmt.Fprintf(presenter.Out(), "  Branch: %s\n", currentBranch)
		//nolint:errcheck // Printing to stdout is best effort.
		fmt.Fprintf(presenter.Out(), "  Commit Message: %s\n", commitMessageFlag)
		//nolint:errcheck // Printing to stdout is best effort.
		fmt.Fprintf(presenter.Out(), "  Staged Changes:\n%s\n", statusOutput)

		if !globals.AssumeYes {
			confirmed, err := presenter.PromptForConfirmation("Proceed?")
			if err != nil || !confirmed {
				//nolint:err113 // Dynamic error is appropriate here.
				return errors.New("commit aborted")
			}
		}

		return client.Commit(ctx, commitMessageFlag)
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
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
