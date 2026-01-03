# Running Code Quality Checks.

Executes a suite of static analysis and quality checks appropriate for the detected project type.
For Go projects, this includes `go mod tidy`, `go vet`, `golangci-lint`, and `govulncheck`.

You can optionally provide file or package paths to run checks on specific targets.

**Examples:**
```bash
# Run on the entire project (default)
contextvibes product quality

# Run on a specific package
contextvibes product quality cmd/factory/...

# Run on a specific file (if supported by the underlying tool)
contextvibes product quality main.go
```
