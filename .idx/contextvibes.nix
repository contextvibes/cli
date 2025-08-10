# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "c6bbd35d451b68185397351fcd6b742142bddfcc";
    hash = "sha256-+/6eZFQpSzhPsXmA/+RLXsEopOFnMY1UEP5khUClA2E=";
  };

  vendorHash = "sha256-Z89OrUXlGAu39ncl6R/MkADPnQje9xELxBT5Rl9QL3w=";

  subPackages = [ "cmd/contextvibes" ];
}