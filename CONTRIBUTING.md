# Contributing to Context Vibes CLI

Thank you for considering contributing to Context Vibes! We welcome improvements, bug fixes, and new features that align with the goal of streamlining development workflows and improving AI context generation.

## Community Guidelines

We strive to maintain a positive and welcoming environment. All participants are expected to act professionally and respectfully toward others, following the simple guidelines outlined in the `README.md`'s Code of Conduct section.

## Getting Started

1.  **Prerequisites:** Ensure you have Go (`1.24` or later recommended) and Git installed and configured correctly. Access to tools like `terraform` or `pulumi` might be needed to test specific commands locally.
2.  **Fork & Clone:** Fork the repository on GitHub (`github.com/contextvibes/cli` - *Adjust URL if needed*) and clone your fork locally:
    ```bash
    # Replace YOUR_USERNAME with your actual GitHub username
    git clone https://github.com/YOUR_USERNAME/cli.git contextvibes-cli
    cd contextvibes-cli
    ```
3.  **Build:** Ensure you can build the binary successfully:
    ```bash
    go build -o contextvibes .
    ```
    You can also install it to your `$GOPATH/bin` for easier testing:
    ```bash
    go install .
    ```

## Making Changes

1.  **Create a Branch:** Before making changes, create a new branch from the `main` branch:
    ```bash
    git checkout main
    git pull origin main # Ensure your main is up-to-date
    git checkout -b feature/your-feature-name # Example: feature/add-nodejs-support
    # or
    git checkout -b fix/issue-description # Example: fix/improve-plan-error-msg
    ```
2.  **Implement:** Make your code changes. Keep changes focused on a single feature or bug fix per branch.
3.  **Follow Style:** Adhere to standard Go formatting (`gofmt`) and linting practices. Use `go vet ./...` to catch common issues.
4.  **Test:**
    *   **Manual:** Run the commands you've modified in relevant test projects (e.g., a simple Git repo, a Terraform project) to ensure they behave as expected.
    *   **Automated:** If adding new functions, especially in `internal/`, please add corresponding unit tests (`_test.go` files). Contributions to increase overall test coverage are highly encouraged. Run tests using:
        ```bash
        go test ./...
        ```
5.  **Commit:** Commit your changes using clear and descriptive commit messages. Consider following the [Conventional Commits](https://www.conventionalcommits.org/) specification (e.g., `feat: ...`, `fix: ...`, `refactor: ...`, `docs: ...`).
    ```bash
    git add .
    git commit -m "feat(plan): Add detection for Rust Cargo.toml"
    ```

## Submitting a Pull Request

1.  **Push:** Push your feature or fix branch to your fork on GitHub:
    ```bash
    git push origin feature/your-feature-name
    ```
2.  **Open PR:** Go to the original `contextvibes/cli` repository on GitHub. GitHub should automatically detect your pushed branch and prompt you to create a Pull Request.
3.  **Describe:** Fill out the Pull Request template, clearly describing the problem you're solving and the changes you've made. Link to any relevant issues.
4.  **Review:** Respond to any feedback or code review comments. The maintainers will review your PR and merge it if it meets the project's standards.

## Known Issues & TODOs

This list tracks known bugs, planned refactorings, and potential enhancements. Feel free to tackle one of these or suggest additions! For a broader view of future plans, see the [ROADMAP.md](ROADMAP.md).

*   **Refactor:**
    *   Improve error message in `cmd/deploy.go` when `tfplan.out` is missing (be more specific about running `plan` first).
    *   Refactor the large `RunE` function in `cmd/describe.go` into smaller helpers for better readability and testability (see also `ROADMAP.md`).
*   **Testing:**
    *   Add comprehensive unit tests. As per `ROADMAP.md`, a high priority is `internal/tools` (including `git.go` utilities) and `internal/project` packages. Contributions to test `internal/git.GitClient` methods are also very welcome.
*   **Enhancements (see `ROADMAP.md` for more details):**
    *   Add a global `--verbose` flag to show more detailed output.
    *   Implement support for a configuration file (e.g., `.contextvibes.yaml`) for settings like Git remote/main branch names.
    *   Improve `quality` command, for example, by adding `go mod tidy` check for Go projects.

Thank you for contributing to Context Vibes!