package services

import "example.com/travel-advisor/utils"

type FileManagerInterface interface {
	SaveContentInFile(filepath string, content *string) error
}

type LocalFileManager struct {
}

func (lfm *LocalFileManager) SaveContentInFile(filepath string, content *string) error {
	err := utils.WriteLocalFile(filepath, []byte(*content), 0644)
	return err
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
