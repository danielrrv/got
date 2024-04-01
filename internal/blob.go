package internal

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrorAlreadyExist = errors.New("file already exist")
	InvalidHash       = errors.New("invalid sha1 hash")
	NotDataToWrite    = errors.New("insufficient data to write")
	NotBlobFound      = errors.New("no blob found in the path provided")
)

type Blob struct {
	Repo   *GotRepository
	Commit *Commit
	Path   string
	Hash   string
	Data   string
}

func (b *Blob) Write() error {
	if len(b.Hash) != sha1.Size*2 {
		return InvalidHash
	}
	if len(b.Data) == 0 {
		return NotDataToWrite
	}
	if !fs.ValidPath(b.Path) {
		return ErrroPathInvalid
	}
	// isAbs := filepath.IsAbs(b.Path)
	fi, err := os.Stat(b.Path)
	if err != nil {
		return err
	}
	if fi.Name() == b.Hash[2:] {
		return ErrorAlreadyExist
	}
	// filepath.Dir(b.Path)
	TryCreateFolderIn(filepath.Join(b.Repo.GotDir, gotRepositoryDirObjects), b.Hash[:2])
	file, err := os.OpenFile(b.Path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	var c bytes.Buffer
	Compress([]byte(b.Data), &c)
	file.Write(c.Bytes())
	return nil
}

func BlobFromPath(repo *GotRepository, path string) (*Blob, error) {
	var realP string
	isAbs := filepath.IsAbs(path)
	if isAbs {
		realP = path
	} else {
		realP = filepath.Join(repo.GotDir, gotRepositoryDirObjects, filepath.Dir(path), filepath.Base(path))
	}

	if !fs.ValidPath(realP) {
		return nil, ErrroPathInvalid
	}

	dirs, err := os.ReadDir(filepath.Join(repo.GotDir, gotRepositoryDirObjects, filepath.Dir(realP)))
	if err != nil {
		return nil, err
	}
	exist := false
	for _, dir := range dirs {
		if filepath.Base(realP) == dir.Name() && !dir.IsDir() {
			exist = true
		}
	}
	if !exist {
		return nil, NotBlobFound
	}
	content, err := os.ReadFile(realP)
	if err != nil {
		return nil, err
	}
	var c bytes.Buffer

	Decompress(content, &c)

	return &Blob{
		Repo:   repo,
		Hash:   fmt.Sprintf("%s%s", filepath.Dir(realP), filepath.Base(realP)),
		Data:   string(c.Bytes()),
		Path:   realP,
		Commit: nil,
	}, nil
}

func BlobFromHash(repo *GotRepository, hash string)  (*Blob, error) {
	if len(hash) != sha1.Size*2 {
		return nil,InvalidHash
	}
	return BlobFromPath(repo,filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
}
