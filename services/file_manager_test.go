package services

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"example.com/travel-advisor/utils"
)

// --- Mocks for utils package ---

var (
	mockWriteLocalFileCalled  bool
	mockDeleteLocalFileCalled bool
	mockOpenLocalFileCalled   bool
	mockWriteLocalFileArgs    struct {
		filepath string
		content  []byte
		perm     os.FileMode
	}
	mockDeleteLocalFileArg string
	mockOpenLocalFileArg   string
	mockOpenLocalFileRet   *os.File
	mockOpenLocalFileErr   error
)

func mockWriteLocalFile(filepath string, content []byte, perm os.FileMode) error {
	mockWriteLocalFileCalled = true
	mockWriteLocalFileArgs.filepath = filepath
	mockWriteLocalFileArgs.content = content
	mockWriteLocalFileArgs.perm = perm
	return nil
}

func mockDeleteLocalFile(filepath string) error {
	mockDeleteLocalFileCalled = true
	mockDeleteLocalFileArg = filepath
	return nil
}

func mockOpenLocalFile(filepath string) (*os.File, error) {
	mockOpenLocalFileCalled = true
	mockOpenLocalFileArg = filepath
	return mockOpenLocalFileRet, mockOpenLocalFileErr
}

func resetMocks() {
	mockWriteLocalFileCalled = false
	mockDeleteLocalFileCalled = false
	mockOpenLocalFileCalled = false
	mockOpenLocalFileRet = nil
	mockOpenLocalFileErr = nil
}

// --- Patch utils functions ---

func init() {
	utils.WriteLocalFile = mockWriteLocalFile
	utils.DeleteLocalFile = mockDeleteLocalFile
	utils.OpenLocalFile = mockOpenLocalFile
}

// --- Tests ---

func TestLocalFileManager_SaveContentInFile(t *testing.T) {
	resetMocks()
	lfm := &LocalFileManager{}
	content := "test content"
	filepath := "test.txt"

	err := lfm.SaveContentInFile(filepath, &content)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !mockWriteLocalFileCalled {
		t.Error("expected WriteLocalFile to be called")
	}
	if mockWriteLocalFileArgs.filepath != filepath {
		t.Errorf("expected filepath %s, got %s", filepath, mockWriteLocalFileArgs.filepath)
	}
	if string(mockWriteLocalFileArgs.content) != content {
		t.Errorf("expected content %s, got %s", content, string(mockWriteLocalFileArgs.content))
	}
}

func TestLocalFileManager_DeleteFile(t *testing.T) {
	resetMocks()
	lfm := &LocalFileManager{}
	filepath := "test.txt"

	err := lfm.DeleteFile(filepath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !mockDeleteLocalFileCalled {
		t.Error("expected DeleteLocalFile to be called")
	}
	if mockDeleteLocalFileArg != filepath {
		t.Errorf("expected filepath %s, got %s", filepath, mockDeleteLocalFileArg)
	}
}

func TestLocalFileManager_OpenFile_Success(t *testing.T) {
	resetMocks()
	lfm := &LocalFileManager{}
	filepath := "test.txt"
	expectedContent := []byte("file data")

	// Create a temporary file with the expected content
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(expectedContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}
	mockOpenLocalFileRet = tmpfile
	mockOpenLocalFileErr = nil

	file, err := lfm.OpenFile(filepath)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	defer file.Close()
	if !mockOpenLocalFileCalled {
		t.Error("expected OpenLocalFile to be called")
	}
	if mockOpenLocalFileArg != filepath {
		t.Errorf("expected filepath %s, got %s", filepath, mockOpenLocalFileArg)
	}
	buf := make([]byte, len(expectedContent))
	n, _ := file.Read(buf)
	if !bytes.Equal(buf[:n], expectedContent) {
		t.Errorf("expected file content %s, got %s", expectedContent, buf[:n])
	}
}

func TestLocalFileManager_OpenFile_Error(t *testing.T) {
	resetMocks()
	lfm := &LocalFileManager{}
	filepath := "test.txt"
	mockOpenLocalFileRet = nil
	mockOpenLocalFileErr = errors.New("open error")

	file, err := lfm.OpenFile(filepath)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err == nil {
		t.Error("expected error, got nil")
	}
	if file != nil {
		t.Error("expected file to be nil on error")
	}
}

func TestGetFileManager_Local(t *testing.T) {
	fm := GetFileManager("local")
	if fm == nil {
		t.Error("expected non-nil FileManagerInterface for 'local'")
	}
	_, ok := fm.(*LocalFileManager)
	if !ok {
		t.Error("expected type *LocalFileManager")
	}
}

func TestGetFileManager_Default(t *testing.T) {
	fm := GetFileManager("unknown")
	if fm == nil {
		t.Error("expected non-nil FileManagerInterface for unknown type")
	}
	_, ok := fm.(*LocalFileManager)
	if !ok {
		t.Error("expected type *LocalFileManager")
	}
}
