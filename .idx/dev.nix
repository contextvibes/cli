# .idx/dev.nix
#
# This file defines the complete, reproducible development environment for the
# contextvibes-cli project using Nix, tailored for Firebase Studio (IDX).
#
# Key Documentation:
# - Nix Package Search: https://search.nixos.org/packages
# - IDX dev.nix Reference: https://firebase.google.com/docs/studio/devnix-reference

{ pkgs, ... }:
let
  # Declaratively build the 'contextvibes' CLI from the local source code.
  # This is the idiomatic Nix approach, ensuring the tool is a reproducible,
  # cacheable package within our environment. It is superior to an imperative
  # 'go install' script in the onCreate hook.
  contextvibes = import ./contextvibes.nix { pkgs = pkgs; };
in
{
  # -----------------------------------------------------------------------------
  # NIXPKGS CHANNEL
  # -----------------------------------------------------------------------------
  # Pin to a specific stable channel for maximum reproducibility. All developers
  # and CI/CD will use the exact same package versions.
  channel = "stable-25.05";

  # -----------------------------------------------------------------------------
  # ENVIRONMENT PACKAGES
  # -----------------------------------------------------------------------------
  # All system-level packages available in the workspace terminal.
  # Packages are grouped by function and sorted alphabetically within groups.
  packages = with pkgs; [
    # --- Core Go Development Toolchain ---
    delve # The Go debugger.
    gcc # Required by Go for CGO support.
    go # The Go compiler and core toolchain.
    gopls # The official Go Language Server for IDE features.
    gotools # Supplementary Go tools used by IDE extensions.
    goimports-reviser # Formats and revises Go import statements.
    golangci-lint # A fast Go linter that runs multiple linters in parallel.
    govulncheck # Scans for known vulnerabilities in Go dependencies.
    python313 # Python interpreter, often needed for various scripts.

    # --- Automation, Containers & Cloud ---
    docker-compose # For orchestrating local multi-container Docker applications.
    go-task # A task runner for project automation (see Taskfile.yml).
    google-cloud-sdk # The `gcloud` CLI for interacting with Google Cloud Platform.

    # --- Code Quality & Formatting (Non-Go) ---
    nodejs # JavaScript runtime, required for markdownlint-cli.
    nodePackages.markdownlint-cli # Linter to enforce standards in Markdown files.
    shellcheck # Linter for finding bugs in shell scripts.
    shfmt # Auto-formatter for shell scripts.

    # --- Version Control & CLI Utilities ---
    file # A utility to determine file types.
    gh # The official GitHub CLI for interacting with GitHub.
    git # The version control system for managing source code.
    gum # A tool for creating glamorous, interactive shell scripts.
    jq # A command-line JSON processor for scripting.
    tree # A utility to display directory structures as a tree.
    yq-go # A portable command-line YAML processor.

    # --- Custom Project Tools ---
    contextvibes # The custom-built 'contextvibes' CLI tool, built via contextvibes.nix.
  ];

  # -----------------------------------------------------------------------------
  # ENVIRONMENT VARIABLES
  # -----------------------------------------------------------------------------
  # Global environment variables for the workspace.
  env = { };

  # -----------------------------------------------------------------------------
  # FIREBASE STUDIO (IDX) CONFIGURATION
  # -----------------------------------------------------------------------------
  idx = {
    # VS Code extensions to install from https://open-vsx.org/
    # Extensions are grouped by function.
    extensions = [
      # --- Core Language Support ---
      "golang.go" # Official Go extension (debugging, testing, linting).
      "ms-python.python" # Python language support.
      "ms-python.debugpy" # Python debugging support.

      # --- Code Quality & Formatting ---
      "DavidAnson.vscode-markdownlint" # Integrates markdownlint into the editor.
      "timonwong.shellcheck" # Integrates shellcheck for live linting of shell scripts.

      # --- Version Control ---
      "eamodio.gitlens" # Supercharges the Git capabilities built into VS Code.
      "GitHub.vscode-pull-request-github" # GitHub Pull Request and Issues integration.
    ];

    workspace = {
      # Runs ONCE when a workspace is first created.
      # The 'contextvibes' CLI is installed declaratively via the `let` block above,
      # which is the preferred Nix pattern. Use this hook for other one-time setup
      # tasks that are not package installations (e.g., initializing a database).
      onCreate = {
        # example-one-time-setup = ''
        #   echo "Bootstrapping one-time environment setup..."
        # '';
      };

      # Runs EVERY time the workspace is started or restarted.
      # Nix automatically manages the PATH for all packages listed above.
      onStart = {
        welcome = "echo '👋 Welcome back to the contextvibes-cli project!'";
        set-cv-alias = "alias cv='contextvibes'";
      };
    };

    # Defines how to run and preview applications within IDX.
    # Currently disabled as this is a CLI tool, not a web service.
    previews = {
      enable = false;
    };
  };
}
