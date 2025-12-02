# .idx/contextvibes.nix
# Downloads a pre-compiled binary from a GitHub Release.
{ pkgs }:

pkgs.stdenv.mkDerivation {
  pname = "contextvibes";
  version = "0.4.0";

  # Fetch the pre-built binary from the GitHub Release.
  src = pkgs.fetchurl {
    # URL for the release asset.
    url = "https://github.com/contextvibes/cli/releases/download/v0.4.0/contextvibes";
    # SHA256 hash of the downloaded file.
    sha256 = "sha256:3a6a5196c90a5e2dc910d1c819e450246f47127855b79a11869f5c6d3274ca6f";
  };

  dontUnpack = true;

  # Install the binary into the output directory.
  # $src refers to the downloaded file.
  # $out is the destination path in the Nix store.
  installPhase = ''
    mkdir -p $out/bin
    install -m 755 -D $src $out/bin/contextvibes
  '';
}
