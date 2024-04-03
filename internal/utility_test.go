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

func init() {
	_rootDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	repoPath = filepath.Join(_rootDir, testFolder)
}


