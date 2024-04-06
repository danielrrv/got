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
	index  *Index
}

func ReadRepoFile(repo *GotRepository, fileDir string) ([]byte, error) {
	content, err := os.ReadFile(filepath.Join(repo.GotDir, fileDir))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func CreateRepoFile(repo *GotRepository, filename string, data []byte) ([]byte, error) {
	path := filepath.Join(repo.GotDir, gotRootRepositoryDir, filename)
	if file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644); err == nil {
		defer file.Close()
		if _, err := file.Write(data); err != nil {
			return nil, err
		}
	}
	return nil, errors.New("unable to find the file in the repo")
}

func newGotRepository(rootP string) (*GotRepository, error) {
	if fi, err := os.Stat(rootP); errors.Is(err, os.ErrNotExist) || !fi.IsDir() {
		return nil, fmt.Errorf("path isn't either a dir or doesn't exists")
	}
	return &GotRepository{
		GotTree:   rootP,
		GotDir:    filepath.Join(rootP, gotRootRepositoryDir),
		GotConfig: GotConfig{},
	}, nil
}

func (gr *GotRepository) SetConfig(config interface{}) {
	gr.GotConfig = config.(GotConfig)
}

func TryCreateFileIn(path, filename string) error {
	if filepath.IsAbs(path) && !pathExist(path, true) {
		os.Mkdir(path, fs.ModeDir|0750)
	}
	filePath := filepath.Join(path, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	file.Chmod(0750)
	return nil
}

func TryCreateFolderIn(path, dir string) {
	dirPath := filepath.Join(path, dir)
	if !pathExist(dirPath, true) {
		os.Mkdir(dirPath, fs.ModeDir|0750)
	}
}

func FindOrCreateRepo(path string) (*GotRepository, error) {
	gotRootPath := filepath.Join(path, gotRootRepositoryDir)
	versionPathFile := filepath.Join(path, gotRootRepositoryDir, "version")
	if !pathExist(path, true) {
		return nil, ErrorPathDoesNotExist
	}
	if pathExist(gotRootPath, true) && pathExist(versionPathFile, false) {
		content, err := os.ReadFile(versionPathFile)
		if err != nil {
			panic(err)
		}
		if VersionRegex.Match(content) {
			return &GotRepository{
				GotTree:   path,
				GotDir:    filepath.Join(path, gotRootRepositoryDir),
				GotConfig: GotConfig{},
			}, nil
		}
	}
	repo, err := newGotRepository(path)
	if err != nil {
		panic(err)
	}
	TryCreateFolderIn(path, gotRootRepositoryDir)

	if versionFile, err := os.OpenFile(filepath.Join(gotRootPath, "version"), os.O_RDWR|os.O_CREATE, 0644); err == nil {
		defer versionFile.Close()
		if _, err := versionFile.Write([]byte(fmt.Sprintf("version: %s", version))); err != nil {
			panic(err)
		}
	}
	TryCreateFolderIn(gotRootPath, gotRepositoryDirRefs)
	TryCreateFolderIn(gotRootPath, gotRepositoryDirObjects)
	TryCreateFolderIn(gotRootPath, filepath.Join(gotRepositoryDirRefs, gotRepositoryDirRefsHeads))
	TryCreateFolderIn(gotRootPath, filepath.Join(gotRepositoryDirRefs, gotRepositoryDirRefsTags))
	return repo, nil
}

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
