# Bootstraps the development environment and installs the CLI.

This is the universal entry point for setting up ContextVibes. It performs
the following sequence:
1.  **Path Configuration:** Ensures `$HOME/go/bin` is in your `.bashrc`.
2.  **CLI Installation:** Runs `go install` to place the binary in your path.
3.  **Toolchain Setup:** Rebuilds essential Go tools (govulncheck, etc.).
4.  **Environment Scaffolding:** Generates `.idx/` and `.vscode/` configurations.

After running this, you simply need to run `source ~/.bashrc` to begin.
