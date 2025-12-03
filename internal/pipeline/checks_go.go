package pipeline

import (
	"context"

	"github.com/contextvibes/cli/internal/exec"
)

// GoModTidyCheck verifies dependencies are tidy.
type GoModTidyCheck struct{}

// Name returns the name of the check.
func (c *GoModTidyCheck) Name() string { return "Go Module Tidy" }

// Run executes the check.
func (c *GoModTidyCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("go") {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Go not found",
			Error:   nil,
			Advice:  "",
			Details: "",
		}
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", "mod", "tidy")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "go mod tidy failed",
			Error:   err,
			Advice:  "Run 'go mod tidy' manually to fix dependencies.",
			Details: stdout + stderr,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Dependencies are tidy",
		Error:   nil,
		Advice:  "",
		Details: "",
	}
}

// GoVetCheck runs go vet.
type GoVetCheck struct{}

// Name returns the name of the check.
func (c *GoVetCheck) Name() string { return "Go Vet" }

// Run executes the check.
func (c *GoVetCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("go") {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Go not found",
			Error:   nil,
			Advice:  "",
			Details: "",
		}
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", "vet", "./...")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "go vet found issues",
			Error:   err,
			Advice:  "Run 'go vet ./...' to see details.",
			Details: stdout + stderr,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Code passes go vet",
		Error:   nil,
		Advice:  "",
		Details: "",
	}
}

// GolangCILintCheck runs the linter.
type GolangCILintCheck struct{}

// Name returns the name of the check.
func (c *GolangCILintCheck) Name() string { return "GolangCI-Lint" }

// Run executes the check.
func (c *GolangCILintCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("golangci-lint") {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "golangci-lint not found",
			Error:   nil,
			Advice:  "Install golangci-lint for better quality checks.",
			Details: "",
		}
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "golangci-lint", "run")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Linter found issues",
			Error:   err,
			Advice:  "Run 'contextvibes product format' to fix some issues automatically.",
			Details: stdout + stderr,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Linter passed",
		Error:   nil,
		Advice:  "",
		Details: "",
	}
}

// GoVulnCheck runs vulnerability scanning.
type GoVulnCheck struct{}

// Name returns the name of the check.
func (c *GoVulnCheck) Name() string { return "Go Vulnerability Check" }

// Run executes the check.
func (c *GoVulnCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("govulncheck") {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "govulncheck not found",
			Error:   nil,
			Advice:  "Install govulncheck to scan for security vulnerabilities.",
			Details: "",
		}
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "govulncheck", "./...")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Vulnerabilities found",
			Error:   err,
			Advice:  "Update dependencies to resolve known vulnerabilities.",
			Details: stdout + stderr,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "No known vulnerabilities found",
		Error:   nil,
		Advice:  "",
		Details: "",
	}
}

// GitleaksCheck scans for secrets.
type GitleaksCheck struct{}

// Name returns the name of the check.
func (c *GitleaksCheck) Name() string { return "Secret Scanning (gitleaks)" }

// Run executes the check.
func (c *GitleaksCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("gitleaks") {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "gitleaks not found",
			Error:   nil,
			Advice:  "Install gitleaks to prevent committing secrets.",
			Details: "",
		}
	}

	// Use -c to point to the config file we just created
	stdout, stderr, err := execClient.CaptureOutput(
		ctx,
		".",
		"gitleaks",
		"detect",
		"--no-git",
		"--verbose",
		"-c",
		".gitleaks.toml",
	)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Secrets detected!",
			Error:   err,
			Advice:  "Check output for leaked secrets and revoke them immediately.",
			Details: stdout + stderr,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "No secrets detected",
		Error:   nil,
		Advice:  "",
		Details: "",
	}
}

// DeadcodeCheck finds unreachable code.
type DeadcodeCheck struct{}

// Name returns the name of the check.
func (c *DeadcodeCheck) Name() string { return "Dead Code Analysis" }

// Run executes the check.
func (c *DeadcodeCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("deadcode") {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "deadcode not found",
			Error:   nil,
			Advice:  "Run 'go install golang.org/x/tools/cmd/deadcode@latest'",
			Details: "",
		}
	}

	stdout, _, err := execClient.CaptureOutput(ctx, ".", "deadcode", "./...")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Dead code analysis failed",
			Error:   err,
			Advice:  "",
			Details: "",
		}
	}

	if len(stdout) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Unreachable code detected",
			Error:   nil,
			Advice:  "Run 'deadcode ./...' to see unused functions.",
			Details: stdout,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "No dead code detected",
		Error:   nil,
		Advice:  "",
		Details: "",
	}
}
