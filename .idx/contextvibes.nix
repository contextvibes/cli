# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "ae3bc6e5065747e60d22dec590e55fb7897b6633";
    hash = "sha256-VYuzgxtJMBWc+qzE0v3cZG7duYQ8fyDldp2F6vGvTiQ=";
  };

  vendorHash = null;
  subPackages = [ "cmd/contextvibes" ];
}