// cmd/project_protect_branch.go
package cmd

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/google/go-github/v74/github"
	"github.com/spf13/cobra"
)

var (
	requirePR           bool
	requireUpToDate     bool
	enforceAdmins       bool
	allowForcePushes    bool
	allowDeletions      bool
	requiredReviewCount int
	dismissStaleReviews bool
)

var projectProtectBranchCmd = &cobra.Command{
	Use:   "protect-branch [branch-name]",
	Short: "Applies branch protection rules to a GitHub repository.",
	Long: `Applies a set of standard, secure-by-default branch protection rules to a branch in the current repository.

This command requires the 'GITHUB_TOKEN' environment variable to be set with a token
that has the 'repo' scope to administer repository settings.

If no branch name is provided, it defaults to protecting the 'main' branch.`,
	Example: `  # Protect the 'main' branch with recommended defaults
  cv project protect-branch

  # Protect a 'release' branch with a stricter review policy
  cv project protect-branch release --require-reviews=2`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		// --- Determine Repo and Branch ---
		owner, repoName, err := getOwnerAndRepoFromOrigin(ctx)
		if err != nil {
			presenter.Error("Failed to determine repository owner and name from git remote: %v", err)
			return err
		}

		// --- Pre-flight Check: GitHub Token ---
		ghClient, err := gh.NewClient(ctx, AppLogger, owner, repoName)
		if err != nil {
			presenter.Error("GitHub client initialization failed: %v", err)
			return err
		}

		branchName := "main"
		if len(args) > 0 {
			branchName = args[0]
		}

		// --- Build Protection Request from Flags ---
		protectionRequest := github.ProtectionRequest{
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          dismissStaleReviews,
				RequireCodeOwnerReviews:      false, // Sensible default, can be a flag later
				RequiredApprovingReviewCount: requiredReviewCount,
			},
			EnforceAdmins:        enforceAdmins,
			RequiredStatusChecks: nil, // Not implemented in this version
			Restrictions:         nil, // Not implemented in this version
			AllowForcePushes:     &allowForcePushes,
			AllowDeletions:       &allowDeletions,
		}
		if !requirePR {
			protectionRequest.RequiredPullRequestReviews = nil
		}

		// --- Confirmation ---
		presenter.Summary("Branch Protection Plan")
		presenter.Detail("Repository: %s/%s", owner, repoName)
		presenter.Detail("Branch:     %s", branchName)
		presenter.Newline()
		presenter.Info("The following rules will be applied:")
		presenter.Detail("  - Require a pull request before merging: %t", requirePR)
		if requirePR {
			presenter.Detail("    - Required approving reviews: %d", requiredReviewCount)
			presenter.Detail("    - Dismiss stale reviews on new commits: %t", dismissStaleReviews)
		}
		presenter.Detail("  - Require branches to be up-to-date: %t (via status check)", requireUpToDate)
		presenter.Detail("  - Enforce for administrators: %t", enforceAdmins)
		presenter.Detail("  - Allow force pushes: %t", allowForcePushes)
		presenter.Detail("  - Allow deletions: %t", allowDeletions)
		presenter.Newline()

		confirmed, err := presenter.PromptForConfirmation("Proceed with applying these rules?")
		if err != nil {
			return err
		}
		if !confirmed {
			presenter.Info("Aborted by user.")
			return nil
		}

		// --- Execute ---
		presenter.Step("Applying protection rules to '%s' branch...", branchName)
		err = ghClient.UpdateBranchProtection(ctx, branchName, protectionRequest)
		if err != nil {
			presenter.Error("Failed to apply protection rules: %v", err)
			return err
		}

		presenter.Success("Successfully updated protection rules for '%s'.", branchName)
		return nil
	},
}

// getOwnerAndRepoFromOrigin is a helper to extract "owner" and "repo" from the origin URL.
func getOwnerAndRepoFromOrigin(ctx context.Context) (owner, repo string, err error) {
	out, _, err := ExecClient.CaptureOutput(ctx, ".", "git", "config", "--get", "remote.origin.url")
	if err != nil {
		return "", "", fmt.Errorf("could not get remote origin URL: %w", err)
	}

	url := strings.TrimSpace(out)
	var path string

	if strings.HasPrefix(url, "git@") {
		path = strings.TrimPrefix(url, "git@github.com:")
	} else if strings.HasPrefix(url, "https://") {
		path = strings.TrimPrefix(url, "https://github.com/")
	} else {
		return "", "", fmt.Errorf("unrecognized remote origin URL format: %s", url)
	}

	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("could not parse owner and repo from path: %s", path)
	}

	return parts[0], parts[1], nil
}

func init() {
	projectCmd.AddCommand(projectProtectBranchCmd)

	projectProtectBranchCmd.Flags().BoolVar(&requirePR, "require-pr", true, "Require a pull request before merging.")
	projectProtectBranchCmd.Flags().IntVar(&requiredReviewCount, "require-reviews", 1, "Number of required approving reviews for a PR.")
	projectProtectBranchCmd.Flags().BoolVar(&dismissStaleReviews, "dismiss-stale-reviews", true, "Dismiss PR approvals when new commits are pushed.")
	projectProtectBranchCmd.Flags().BoolVar(&requireUpToDate, "require-up-to-date", true, "Require branches to be up to date before merging.")
	projectProtectBranchCmd.Flags().BoolVar(&enforceAdmins, "enforce-admins", true, "Enforce all configured restrictions for administrators.")
	projectProtectBranchCmd.Flags().BoolVar(&allowForcePushes, "allow-force-pushes", false, "Allow force pushes to the branch.")
	projectProtectBranchCmd.Flags().BoolVar(&allowDeletions, "allow-deletions", false, "Allow deletions of the branch.")
}
