package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
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
	assert.NoError(t, err, "DeleteLocalFile should not return an error for existing file")
	assert.NoFileExists(t, nonExistentFile, "file should not exist after DeleteLocalFile")
}

func TestDeleteLocalFile_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	err := DeleteLocalFile(tmpDir)
	assert.Error(t, err, "expected error when trying to delete a directory as a file")
}
func TestOpenLocalFile_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "testfile.txt")
	content := []byte("test content")
	err := os.WriteFile(filePath, content, 0o644)
	assert.NoError(t, err, "setup: should be able to create file")

	f, err := OpenLocalFile(filePath)
	if f != nil {
		defer f.Close() // ensure we close the file if it was opened
	}
	assert.NoError(t, err, "OpenLocalFile should not return error for existing file")
	assert.NotNil(t, f, "file handle should not be nil")

	readContent := make([]byte, len(content))
	n, err := f.Read(readContent)
	assert.NoError(t, err, "should be able to read file")
	assert.Equal(t, len(content), n, "read bytes count mismatch")
	assert.Equal(t, content, readContent, "file content mismatch")
}

func TestOpenLocalFile_FileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "doesnotexist.txt")

	f, err := OpenLocalFile(nonExistentFile)
	if f != nil {
		defer f.Close() // ensure we close the file if it was opened
	}
	assert.Error(t, err, "should return error for non-existent file")
	assert.Nil(t, f, "file handle should be nil for non-existent file")
}

func TestOpenLocalFile_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	f, err := OpenLocalFile(tmpDir)
	if f != nil {
		defer f.Close() // ensure we close the file if it was opened
	}
	assert.Error(t, err, "should return error when path is a directory")
	assert.Nil(t, f, "file handle should be nil for directory path")
}

func TestOpenLocalFile_PermissionDenied(t *testing.T) {
	// This test is best-effort and may not work on all OSes, but we try to create a file with no read permission
	err := godotenv.Load("../.env")
	assert.NoError(t, err)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "noperm.txt")
	content := []byte("secret")
	err = os.WriteFile(filePath, content, 0o200) // write-only
	assert.NoError(t, err, "setup: should be able to create file")

	f, err := OpenLocalFile(filePath)
	if f != nil {
		defer f.Close() // ensure we close the file if it was opened
	}
	if os.Getenv("CI") == "" { // skip assertion in CI where permissions may not be enforced
		assert.Error(t, err, "should return error when lacking read permission")
		assert.Nil(t, f, "file handle should be nil when lacking read permission")
	}
}
