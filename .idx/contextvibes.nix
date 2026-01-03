# -----------------------------------------------------------------------------
# Package: ContextVibes CLI
# Version: 0.7.0-alpha.2
# -----------------------------------------------------------------------------
{ pkgs }:

let
  # Map Nix system architecture to Go architecture naming
  arch = if pkgs.stdenv.hostPlatform.isAarch64 then "arm64" else "amd64";
  os = "linux"; # IDX is Linux-based
in
pkgs.stdenv.mkDerivation rec {
  pname = "contextvibes";
  version = "0.7.0-alpha.2";

  src = pkgs.fetchurl {
    # URL format matches GoReleaser: contextvibes_0.7.0-alpha.2_linux_amd64.tar.gz
    url = "https://github.com/contextvibes/cli/releases/download/v${version}/contextvibes_${version}_${os}_${arch}.tar.gz";

    # TODO: Update this hash!
    # Run: nix-prefetch-url https://github.com/contextvibes/cli/releases/download/v0.7.0-alpha.2/contextvibes_0.7.0-alpha.2_linux_amd64.tar.gz
    sha256 = "0n1mchl9nphy1q0rxi9468y2hyp6brvf9iisflycvni20w3v4c4i";
  };

  # The archive unpacks into the current directory
  sourceRoot = ".";

  installPhase = ''
    mkdir -p $out/bin
    install -m 755 contextvibes $out/bin/contextvibes
  '';
}
