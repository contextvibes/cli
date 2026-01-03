# -----------------------------------------------------------------------------
# Package: GolangCI-Lint (Precompiled)
# Version: 2.7.2
# -----------------------------------------------------------------------------
{ pkgs }:

pkgs.stdenv.mkDerivation rec {
  name = "golangci-lint-bin-${version}";
  version = "2.7.2";

  src = pkgs.fetchurl {
    url = "https://github.com/golangci/golangci-lint/releases/download/v${version}/golangci-lint-${version}-linux-amd64.tar.gz";
    sha256 = "sha256-zkah8diQ57ZnJZ9wuyNil/XPh5GptrmLQbKD2Ttbbog=";
  };

  # The builder automatically enters the extracted folder, so the binary is just 'golangci-lint'
  installPhase = ''
    mkdir -p $out/bin
    install -m 755 golangci-lint $out/bin/
  '';
}
