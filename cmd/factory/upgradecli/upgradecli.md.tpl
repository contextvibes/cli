# Upgrades the ContextVibes CLI version in the Nix environment.

Checks GitHub for the latest release of the CLI. If a newer version is available:
1. Calculates the new SHA256 hash using 'nix-prefetch-url'.
2. Updates '.idx/contextvibes.nix' with the new version and hash.
3. Prompts you to rebuild the environment.
