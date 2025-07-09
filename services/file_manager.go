package services

import (
	"io"

	"example.com/travel-advisor/utils"
)

type FileManagerInterface interface {
	SaveContentInFile(filepath string, content *string) error
	DeleteFile(filepath string) error
	OpenFile(filepath string) (io.ReadSeekCloser, error)
}

type LocalFileManager struct {
}

func (lfm *LocalFileManager) SaveContentInFile(filepath string, content *string) error {
	err := utils.WriteLocalFile(filepath, []byte(*content), 0644)
	return err
}

func (lfm *LocalFileManager) DeleteFile(filepath string) error {
	err := utils.DeleteLocalFile(filepath)
	return err
}

func (lfm *LocalFileManager) OpenFile(filepath string) (io.ReadSeekCloser, error) {
	file, err := utils.OpenLocalFile(filepath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

var GetFileManager = func(fileManagerType string) FileManagerInterface {
	//TODO AWS S3 support for file management
	if fileManagerType == "s3" {

	} else if fileManagerType == "local" {
		return &LocalFileManager{}
	} else {
		return &LocalFileManager{}
	}

	return nil
}
