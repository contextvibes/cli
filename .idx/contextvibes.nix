# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "c293095523fe5b5d9440c4577db3c24e18acb2c1";
    hash = "sha256-fNqyT2BJSe3olQ7C+lc1EEXwb3Tv6dvL9obLXpwnCUc=";
  };

  vendorHash = null;
  subPackages = [ "cmd/contextvibes" ];
}