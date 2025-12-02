// Package project provides functionality to detect the type of a project based on its file structure.
package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Detect determines the project type based on files in the directory.
func Detect(dir string) (Type, error) {
	checks := []struct {
		typ   Type
		check func(string) (bool, error)
	}{
		{Terraform, func(d string) (bool, error) { return hasFiles(d, "*.tf") }},
		{Pulumi, func(d string) (bool, error) { return fileExists(d, "Pulumi.yaml") }},
		{Go, func(d string) (bool, error) { return fileExists(d, "go.mod") }},
		{Python, func(d string) (bool, error) { return fileExists(d, "requirements.txt") }},
		{Python, func(d string) (bool, error) { return fileExists(d, "pyproject.toml") }},
	}

	for _, checkItem := range checks {
		exists, err := checkItem.check(dir)
		if err != nil {
			return Unknown, err
		}

		if exists {
			return checkItem.typ, nil
		}
	}

	return Unknown, nil
}

func hasFiles(dir, pattern string) (bool, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return false, fmt.Errorf("error checking for %s files: %w", pattern, err)
	}

	return len(matches) > 0, nil
}

func fileExists(dir, filename string) (bool, error) {
	path := filepath.Join(dir, filename)

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("error checking for %s: %w", filename, err)
}
