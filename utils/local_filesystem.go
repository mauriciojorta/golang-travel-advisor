package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

var WriteLocalFile = func(p string, content []byte, perm os.FileMode) error {
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

var DeleteLocalFile = func(p string) error {
	// Check if the file exists.
	if _, err := os.Stat(p); os.IsNotExist(err) {
		// If it doesn't, then it is already deleted, so we consider the operation successful.
		return nil
	}

	// Check if the file is a directory
	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
		return fmt.Errorf("path %s is a directory, not a file", p)
	}

	// Remove the file
	if err := os.Remove(p); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", p, err)
	}

	return nil
}
