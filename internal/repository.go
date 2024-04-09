package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

const (
	gotRepositoryDirRefs      = "refs"
	gotRepositoryDirRefsTags  = "tags"
	gotRepositoryDirRefsHeads = "heads"
	gotRepositoryDirObjects   = "objects"
	gotRootRepositoryDir      = ".got"
	version                   = "v1.0.0"
)

var (
	ErrorPathInvalid            = errors.New("path is invalid")
	ErrorPathDoesNotExist       = errors.New("path does not exist")
	ErrorRepositoryDoesNotExist = errors.New("repository does not exist")
	ErrorLoadConfig             = errors.New("unable to load the configuration")
	ErrorOpeningFile            = errors.New("unable to find the file in the repo")
)
var (
	VersionRegex = regexp.MustCompile("^version")
)

type GotConfig struct{}

// https://github.com/atom/git-utils/blob/master/src/repository.h
type GotRepository struct {
	// The root folder of the repository.
	GotTree string
	// The repository configurations.
	GotConfig GotConfig
	// .got folder name.
	GotDir string
	// Temporary database.
	index *Index
}

// Create a file inside of the repo dir(.got)
func CreateRepoFile(repo *GotRepository, filename string, data []byte) ([]byte, error) {
	path := filepath.Join(repo.GotDir, filename)
	//TODO: Permission rw for owner, read for rest.
	//if file exist, the data will be appended.
	if file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644); err == nil {
		defer file.Close()
		if _, err := file.Write(data); err != nil {
			return nil, err
		}
	}
	return nil, ErrorOpeningFile
}

// Set the repo configuration after setup. Future usage.
func (gr *GotRepository) SetConfig(config interface{}) {
	gr.GotConfig = config.(GotConfig)
}

func FindRecursivelyFolder(path string, folder string, until int) (string, error) {
	if until == 0 {
		return "", errors.New("recursive search exhausted")
	}
	if pathExist(filepath.Join(path, folder), true) {
		return path, nil
	}
	return FindRecursivelyFolder(filepath.Dir(path), folder, until-1)
}

// Find the repo if exist in the path or create new one.
func FindOrCreateRepo(path string) (*GotRepository, error) {
	if !pathExist(path, true) {
		return nil, ErrorPathDoesNotExist
	}
	// Default folder location in case not .got folder found.
	gotDir := filepath.Join(path, gotRootRepositoryDir)
	treeDir := path
	dirPath, err := FindRecursivelyFolder(path, gotRootRepositoryDir, 4)
	// No error found, then dirPath is the rootDir where .got lives.
	if err == nil{
		treeDir = dirPath
		gotDir = filepath.Join(dirPath, gotRootRepositoryDir)
	} 
	// Possible repo representation. No folder created yet.
	repo := &GotRepository{
		GotTree: treeDir,
		GotDir:   gotDir,
		GotConfig: GotConfig{},
	}
	//The folder path/.got exist and it has a version file that will determine this is already created repo.
	if pathExist(filepath.Join(repo.GotDir, "version"), false) {
		content, err := os.ReadFile(filepath.Join(repo.GotDir, "version"))
		if err != nil {
			panic(err)
		}
		// The regex acts a validation mechanism to determine the repo existance. Can be changed.
		if VersionRegex.Match(content) {
			return repo, nil
		}
	}
	//The path/.got doesn't exist. Let's create it.
	if dir := gotDir; !pathExist(dir, true) {
		os.Mkdir(dir, fs.ModePerm|0755)
	}
	//Create the version's file and write in it.
	if _, err := CreateRepoFile(repo, "version", []byte(fmt.Sprintf("version: %s", version))); err != nil {
		panic(err)
	}
	//Create path/refs
	if dir := filepath.Join(gotDir, gotRepositoryDirRefs); !pathExist(dir, true) {
		os.Mkdir(dir, fs.ModePerm|0755)
	}
	//Create path/objects
	if dir := filepath.Join(gotDir, gotRepositoryDirObjects); !pathExist(dir, true) {
		os.Mkdir(dir, fs.ModePerm|0755)
	}
	//Create path/refs/heads
	if dir := filepath.Join(gotDir, gotRepositoryDirRefs, gotRepositoryDirRefsHeads); !pathExist(dir, true) {
		os.Mkdir(dir, fs.ModePerm|0755)
	}
	//Create path/refs/tags
	if dir := filepath.Join(gotDir, gotRepositoryDirRefs, gotRepositoryDirRefsTags); !pathExist(dir, true) {
		os.Mkdir(dir, fs.ModePerm|0755)
	}
	return repo, nil
}

// Util function to determine whether the file/dir exists.
func pathExist(path string, mustBeDir bool) bool {
	fi, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		panic(err)
	}
	if mustBeDir {
		return !errors.Is(err, os.ErrNotExist) && fi.IsDir()
	}
	return true
}
