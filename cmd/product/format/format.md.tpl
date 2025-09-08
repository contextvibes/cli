# Applies code formatting and auto-fixes linter issues.

Detects project type (Go, Python, Terraform) and applies standard formatting
and auto-fixable linter suggestions, modifying files in place. This is the primary
command for remediating code quality issues.

- Go: Runs 'golangci-lint run --fix', which applies all configured formatters and linters.
- Python: Runs 'isort .' and 'black .'.
- Terraform: Runs 'terraform fmt -recursive .'.
