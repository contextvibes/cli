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

    # --- Version Control ---
    git # Essential version control system
    gh
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
      onCreate = { };
      # Runs every time a workspace is started
      onStart = { };
    };

    # Enable previews and customize configuration if you're running web services
    previews = {
      enable = false;
    };
  };
}
