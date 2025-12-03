#.idx/contextvibes.nix
{ pkgs }:

pkgs.stdenv.mkDerivation {
  pname = "contextvibes";
  version = "0.2.1";

  # We wrap the fetchurl call in parentheses to apply overrideAttrs.
  # This modifies the fixed-output derivation created by fetchurl.
  src = (pkgs.fetchurl {
    # URL for the release asset.
    url = "https://github.com/contextvibes/cli/releases/download/v0.4.1-rc3/contextvibes";
    # SHA256 hash of the downloaded file.
    sha256 = "sha256:1aa09c34c750056e78a69e84f6d7c38a0c6220e3dc46af460256f9a38978ff8a";
  }).overrideAttrs (finalAttrs: previousAttrs: {
    # Enable structured attributes to allow passing complex sets.
    __structuredAttrs = true;

    # The Critical Fix:
    # Explicitly instruct Nix to ignore any store path references found in the downloaded file.
    # 'out' refers to the default output of the fetchurl derivation.
    unsafeDiscardReferences = {
      out = true;
    };
  });

  dontUnpack = true;

  # Install the binary into the output directory.
  installPhase = ''
    mkdir -p $out/bin
    install -m 755 -D $src $out/bin/contextvibes
  '';
}
