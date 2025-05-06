# .idx/dev.nix
# Merged and Go-focused Nix configuration for Project IDX environment.
# To learn more about how to use Nix to configure your environment
# see: https://developers.google.com/idx/guides/customize-idx-env

{ pkgs, ... }: {
  # Which nixpkgs channel to use. (https://status.nixos.org/)
  channel = "stable-24.11"; # Or choose a specific Nixpkgs commit/tag

  # Use https://search.nixos.org/packages to find packages for Go development
  packages = [
    # --- Core Go Development ---
    pkgs.go # The Go compiler and runtime

    # --- Version Control ---
    pkgs.git # Essential version control system
    pkgs.gh
  ];

  # Sets environment variables in the workspace
  env = {
  };

  # Enable Docker daemon service if you need to build/run containers
  services.docker.enable = true;

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
      };
      # Runs every time a workspace is started
      onStart = {
      };
    };

    # Enable previews and customize configuration if you're running web services
    previews = {
      enable = false;
    };
  };
}
