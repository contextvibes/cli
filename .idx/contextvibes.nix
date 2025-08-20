# .idx/contextvibes.nix
{ pkgs }:

pkgs.buildGoModule {
  pname = "contextvibes";
  version = "0.2.0-dev";

  src = pkgs.fetchFromGitHub {
    owner = "contextvibes";
    repo = "cli";
    rev = "ed52c883416b7219d022c82779f2963fb2a7cab7";
    hash = "sha256-u7ci0hqmWPqYsT9LnPPVj4LmHjTuVG9zz+z/jF55V60=";
  };

  vendorHash = null;
  subPackages = [ "cmd/contextvibes" ];
}
