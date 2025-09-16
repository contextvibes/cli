# .idx/dev.nix
#
# This file defines the complete, reproducible development environment for the
# project using Nix, specifically tailored for Firebase Studio (IDX).
#
# By defining all tools, packages, and services declaratively, we ensure that
# every developer and CI/CD pipeline operates in an identical environment.
#
# For more information, see the official documentation:
# - Nix Package Search: https://search.nixos.org/packages
# - IDX dev.nix Reference: https://firebase.google.com/docs/studio/devnix-reference

{ pkgs, ... }:

let
  # Import and build the 'contextvibes' CLI from the local contextvibes.nix file.
  # This is the idiomatic Nix approach, treating our own tool as a reproducible
  # package within the environment. It is more robust than an imperative build script.
  contextvibes = import ./contextvibes.nix { pkgs = pkgs; };

in
{
  # -----------------------------------------------------------------------------
  # NIXPKGS CHANNEL
  # -----------------------------------------------------------------------------
  # Pin the environment to a specific Nixpkgs channel. This guarantees that all
  # packages are sourced from the exact same revision, ensuring maximum
  # reproducibility across all machines and over time.
  channel = "stable-25.05";

  # -----------------------------------------------------------------------------
  # ENVIRONMENT PACKAGES
  # -----------------------------------------------------------------------------
  # Defines all system-level packages available in the workspace terminal.
  # Packages are grouped by their function and sorted alphabetically within each group.
  packages = with pkgs; [
    # --- Go Development Toolchain ---
    delve              # The premier debugger for the Go language.
    gcc                # The GNU Compiler Collection, required by Go for CGO support.
    go                 # The Go compiler and core toolchain.
    goimports-reviser  # A tool to format and revise Go import statements.
    golangci-lint      # A fast, parallel Go linter that aggregates multiple linters.
    gopls              # The official Go Language Server, providing IDE features.
    gotools            # A collection of supplementary Go tools used by IDE extensions.
    govulncheck        # Scans Go source code for known vulnerabilities in dependencies.
    python313          # The Python interpreter, often needed for various helper scripts.

    # --- Cloud, Containers & Automation ---
    docker             # The Docker CLI and engine for building and running containers.
    docker-compose     # A tool for defining and running multi-container Docker applications.
    google-cloud-sdk   # The `gcloud` CLI for interacting with Google Cloud Platform.
    pulumi             # Infrastructure as Code tool for creating and managing cloud resources.
    pulumiPackages.pulumi-go   # The Go language plugin for Pulumi.

    # --- Code Quality & Formatting (Non-Go) ---
    nodePackages.markdownlint-cli # Linter to enforce standards in Markdown files.
    nodejs             # JavaScript runtime, required by markdownlint-cli.
    shellcheck         # A static analysis tool for finding bugs in shell scripts.
    shfmt              # An auto-formatter for shell scripts to ensure consistent style.

    # --- Version Control & CLI Utilities ---
    file               # A utility to determine file types.
    gh                 # The official GitHub CLI for managing repositories and PRs.
    git                # The distributed version control system.
    gum                # A tool for creating glamorous, interactive shell scripts.
    jq                 # A command-line JSON processor for scripting and data manipulation.
    tree               # A utility to display directory structures in a tree-like format.
    yq-go              # A portable command-line YAML, JSON, and XML processor.

    # --- Project-Specific Tools ---
    contextvibes       # The custom-built 'cv' CLI, managed via contextvibes.nix.
  ];

  # -----------------------------------------------------------------------------
  # ENVIRONMENT VARIABLES
  # -----------------------------------------------------------------------------
  # Sets global environment variables for the entire workspace.
  env = {
    # Tells the Go toolchain that modules under this path are private.
    # This ensures they are fetched directly using Git, bypassing the public
    # Go proxy and checksum database (sum.golang.org).
    GOPRIVATE = "github.com/duizendstra-com/*";
  };

  # -----------------------------------------------------------------------------
  # SYSTEM SERVICES
  # -----------------------------------------------------------------------------
  # Enables and configures system-level services within the environment.
  services.docker.enable = true; # Enable and start the Docker daemon.

  # -----------------------------------------------------------------------------
  # FIREBASE STUDIO (IDX) CONFIGURATION
  # -----------------------------------------------------------------------------
  # IDX-specific settings that configure the editor environment.
  idx = {
    # A list of VS Code extensions to automatically install from the Open VSX Registry.
    # Extensions are grouped by function for clarity.
    extensions = [
      # --- Core Language Support ---
      "golang.go"          # Official Go extension (debugging, testing, linting).
      "ms-python.debugpy"  # Python debugging support.
      "ms-python.python"   # Core Python language support.

      # --- Code Quality & Formatting ---
      "DavidAnson.vscode-markdownlint" # Integrates markdownlint into the editor.
      "timonwong.shellcheck"         # Integrates shellcheck for live linting of shell scripts.

      # --- Version Control & Containers ---
      "GitHub.vscode-pull-request-github" # GitHub Pull Request and Issues integration.
      "ms-azuretools.vscode-containers"   # Adds container-related features and commands.
      "ms-azuretools.vscode-docker"       # Provides Docker integration for VS Code.
    ];
  };
}
