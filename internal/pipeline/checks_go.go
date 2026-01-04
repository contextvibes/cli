package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
)

// GoVetCheck runs go vet.
type GoVetCheck struct {
	Paths []string
}

func (c *GoVetCheck) Name() string { return "Go Vet" }

func (c *GoVetCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("go") {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "Go not found"}
	}

	args := []string{"vet"}

	if len(c.Paths) > 0 {
		for _, p := range c.Paths {
			if !strings.HasPrefix(p, ".") && !strings.HasPrefix(p, "/") {
				p = "./" + p
			}

			args = append(args, p)
		}
	} else {
		args = append(args, "./...")
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", args...)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "go vet found issues",
			Error:   err,
			Details: stdout + stderr,
		}
	}

	return Result{Name: c.Name(), Status: StatusPass, Message: "Code passes go vet"}
}

// GolangCILintCheck runs the linter with a specific configuration mode.
type GolangCILintCheck struct {
	Paths      []string
	ConfigType config.AssetType // The specific lens to use
}

func (c *GolangCILintCheck) Name() string {
	return "GolangCI-Lint (" + string(c.ConfigType) + ")"
}

func (c *GolangCILintCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("golangci-lint") {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "golangci-lint not found"}
	}

	args := []string{"run"}

	var tempConfigFile string

	// If a specific config type is requested (not empty/default), load it.
	if c.ConfigType != "" {
		configBytes, err := config.GetLanguageAsset("go", c.ConfigType)
		if err != nil {
			return Result{Name: c.Name(), Status: StatusFail, Message: "Failed to load config asset", Error: err}
		}

		tmpFile, err := os.CreateTemp(".", ".golangci-"+string(c.ConfigType)+"-*.yml")
		if err != nil {
			return Result{Name: c.Name(), Status: StatusFail, Message: "Failed to create temp config", Error: err}
		}

		tempConfigFile = tmpFile.Name()

		defer func() {
			_ = os.Remove(tempConfigFile)
		}()

		if _, err := tmpFile.Write(configBytes); err != nil {
			_ = tmpFile.Close()

			return Result{Name: c.Name(), Status: StatusFail, Message: "Failed to write temp config", Error: err}
		}

		_ = tmpFile.Close()

		args = append(args, "-c", tempConfigFile)
	}

	if len(c.Paths) > 0 {
		args = append(args, c.Paths...)
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "golangci-lint", args...)
	if err != nil {
		if tempConfigFile != "" {
			stderr = strings.ReplaceAll(stderr, tempConfigFile, ".golangci.yml (generated)")
		}

		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Linter found issues",
			Error:   err,
			Details: stdout + stderr,
		}
	}

	return Result{Name: c.Name(), Status: StatusPass, Message: "Linter passed"}
}

// GoVulnCheck runs vulnerability scanning.
type GoVulnCheck struct {
	Paths []string
}

func (c *GoVulnCheck) Name() string { return "Go Vulnerability Check" }

func (c *GoVulnCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("go") {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "Go not found"}
	}

	args := []string{}

	if len(c.Paths) > 0 {
		for _, p := range c.Paths {
			if strings.HasSuffix(p, ".go") {
				p = filepath.Dir(p)
			}

			if !strings.HasPrefix(p, ".") && !strings.HasPrefix(p, "/") {
				p = "./" + p
			}

			args = append(args, p)
		}
	} else {
		args = append(args, "./...")
	}

	// Use `go run` to ensure we use the version of govulncheck compatible with the project's go version.
	runArgs := []string{"run", "golang.org/x/vuln/cmd/govulncheck@latest"}
	runArgs = append(runArgs, args...)

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", runArgs...)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Vulnerabilities found",
			Error:   err,
			Details: stdout + stderr,
		}
	}

	return Result{Name: c.Name(), Status: StatusPass, Message: "No known vulnerabilities found"}
}

// GitleaksCheck scans for secrets.
type GitleaksCheck struct{}

func (c *GitleaksCheck) Name() string { return "Secret Scanning (gitleaks)" }

func (c *GitleaksCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("gitleaks") {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "gitleaks not found"}
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "gitleaks", "detect", "--no-git", "--verbose", "-c", ".gitleaks.toml")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Secrets detected!",
			Error:   err,
			Details: stdout + stderr,
		}
	}

	return Result{Name: c.Name(), Status: StatusPass, Message: "No secrets detected"}
}

// DeadcodeCheck finds unreachable code.
type DeadcodeCheck struct{}

func (c *DeadcodeCheck) Name() string { return "Dead Code Analysis" }

func (c *DeadcodeCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("deadcode") {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "deadcode not found"}
	}

	stdout, _, err := execClient.CaptureOutput(ctx, ".", "deadcode", "-test", "./...")
	if err != nil {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "Dead code analysis failed", Error: err}
	}

	if len(stdout) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Unreachable code detected",
			Details: stdout,
		}
	}

	return Result{Name: c.Name(), Status: StatusPass, Message: "No dead code detected"}
}

// GoBuildCheck runs the Go compiler to check for syntax and type errors.
type GoBuildCheck struct {
	Paths []string
}

func (c *GoBuildCheck) Name() string { return "Go Compiler (Build)" }

func (c *GoBuildCheck) Run(ctx context.Context, execClient *exec.ExecutorClient) Result {
	if !execClient.CommandExists("go") {
		return Result{Name: c.Name(), Status: StatusWarn, Message: "Go not found"}
	}

	// We use -o /dev/null (or NUL on Windows) to check compilation without writing a binary.
	// However, to be cross-platform safe and simple, we can just run build on the packages.
	// If we are at root, "go build ./..." is standard.
	args := []string{"build", "-v"}

	if len(c.Paths) > 0 {
		for _, p := range c.Paths {
			if !strings.HasPrefix(p, ".") && !strings.HasPrefix(p, "/") {
				p = "./" + p
			}

			args = append(args, p)
		}
	} else {
		args = append(args, "./...")
	}

	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", args...)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Compilation failed",
			Error:   err,
			Details: stdout + stderr,
		}
	}

	return Result{Name: c.Name(), Status: StatusPass, Message: "Code compiles successfully"}
}
