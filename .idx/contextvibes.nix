# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "92503c4e1debd8cb36248e26cc9341cb5831bb91";
    hash = "sha256-Zos8r3IOz4lzN5Yox6TNtFus8pm955IxdHsbZCPos4g=";
  };

  vendorHash = null;
  subPackages = [ "cmd/cv" ];
}
