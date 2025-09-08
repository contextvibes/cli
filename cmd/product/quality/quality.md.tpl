# Runs a comprehensive suite of code quality checks.

Detects project type (Go, Python, Terraform) and runs a suite of formatters,
linters, and vulnerability scanners in a read-only "check" mode.

- Go: Verifies 'go mod' is tidy, runs 'go vet', runs 'golangci-lint', and scans for
  vulnerabilities with 'govulncheck'. All formatting checks are handled by 'golangci-lint'.
- Python: Runs 'isort --check', 'black --check', 'flake8'.
- Terraform: Runs 'terraform fmt -check', 'terraform validate', 'tflint'.

This command acts as a quality gate and will fail if any issues are found. To fix
many of the reported issues automatically, run 'contextvibes product format'.
