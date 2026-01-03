# -----------------------------------------------------------------------------
# IDX Profile: Go Container Environment (Low-Resource Optimized)
# Version: 1.2.0 (Audited)
# -----------------------------------------------------------------------------
{ pkgs, ... }:

let
  # 1. Define the local config path
  localConfigPath = ./local.nix;

  # 2. Safely import local.nix.
  #    Returns an empty set {} if the file is missing.
  localEnv = if builtins.pathExists localConfigPath
             then import localConfigPath
             else {};
in
{
  # Pin to Nixpkgs version (May 2025 release)
  channel = "stable-25.05";

  packages = with pkgs; [
    # --- Go Toolchain ---
    go_1_25
    gotools     # godoc, goimports, etc.
    govulncheck # Vulnerability detection
    gcc         # Keep this! Needed for 'go test -race' even if CGO is off by default.

    # --- Cloud & Containers ---
    google-cloud-sdk
    docker
    docker-compose

    # --- Security & Identity ---
    gnupg
    pass
    pinentry-curses
    gitleaks

    # --- Utilities ---
    git
    gh

    # --- Local Imports ---
    (import ./contextvibes.nix { inherit pkgs; })
    (import ./golangci-lint.nix { inherit pkgs; })
  ];

  # Enable Docker Daemon
  services.docker.enable = true;

  # ---------------------------------------------------------------------------
  # Environment Configuration
  # Logic: Defaults (Left) // Overrides (Right)
  # ---------------------------------------------------------------------------
  env = {
    # --- Functional Defaults ---
    GOPRIVATE = "github.com/duizendstra-com/*";
    CGO_ENABLED = "0"; # Default to static, override to "1" in local.nix if needed

    # --- Low Resource Tuning (Defaults) ---
    # -p=1 reduces RAM usage but slows builds.
    # Override this in local.nix if you have >4GB RAM.
    GOFLAGS = "-p=1";

    # Cap Runtime Memory to prevent OOM kills
    GOMEMLIMIT = "1024MiB";

    # Limit OS threads to prevent starvation on small VMs
    GOMAXPROCS = "1";

  } // localEnv; # <--- MERGE: local.nix values overwrite the defaults above

  # VS Code & Workspace Lifecycle
  idx = {
    extensions = [
      "golang.go"
    ];

    workspace = {
      # Runs when the workspace starts (every time)
      onStart = {
        # Set GPG_TTY dynamically for the current session to enable pinentry
        init-shell = ''
          if ! grep -q "GPG_TTY" ~/.bashrc; then
            echo '# GPG Signing Fix' >> ~/.bashrc
            echo 'export GPG_TTY=$(tty)' >> ~/.bashrc
          fi
        '';
      };
    };
  };
}
