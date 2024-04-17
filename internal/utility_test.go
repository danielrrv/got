package internal_test

import (
	// "io/fs"
	"os"
	"path/filepath"

	// internal "github.com/danielrrv/got/internal"
)

const (
	testFolder = "test-repo-location"
)
var (
	repoPath string
)

type TestingFile struct{
	Name string
	Data []byte
	RelativePath string
}




func CreateFilesTesting(projectTemporalFolder string, folders []string, files []TestingFile) {
	// fileA := filepath.Join(projectTemporalFolder, "src", "readme.md")
	// fileB := filepath.Join(projectTemporalFolder, "src", "cache.rs")
	// fileC := filepath.Join(projectTemporalFolder, "src", "base64.c")
	// files := []string{fileA, fileB, fileC}
	for _, dir := range folders {
		err := os.MkdirAll(filepath.Join(projectTemporalFolder, dir), 0744)
		if err != nil {

		}
	}

	for _, file := range files {
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
