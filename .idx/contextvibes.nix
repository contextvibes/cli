# -----------------------------------------------------------------------------
# Package: ContextVibes CLI
# Version: Dynamic (Defaults to 0.7.0-alpha.2, overrides via local.nix)
# -----------------------------------------------------------------------------
{ pkgs, overrides ? {} }:

let
  # --- Defaults (Tracked in Git) ---
  defaultVersion = "0.7.0-alpha.2";
  defaultHash    = "0n1mchl9nphy1q0rxi9468y2hyp6brvf9iisflycvni20w3v4c4i";

  # --- Resolution Logic ---
  # Use value from local.nix if present, otherwise use default
  version = overrides.CONTEXTVIBES_VERSION or defaultVersion;
  sha256  = overrides.CONTEXTVIBES_HASH    or defaultHash;

  # Map Nix system architecture to Go architecture naming
  arch = if pkgs.stdenv.hostPlatform.isAarch64 then "arm64" else "amd64";
  os = "linux"; # IDX is Linux-based
in
pkgs.stdenv.mkDerivation rec {
  pname = "contextvibes";
  inherit version;

  src = pkgs.fetchurl {
    # URL format matches GoReleaser: contextvibes_0.7.0-alpha.2_linux_amd64.tar.gz
    url = "https://github.com/contextvibes/cli/releases/download/v${version}/contextvibes_${version}_${os}_${arch}.tar.gz";
    inherit sha256;
  };

  # The archive unpacks into the current directory
  sourceRoot = ".";

  installPhase = ''
    mkdir -p $out/bin
    install -m 755 contextvibes $out/bin/contextvibes
  '';
}
