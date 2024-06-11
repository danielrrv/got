package internal_test

import (
	// "io/fs"
	"fmt"
	"os"
	"path/filepath"
	// internal "github.com/danielrrv/got/internal"
)


type TestingFile struct{
	Name string
	Data []byte
	RelativePath string
}




func CreateFilesTesting(projectTemporalFolder string, folders []string, files []TestingFile) {

	for _, dir := range folders {
		err := os.MkdirAll(filepath.Join(projectTemporalFolder, dir), 0744)
		if err != nil {
			panic(err)
		}
	}

	for _, file := range files {
		fmt.Println("Creating file", filepath.Join(projectTemporalFolder, file.RelativePath))
		fd, err := os.OpenFile(filepath.Join(projectTemporalFolder, file.RelativePath), os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic(err)
		}

		if _, err := fd.Write([]byte(file.Data)); err != nil {
			panic(err)
		}
		fd.Close()
	}
}
