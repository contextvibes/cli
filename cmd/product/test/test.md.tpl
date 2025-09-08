# Runs project-specific tests (e.g., go test, pytest).

Detects the project type (Go, Python) and runs the appropriate test command.
Any arguments passed to 'contextvibes product test' will be forwarded to the underlying test runner.

- Go: Runs 'go test ./...'
- Python: Runs 'pytest' (if available). Falls back to 'python -m unittest discover' if pytest not found.

For other project types, or if no specific test runner is found, it will indicate no action.
