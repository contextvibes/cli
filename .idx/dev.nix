# .idx/dev.nix
# Merged and Go-focused Nix configuration for Project IDX environment.
# To learn more about how to use Nix to configure your environment
# see: https://developers.google.com/idx/guides/customize-idx-env

{ pkgs, ... }: {
  # Pin to a specific Nixpkgs channel for reproducibility.
  channel = "stable-25.05";

  # The 'pkgs' block defines system-level packages available in your workspace.
  packages = with pkgs; [
    # --- Core Go Development ---
    go # The Go compiler and runtime

    # --- Version Control & GitHub Integration ---
    git # Essential for version control.
    gh # The official GitHub CLI.

    # --- Utilities ---
    tree # Useful for viewing directory structures.
  ];

  # Sets environment variables in the workspace
  env = { };

  idx = {
    # Search for extensions on https://open-vsx.org/ and use "publisher.id"
    extensions = [
      # --- Go Language Support ---
      "golang.go" # Official Go extension (debugging, testing, linting/formatting)

      # --- Version Control ---
      "GitHub.vscode-pull-request-github" # GitHub Pull Request and Issues integration
    ];

    workspace = {
      # Runs when a workspace is first created with this `dev.nix` file
      onCreate = {
        installContextVibesCli = ''
          echo "Installing contextvibes CLI into ./bin..."
          LOCAL_BIN_DIR="$(pwd)/bin"
          mkdir -p "$LOCAL_BIN_DIR"
          export GOBIN="$LOCAL_BIN_DIR"
          if go install github.com/contextvibes/cli/cmd/contextvibes@latest; then
            echo "✅ Successfully installed contextvibes to $LOCAL_BIN_DIR"
          else
            echo "❌ ERROR: Failed to install contextvibes."
          fi
          unset GOBIN
        '';
      };
      # Runs every time a workspace is started
            onStart = {
        # Welcome message and version checks for quick diagnostics.
        welcome = "echo '👋 Welcome back!";

        # Add the local ./bin directory (if it exists) to the PATH.
        add-local-bin-to-path = ''
          LOCAL_BIN_DIR="$(pwd)/bin"
          if [ -d "$LOCAL_BIN_DIR" ]; then
            export PATH="$LOCAL_BIN_DIR:$PATH"
            echo "✔️  Local ./bin directory added to PATH."
          fi
        '';
      };
    };

    # Enable previews and customize configuration if you're running web services
    previews = {
      enable = false;
    };
  };
}
