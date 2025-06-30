package utils

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var OpenLocalFile = func(p string) (*os.File, error) {
	// Check if the file exists
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist: %w", p, err)
	}

	// Check if the file is a directory
	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
		return nil, fmt.Errorf("path %s is a directory, not a file", p)
	}

	// Open the file for reading
	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", p, err)
	}

	return file, nil
}

var WriteLocalFile = func(p string, content []byte, perm os.FileMode) error {
	// Create directories leading up to the file
	if err := os.MkdirAll(filepath.Dir(p), perm); err != nil {
		log.Errorf("failed to create directories for path %s: %v", p, err)
		return fmt.Errorf("failed to create path %s: %w", p, err)
	}

	// Write the content to the file
	if err := os.WriteFile(p, content, perm); err != nil {
		log.Errorf("failed to write file %s: %v", p, err)
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
		log.Errorf("path %s is a directory, not a file", p)
		return fmt.Errorf("path %s is a directory, not a file", p)
	}

	// Remove the file
	if err := os.Remove(p); err != nil {
		log.Errorf("failed to delete file %s: %v", p, err)
		return fmt.Errorf("failed to delete file %s: %w", p, err)
	}

	return nil
}
