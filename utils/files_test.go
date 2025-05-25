package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteFile_CreatesFileAndWritesContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "testfile.txt")
	content := []byte("hello world")

	err := WriteFile(filePath, content, 0o644)
	assert.NoError(t, err, "WriteFile should not return an error")
	assert.FileExists(t, filePath, "file should exist after WriteFile")

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "ReadFile should not return an error")
	assert.Equal(t, content, data, "file content does not match expected data")
}

func TestWriteFile_InvalidPath(t *testing.T) {
	// On most systems, writing to an invalid path like "" should fail
	err := WriteFile("", []byte("data"), 0o644)
	assert.Error(t, err, "expected an error for invalid path")
}

func TestWriteFile_PermissionDenied(t *testing.T) {
	// Try to write to a directory as a file, which should fail
	tmpDir := t.TempDir()
	err := WriteFile(tmpDir, []byte("data"), 0o644)
	assert.Error(t, err, "expected an error when writing to a directory")
}
