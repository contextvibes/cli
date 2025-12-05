# Force rebuilds and installs development tools.

This command addresses environment mismatches where tools provided by Nix (like
`govulncheck` or `golangci-lint`) may be compiled with an older Go version than
the one currently active in the shell.

It performs the following:
1. Verifies the current Go environment.
2. Ensures `$HOME/go/bin` is prepended to your `PATH` in `.bashrc` (to prioritize local tools).
3. Force-reinstalls standard tools using `go install -a`, ensuring they are compiled with the current Go version.
