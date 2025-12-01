# .idx/dev.nix
# Defines the complete, reproducible development environment for the project using Nix.

{ pkgs, ... }:

let
  # Imports custom package definitions to keep the main environment configuration clean and modular.
  contextvibes = import ./contextvibes.nix { pkgs = pkgs; };
  golangci-lint = import ./golangci-lint.nix { pkgs = pkgs; };

in
{
  # Pins the environment to the November 2025 release to guarantee that all developers
  # and CI pipelines operate with the exact same tool versions.
  channel = "stable-25.05";

  # Installs the specific system-level tools required for the development workflow.
  packages = with pkgs; [
    # The core language toolchain required to build and test the application.
    go_1_25

    # Required for CGO support (building tools like Delve and gopls).
    gcc

    # Tools for managing source code history and interacting with GitHub.
    gh
    git

    # Utilities for managing GPG keys and secrets, enabling signed commits and secure identity.
    pass
    gnupg
    pinentry-curses

    # Custom CLI tools specific to this project's workflow and code quality standards.
    contextvibes
    golangci-lint
  ];

  # Configures the VS Code editor environment to automatically provide Go language support upon startup.
  idx = {
    extensions = [
      "golang.go"
    ];
  };
}
