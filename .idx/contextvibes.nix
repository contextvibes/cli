# .idx/contextvibes.nix
# Downloads a pre-compiled binary from a GitHub Release.
{ pkgs }:

pkgs.stdenv.mkDerivation {
  pname = "contextvibes";
  version = "0.2.1";

  # Fetch the pre-built binary from the GitHub Release.
  src = pkgs.fetchurl {
    # URL for the release asset.
    url = "https://github.com/contextvibes/cli/releases/download/v0.2.1/contextvibes";
    # SHA256 hash of the downloaded file.
    sha256 = "sha256:a7a69036bf2d9414f033a6c1bb54fa7faef4c3e0d92f0653c6ce8434314d9264";
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
