// cmd/bootstrap.go
package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/bootstrap"
	gh "github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Creates a new, standardized Go project from a template.",
	Long: `Launches an interactive wizard to bootstrap a new project.

This command will guide you through a series of questions to gather the necessary
details, then perform the following actions:
1. Create a new repository on GitHub using your authenticated credentials.
2. Clone the new repository to your local machine.
3. Scaffold a best-practice project structure with template files.
4. Create and push the initial commit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Header("--- Project Bootstrap Wizard ---")

		// --- Pre-flight Check: GitHub Token and get authenticated user---
		// We initialize a temporary client just to get the user's login
		tempGHClient, err := gh.NewClient(ctx, AppLogger, "", "") // Owner/repo not needed yet
		if err != nil {
			presenter.Error("GitHub client initialization failed: %v", err)
			presenter.Advice(
				"Please create a GitHub Personal Access Token (PAT) with 'repo' scope and set it as the '%s' environment variable.",
				gh.GHTokenEnvVar,
			)
			return err
		}

		authedUser, err := tempGHClient.GetAuthenticatedUserLogin(ctx)
		if err != nil {
			presenter.Error("Failed to get authenticated user from GitHub token: %v", err)
			return err
		}

		// --- Interactive Data Gathering ---
		var (
			repoName     string
			description  string
			visibility   string
			goModulePath string
			isPrivate    bool
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Repository Name?").Value(&repoName),
				huh.NewInput().Title("Description?").Value(&description),
				huh.NewSelect[string]().Title("Visibility?").Options(
					huh.NewOption("Private", "private"),
					huh.NewOption("Public", "public"),
				).Value(&visibility),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		isPrivate = (visibility == "private")

		defaultModulePath := fmt.Sprintf("github.com/%s/%s", authedUser, repoName)
		modulePathForm := huh.NewInput().
			Title("Go Module Path?").
			Placeholder(defaultModulePath).
			Value(&goModulePath)

		if err := modulePathForm.Run(); err != nil {
			return err
		}
		if goModulePath == "" {
			goModulePath = defaultModulePath
		}

		// --- Confirmation ---
		presenter.Newline()
		presenter.Summary("Bootstrap Plan")
		presenter.Detail("GitHub Repo:  %s/%s (%s)", authedUser, repoName, visibility)
		presenter.Detail("Local Path:   ./%s", repoName)
		presenter.Detail("Go Module:    %s", goModulePath)
		presenter.Newline()

		confirmed, err := presenter.PromptForConfirmation("Proceed with this plan?")
		if err != nil {
			return err
		}
		if !confirmed {
			presenter.Info("Bootstrap aborted by user.")
			return nil
		}

		// --- Initialize final client now we have all info ---
		ghClient, err := gh.NewClient(ctx, AppLogger, authedUser, repoName)
		if err != nil {
			presenter.Error("GitHub client initialization failed: %v", err)
			return err
		}

		// --- Workflow Execution ---
		createRepoStep := &bootstrap.CreateRemoteRepoStep{
			GHClient:        ghClient,
			Presenter:       presenter,
			Owner:           authedUser, // Pass the owner
			RepoName:        repoName,
			RepoDescription: description,
			IsPrivate:       isPrivate,
		}
		cloneRepoStep := &bootstrap.CloneRepoStep{
			ExecClient: ExecClient,
			Presenter:  presenter,
			LocalPath:  repoName,
		}
		scaffoldStep := &bootstrap.ScaffoldProjectStep{
			Presenter:    presenter,
			LocalPath:    repoName,
			AppName:      repoName,
			GoModulePath: goModulePath,
		}
		commitPushStep := &bootstrap.InitialCommitAndPushStep{
			ExecClient: ExecClient,
			Presenter:  presenter,
			LocalPath:  repoName,
		}

		presenter.Newline()
		presenter.Step(createRepoStep.Description())
		if err := createRepoStep.Execute(ctx); err != nil {
			return err
		}

		cloneRepoStep.CloneURL = createRepoStep.CloneURL // Pass output to next step
		presenter.Step(cloneRepoStep.Description())
		if err := cloneRepoStep.Execute(ctx); err != nil {
			return err
		}

		presenter.Step(scaffoldStep.Description())
		if err := scaffoldStep.Execute(ctx); err != nil {
			return err
		}

		presenter.Step(commitPushStep.Description())
		if err := commitPushStep.Execute(ctx); err != nil {
			return err
		}

		presenter.Newline()
		presenter.Success("Project '%s' successfully bootstrapped!", repoName)
		presenter.Advice("Next steps: cd %s", repoName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}
