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
	ErrorInvalidHash       = errors.New("invalid sha1 hash")
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

func (b *Blob) IntermediateBlobObject() (hash string, err error){
	hash, err = CreatePossibleObjectFromData(b.Repo,[]byte(b.Data), blobHeaderName)
	return hash, err
}

func (b *Blob) Persist() error {
	if len(b.Data) == 0 {
		return NotDataToWrite
	}
	hash, err := WriteObject(b.Repo, []byte(b.Data), blobHeaderName)
	if err != nil {
		return err
	}
	path, err := HashToPath(b.Repo, hash)
	if err != nil {
		return err
	}
	b.Hash = hash
	b.Path = path
	return nil
}


func ReadBlob(repo *GotRepository, objId string) (*Blob, error) {
	path, err := HashToPath(repo, objId)
	if err != nil {
		return nil, err
	}
	data, err := ReadObject(repo, blobHeaderName, objId)
	if err != nil {
		return nil, err
	}

	return &Blob{
		Repo:   repo,
		Hash:   objId,
		Data:   string(data),
		Path:   path,
		Commit: nil,
	}, nil

}


//Deprecated
func BlobFromPath(repo *GotRepository, path string) (*Blob, error) {
	var realP string
	isAbs := filepath.IsAbs(path)
	if isAbs {
		realP = path
	} else {
		realP = filepath.Join(repo.GotDir, gotRepositoryDirObjects, filepath.Dir(path), filepath.Base(path))
	}

	if !fs.ValidPath(realP) {
		return nil, ErrorPathInvalid
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
		Data:  	c.String(),
		Path:   realP,
		Commit: nil,
	}, nil
}

func BlobFromHash(repo *GotRepository, hash string) (*Blob, error) {
	if len(hash) != sha1.Size*2 {
		return nil, ErrorInvalidHash
	}
	return BlobFromPath(repo, filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
}
