# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "953256dfd1d1e66fadc47938b8c9e738d7d5a5af";
    hash = "sha256-opDH5s0hKFKAIUhSEl8qgOPW92Es6NF4FB0eX1vvQQ4=";
  };

  vendorHash = null;
  subPackages = [ "cmd/contextvibes" ];
}