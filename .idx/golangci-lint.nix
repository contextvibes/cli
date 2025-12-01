# .idx/golangci-lint.nix
# Defines a Nix package for a specific, precompiled version of golangci-lint.

{ pkgs }:

pkgs.stdenv.mkDerivation {
  # --- PACKAGE METADATA ---
  pname = "golangci-lint-bin";
  version = "2.6.2";

  # --- SOURCE FETCHING ---
  # WHAT: Fetch the precompiled binary archive from its GitHub release URL.
  # WHY:  To get the tool without needing to build it from source.
  src = pkgs.fetchurl {
    url = "https://github.com/golangci/golangci-lint/releases/download/v2.6.2/golangci-lint-2.6.2-linux-amd64.tar.gz";
    
    # WHAT: A cryptographic hash of the downloaded file.
    # WHY:  Ensures the binary is exactly what we expect, providing security and reproducibility.
    sha256 = "sha256-SZyGS1/ZhBxPqOgLXivjD3Pwhc8YbxsRH/gaJ4O33hI=";
  };

  # --- INSTALLATION SCRIPT ---
  # WHAT: A script to copy the binary into the Nix store.
  # WHY:  To make the executable available in the environment's PATH.
  installPhase = ''
    mkdir -p $out/bin
    install -m 755 golangci-lint $out/bin/
  '';
}
