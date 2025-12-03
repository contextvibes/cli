// Package prdescription provides the command to generate PR description prompts.
package prdescription

import (
	_ "embed"
	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed prdescription.md.tpl
var prDescriptionLongDescription string

// PRDescriptionCmd represents the craft pr-description command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var PRDescriptionCmd = &cobra.Command{
	Use:     "pr-description",
	Aliases: []string{"pr"},
	Short:   "Generates a prompt for an AI to write your Pull Request description.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

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

		mainBranch := client.MainBranchName()
		// Get log and diff from the merge base (changes in this branch vs main)
		log, diff, err := client.GetLogAndDiffFromMergeBase(ctx, mainBranch)
		if err != nil {
			presenter.Error("Failed to get changes against '%s': %v", mainBranch, err)

			return fmt.Errorf("failed to get branch changes: %w", err)
		}

		// Construct the Prompt
		// Note: We use ~~~ for markdown fences to avoid conflict with Go's backtick string literal.
		prompt := fmt.Sprintf(`
# Role
You are a senior software engineer.

# Goal
Write a clear and comprehensive Pull Request description based on the following changes.

# Instructions
1.  **Summary**: Write a high-level summary of the problem solved and the solution.
2.  **Changes**: Use a bulleted list to detail specific changes.
3.  **Format**: Output raw Markdown suitable for a GitHub PR body.

# Commit History
%s

# Code Diff
~~~diff
%s
~~~
`, log, diff)

		presenter.Header("--- Copy the text below to your AI ---")
		//nolint:forbidigo // Printing prompt to stdout is the core feature.
		fmt.Println(prompt)
		presenter.Header("--- End of Prompt ---")

		presenter.Success("Prompt generated. Paste this into your AI chat.")

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(prDescriptionLongDescription, nil)
	if err != nil {
		panic(err)
	}

	PRDescriptionCmd.Short = desc.Short
	PRDescriptionCmd.Long = desc.Long
}
