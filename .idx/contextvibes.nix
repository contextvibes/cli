# -----------------------------------------------------------------------------
# Package: ContextVibes CLI (Hybrid: Binary or Source)
# -----------------------------------------------------------------------------
{ pkgs,
  # Defaults (Binary Mode)
  buildType ? "binary",   # "binary" or "source"
  version ? "0.6.0",      # The tag or version string

  # Binary Specific
  binHash ? "sha256-bdbf55bf902aa567851fcbbc07704b416dee85065a276a47e7df19433c5643ea",

  # Source Specific (Required if buildType == "source")
  rev ? "",               # Commit hash or branch name (e.g., "main")
  srcHash ? "",           # Hash of the source code
  vendorHash ? ""         # Hash of the Go modules (go.mod/sum)
}:

if buildType == "source" then
  # --- Option A: Build from Source ---
  pkgs.buildGoModule {
    pname = "contextvibes";
    version = version; # e.g., "unstable-${rev}"

    src = pkgs.fetchFromGitHub {
      owner = "contextvibes";
      repo = "cli";
      rev = rev;
      hash = srcHash;
    };

    vendorHash = vendorHash;

    # Disable tests during build to speed it up (optional)
    doCheck = false;
  }

else
  # --- Option B: Download Binary (Default) ---
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
