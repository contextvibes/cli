# .idx/contextvibes.nix
{ pkgs }:

pkgs.stdenv.mkDerivation {
  pname = "contextvibes";
  version = "0.5.0";

  src = pkgs.fetchurl {
    url = "https://github.com/contextvibes/cli/releases/download/v0.5.0/contextvibes";
    sha256 = "sha256:c519ee03b6b77721dfc78bb03b638c3327096affafd8968d49b2bbd9a89ffc10";
  };

  dontUnpack = true;

  installPhase = ''
    mkdir -p $out/bin
    install -m 755 -D $src $out/bin/contextvibes
  '';
}
