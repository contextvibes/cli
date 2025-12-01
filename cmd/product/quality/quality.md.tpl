# Running Code Quality Checks.

Executes a suite of static analysis and quality checks appropriate for the detected project type.
For Go projects, this includes `go mod tidy`, `go vet`, `golangci-lint`, and `govulncheck`.
