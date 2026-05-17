package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func loadInstruction(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}
	if filepath.IsAbs(path) || !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read instruction file %q: %w", path, err)
	}

	// Support running either from agent/ or from the repository root.
	repoRootPath := filepath.Join("agent", path)
	data, fallbackErr := os.ReadFile(repoRootPath)
	if fallbackErr != nil {
		return "", fmt.Errorf("read instruction file %q or %q: %w", path, repoRootPath, fallbackErr)
	}
	return string(data), nil
}
