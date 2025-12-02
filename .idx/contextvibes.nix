# .idx/contextvibes.nix
{ pkgs }:

pkgs.stdenv.mkDerivation {
  pname = "contextvibes";
  version = "v0.4.1-rc3";

  src = pkgs.fetchurl {
    url = "https://github.com/contextvibes/cli/releases/download/v0.4.1-rc3/contextvibes";
    sha256 = "1aa09c34c750056e78a69e84f6d7c38a0c6220e3dc46af460256f9a38978ff8a";
  };

  dontUnpack = true;

  installPhase = ''
    mkdir -p $out/bin
    install -m 755 -D $src $out/bin/contextvibes
  '';
}
