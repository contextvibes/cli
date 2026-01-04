# -----------------------------------------------------------------------------
# Package: ContextVibes CLI
# Version: 0.6.0
# -----------------------------------------------------------------------------
{ pkgs }:

pkgs.stdenv.mkDerivation rec {
  name = "contextvibes-${version}";
  version = "0.6.0";

  src = pkgs.fetchurl {
    url = "https://github.com/contextvibes/cli/releases/download/v${version}/contextvibes";
    sha256 = "sha256:bdbf55bf902aa567851fcbbc07704b416dee85065a276a47e7df19433c5643ea";
  };

  dontUnpack = true;

  installPhase = ''
    mkdir -p $out/bin
    install -m 755 $src $out/bin/contextvibes
  '';
}
