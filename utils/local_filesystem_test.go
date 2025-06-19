package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteLocalFile_CreatesFileAndWritesContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "testfile.txt")
	content := []byte("hello world")

	err := WriteLocalFile(filePath, content, 0o644)
	assert.NoError(t, err, "WriteFile should not return an error")
	assert.FileExists(t, filePath, "file should exist after WriteFile")

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "ReadFile should not return an error")
	assert.Equal(t, content, data, "file content does not match expected data")
}

func TestWriteFile_InvalidPath(t *testing.T) {
	// On most systems, writing to an invalid path like "" should fail
	err := WriteLocalFile("", []byte("data"), 0o644)
	assert.Error(t, err, "expected an error for invalid path")
}

func TestWriteFile_PermissionDenied(t *testing.T) {
	// Try to write to a directory as a file, which should fail
	tmpDir := t.TempDir()
	err := WriteLocalFile(tmpDir, []byte("data"), 0o644)
	assert.Error(t, err, "expected an error when writing to a directory")
}
func TestDeleteLocalFile_DeletesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "testfile.txt")
	content := []byte("to be deleted")

	// Create the file first
	err := os.WriteFile(filePath, content, 0o644)
	assert.NoError(t, err, "setup: should be able to create file")

	// Delete the file
	err = DeleteLocalFile(filePath)
	assert.NoError(t, err, "DeleteLocalFile should not return an error for existing file")
	assert.NoFileExists(t, filePath, "file should not exist after DeleteLocalFile")
}

func TestDeleteLocalFile_FileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "doesnotexist.txt")

	err := DeleteLocalFile(nonExistentFile)
	assert.Error(t, err, "expected error when deleting non-existent file")
	assert.Contains(t, err.Error(), "does not exist", "error message should mention file does not exist")
}

func TestDeleteLocalFile_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	err := DeleteLocalFile(tmpDir)
	assert.Error(t, err, "expected error when trying to delete a directory as a file")
}
