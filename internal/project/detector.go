package project

import (
	"fmt"
	"os"
	"path/filepath"
)

func Detect(dir string) (Type, error) {
	tfFiles, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return Unknown, fmt.Errorf("error checking for Terraform files: %w", err)
	}
	if len(tfFiles) > 0 {
		return Terraform, nil
	}

	pulumiYamlPath := filepath.Join(dir, "Pulumi.yaml")
	if _, err := os.Stat(pulumiYamlPath); err == nil {
		return Pulumi, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for Pulumi.yaml: %w", err)
	}

	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		return Go, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for go.mod: %w", err)
	}

	pyReqPath := filepath.Join(dir, "requirements.txt")
	pyProjPath := filepath.Join(dir, "pyproject.toml")
	if _, err := os.Stat(pyReqPath); err == nil {
		return Python, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for requirements.txt: %w", err)
	}
	if _, err := os.Stat(pyProjPath); err == nil {
		return Python, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for pyproject.toml: %w", err)
	}

	return Unknown, nil
}
