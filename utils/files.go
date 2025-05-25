package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

var WriteFile = func(p string, content []byte, perm os.FileMode) error {
	// Create directories leading up to the file
	if err := os.MkdirAll(filepath.Dir(p), perm); err != nil {
		return fmt.Errorf("failed to create path %s: %w", p, err)
	}

	// Write the content to the file
	if err := os.WriteFile(p, content, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", p, err)
	}

	return nil
}
