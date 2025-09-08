# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "8c9fe4f307fea4234bdd238913153753a5a7a688";
    hash = "sha256-KXmq00huV+e9LIzLt5xhKQF/HfCCg1GrY1BYkSv4MJo=";
  };

  vendorHash = null;
  subPackages = [ "cmd/contextvibes" ];
}
