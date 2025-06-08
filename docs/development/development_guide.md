---
title: "ContextVibes CLI: Local Development Guide"
artifactVersion: "1.0.0"
summary: "A comprehensive guide for developers on setting up a local development environment for the ContextVibes CLI. Covers prerequisites, initial setup, building the binary, and running common development tasks like testing, linting, formatting, and debugging."
owner: "Scribe"
createdDate: "2025-06-08T12:00:00Z"
lastModifiedDate: "2025-06-08T12:00:00Z"
defaultTargetPath: "docs/DEVELOPMENT.md"
usageGuidance:
  - "Use as the primary guide for setting up a local development environment for the ContextVibes CLI."
  - "Consult for standard commands to run tests (`go test`), lint (`go vet`), and format code (`go fmt`)."
  - "Reference for building the CLI from source and for debugging instructions using Delve."
  - "Provides instructions on how to 'dogfood' (use `contextvibes` on its own codebase)."
tags:
  - "development-guide"
  - "contributing"
  - "local-setup"
  - "go"
  - "git"
  - "nix"
  - "delve"
  - "testing"
  - "linting"
  - "formatting"
  - "building"
  - "debugging"
  - "dogfooding"
  - "go-test"
  - "go-vet"
  - "go-fmt"
  - "golangci-lint"
  - "contextvibes"
  - "cli"
---

# Local Development Guide for ContextVibes CLI

This guide provides instructions for setting up your local development environment to work on the `contextvibes` Go CLI, run tests, and follow project conventions.

## Prerequisites

Before you begin, ensure you have the following installed and configured:

1.  **Git:** For version control. [Installation Guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).
2.  **Go:** Refer to the `go.mod` file for the specific version (e.g., 1.24.x or later). It's recommended to manage Go versions using a tool like `gvm` or asdf, or ensure your system Go matches. [Official Go Downloads](https://go.dev/dl/).
3.  **(Optional but Recommended) Nix:** If a `.idx/dev.nix` or `flake.nix` is present in the future for this project, using Nix can help create a reproducible development environment with all tools.
4.  **External Tools for Full Command Testing:** To test all `contextvibes` commands, you'll need the tools it wraps installed and in your PATH:
    *   Terraform CLI (`terraform`, `tflint`)
    *   Pulumi CLI (`pulumi`)
    *   Python (`python`, `pip`), and Python tools (`isort`, `black`, `flake8`)
    *   Other Go tools if used by quality checks (e.g., `golangci-lint`)

## Initial Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/contextvibes/cli.git
    cd cli
    ```

2.  **(If using Nix) Enter the Nix Development Environment:**
    ```bash
    # Example if a flake.nix is added later:
    # nix develop .#
    ```
    If a Nix environment is defined, it will make specific versions of Go and other tools available.

3.  **Install Go Module Dependencies:**
    ```bash
    go mod download
    go mod tidy
    ```

4.  **Build the CLI:**
    You can build the CLI for your local system:
    ```bash
    go build -o contextvibes ./cmd/contextvibes/main.go
    ```
    You can then run it as `./contextvibes`. Alternatively, install it to your $GOPATH/bin:
    ```bash
    go install ./cmd/contextvibes/main.go
    # Ensure /bin is in your PATH
    ```

## Common Development Tasks

This project uses standard Go commands. You can also use a development build of `contextvibes` itself to manage its own workflow (dogfooding).

### 1. Running Unit Tests

Run all unit tests (excluding integration tests, if any are tagged separately):

```bash
go test ./...
```

To detect race conditions (highly recommended):

```bash
go test -race ./...
```

To get coverage information:

```bash
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### 2. Running Integration Tests (If Applicable)

If integration tests exist (e.g., in an `integration_test/` directory with build tags):

```bash
# Example: go test -tags=integration -v ./integration_test/...
# Refer to specific integration testing docs if available.
```

### 3. Linting

This project may use `golangci-lint` (check for a `.golangci.yml` file). If so:

```bash
golangci-lint run ./...
```

Also, always run:

```bash
go vet ./...
```

You can use a development build of `contextvibes quality` on its own codebase too.

### 4. Formatting Code

Ensure your code is formatted according to Go standards:

```bash
go fmt ./...
# If goimports-reviser or similar is standard for the project:
# goimports-reviser -format ./...
```

Alternatively, use `./contextvibes format` (your dev build).

### 5. Tidying Go Modules

After adding or updating dependencies:

```bash
go mod tidy
```

Commit both `go.mod` and `go.sum` after changes.

## Debugging

The Go ecosystem includes `delve` for debugging.
You can run tests with Delve or attach it to a running process.
For VS Code users, the Go extension provides debugging capabilities that should work with Delve.

Example of running a specific test with Delve:

```bash
# dlv test ./path/to/package -test.run TestSpecificFunction
dlv test ./internal/config -test.run TestConfigLoading # Example
```

## Using `contextvibes` for its Own Development

Once you have a working build of `contextvibes`, you are encouraged to use it for managing your development workflow on the CLI itself:

*   **Daily Branches:** `./contextvibes kickoff -b feature/my-new-cli-feature`
*   **Committing:** `./contextvibes commit -m "feat(command): Add new flag"`
    *   This will use the commit message validation rules defined in `.contextvibes.yaml` (once created for this repo).
*   **Formatting/Quality:** `./contextvibes format` and `./contextvibes quality`
*   **Syncing:** `./contextvibes sync`

## Configuration for Development (`.contextvibes.yaml`)

It's recommended to have a `.contextvibes.yaml` file in the root of this CLI's repository for testing its own configuration loading features. You can start with the sample provided in the documentation or generate one if an `init-config` command is added.

---

This guide should help you get started with developing the ContextVibes CLI.
```