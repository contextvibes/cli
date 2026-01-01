package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/exec"
)

const (
	devNixTemplate = `# -----------------------------------------------------------------------------
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
    firebase-tools
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
`

	contextvibesNixTemplate = `# -----------------------------------------------------------------------------
# Package: ContextVibes CLI (Hybrid: Binary or Source)
# -----------------------------------------------------------------------------
{ pkgs, 
  # Defaults (Binary Mode - Updated by 'factory upgrade-cli')
  buildType ? "binary",   
  version ? "0.6.0",      
  
  # Binary Specific
  binHash ? "sha256-bdbf55bf902aa567851fcbbc07704b416dee85065a276a47e7df19433c5643ea",
  
  # Source Specific (Required if buildType == "source")
  rev ? "",               
  srcHash ? "",           
  vendorHash ? ""         
}:

if buildType == "source" then
  pkgs.buildGoModule {
    pname = "contextvibes";
    version = version; 
    src = pkgs.fetchFromGitHub {
      owner = "contextvibes";
      repo = "cli";
      rev = rev;
      hash = srcHash;
    };
    vendorHash = vendorHash;
    doCheck = false; 
    postInstall = ''
      mv $out/bin/cli $out/bin/contextvibes || true
    '';
  }
else
  pkgs.stdenv.mkDerivation rec {
    name = "contextvibes-${version}";
    src = pkgs.fetchurl {
      url = "https://github.com/contextvibes/cli/releases/download/v${version}/contextvibes";
      sha256 = binHash;
    };
    dontUnpack = true;
    installPhase = ''
      mkdir -p $out/bin
      install -m 755 $src $out/bin/contextvibes
    '';
  }
`

	golangciLintNixTemplate = `# -----------------------------------------------------------------------------
# Package: GolangCI-Lint (Precompiled)
# Version: 1.64.5
# -----------------------------------------------------------------------------
{ pkgs }:

pkgs.stdenv.mkDerivation rec {
  name = "golangci-lint-bin-${version}";
  version = "1.64.5";

  src = pkgs.fetchurl {
    //nolint:lll // URL is long but necessary.
    url = "https://github.com/golangci/golangci-lint/releases/download/v${version}/golangci-lint-${version}-linux-amd64.tar.gz";
    sha256 = "sha256-zkah8diQ57ZnJZ9wuyNil/XPh5GptrmLQbKD2Ttbbog=";
  };

  # The builder automatically enters the extracted folder, so the binary is just 'golangci-lint'
  installPhase = ''
    mkdir -p $out/bin
    install -m 755 golangci-lint $out/bin/
  '';
}
`

	localNixTemplate = `{
  # Identity
  # GPG_KEY_ID = "YOUR_KEY_ID_HERE";

  # Optional: Override resource limits if you are on a High-Mem instance
  # GOMEMLIMIT = "4096MiB";
  # GOMAXPROCS = "4";
  # GOFLAGS = ""; # Remove the -p=1 restriction
}
`
)

// ScaffoldIDXStep generates the .idx configuration files.
type ScaffoldIDXStep struct {
	Presenter PresenterInterface
	AssumeYes bool
}

// Description returns the step description.
func (s *ScaffoldIDXStep) Description() string {
	return "Scaffold Project IDX Environment (.idx/)"
}

// PreCheck performs pre-flight checks.
func (s *ScaffoldIDXStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step logic.
func (s *ScaffoldIDXStep) Execute(_ context.Context) error {
	idxDir := ".idx"
	//nolint:mnd // 0750 is standard dir permission.
	if err := os.MkdirAll(idxDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .idx directory: %w", err)
	}

	files := map[string]string{
		"dev.nix":           devNixTemplate,
		"contextvibes.nix":  contextvibesNixTemplate,
		"golangci-lint.nix": golangciLintNixTemplate,
		"local.nix":         localNixTemplate,
	}

	for filename, content := range files {
		path := filepath.Join(idxDir, filename)
		shouldWrite := true

		// Check existence
		if _, err := os.Stat(path); err == nil {
			if s.AssumeYes {
				s.Presenter.Info("  ! %s exists. Overwriting (due to --yes).", filename)
			} else {
				confirm, _ := s.Presenter.PromptForConfirmation(fmt.Sprintf("  ? %s exists. Overwrite?", filename))
				if !confirm {
					s.Presenter.Info("  ~ Skipping %s.", filename)
					shouldWrite = false
				}
			}
		}

		if shouldWrite {
			//nolint:mnd // 0600 is standard secure file permission.
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				return fmt.Errorf("failed to write %s: %w", filename, err)
			}
			s.Presenter.Success("  + Wrote %s", filename)
		}
	}

	s.Presenter.Newline()
	s.Presenter.Advice("Environment scaffolded. You may need to rebuild your environment for changes to take effect.")
	s.Presenter.Advice("Edit .idx/local.nix to set your GPG_KEY_ID.")

	return nil
}

// ScaffoldFirebaseStep initializes Firebase.
type ScaffoldFirebaseStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  PresenterInterface
}

// Description returns the step description.
func (s *ScaffoldFirebaseStep) Description() string {
	return "Scaffold Firebase Environment"
}

// PreCheck performs pre-flight checks.
func (s *ScaffoldFirebaseStep) PreCheck(_ context.Context) error {
	if !s.ExecClient.CommandExists("firebase") {
		s.Presenter.Error("Firebase CLI not found.")
		s.Presenter.Advice("Please rebuild your environment (dev.nix) to include 'firebase-tools'.")
		//nolint:err113 // Dynamic error is appropriate here.
		return fmt.Errorf("firebase-tools missing")
	}
	return nil
}

// Execute runs the step logic.
func (s *ScaffoldFirebaseStep) Execute(ctx context.Context) error {
	// Login Check
	_, _, err := s.ExecClient.CaptureOutput(ctx, ".", "firebase", "projects:list", "--json")
	if err != nil {
		s.Presenter.Warning("You do not appear to be logged in to Firebase.")
		s.Presenter.Step("Running 'firebase login'...")
		err = s.ExecClient.Execute(ctx, ".", "firebase", "login")
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	s.Presenter.Step("Initializing Firebase Project Structure...")
	s.Presenter.Info("This will guide you through creating firebase.json and .firebaserc")

	err = s.ExecClient.Execute(ctx, ".", "firebase", "init")
	if err != nil {
		return fmt.Errorf("firebase init failed: %w", err)
	}

	s.Presenter.Success("Firebase environment scaffolded successfully.")
	return nil
}
