package internal

import (
	"errors"

	// "io/fs"
	"os"
	"path/filepath"
)

var (
	//File already exist.
	ErrorAlreadyExist = errors.New("file already exist")
	// Invalid sha1 size.
	ErrorInvalidHash = errors.New("invalid sha1 hash")
	//No data provided.
	ErrorNotDataToWrite = errors.New("insufficient data to write")
	// No file found.
	ErrorNotBlobFound = errors.New("no blob found in the path provided")
)

type Blob struct {
	// The repo representation. It can be null in some runtine moments.
	Repo *GotRepository
	// The commit associated with the blob obj. It can be null in some runtime moments.
	Commit *Commit
	// Tree location on the user location. Similar to OFS.
	Path string
	// The hash associated. The hash <-> object-path can be derived one from the other.
	Hash string
	// Decompress blob data.
	FileContent []byte
}

// Reads the blob raw data from the path/
func (b Blob) Serialize() []byte {
	content, err := os.ReadFile(b.Path)
	if err != nil {
		panic(err)
	}
	return content
}

// The deserialization of the blob is its content.
func (b Blob) Deserialize(d []byte) Blob {
	b.FileContent = d
	return b
}

// Blob path in the .got/objects folder.
func (b Blob) Location() string {
	if b.Hash == "" {
		panic("hash no generated yet")
	}
	return filepath.Join(b.Hash[:2], b.Hash[2:])
}

// Persist on this the blob.
func (b *Blob) Persist() error {
	if len(b.FileContent) == 0 {
		return ErrorNotDataToWrite
	}
	hash, err := WriteObject(b.Repo, *b, BlobHeaderName)
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

// Read the blob from got object folders.
func ReadBlobObject(repo *GotRepository, objId string) (*Blob, error) {
	path, err := HashToPath(repo, objId)
	if err != nil {
		return nil, err
	}
	data, err := ReadObject(repo, BlobHeaderName, objId)
	if err != nil {
		return nil, err
	}

	return &Blob{
		Repo:        repo,
		Hash:        objId,
		FileContent: data,
		Path:        path,
		Commit:      nil,
	}, nil
}

// Create a new instance of blob from the given user path.
func BlobFromUserPath(repo *GotRepository, path string) (*Blob, error) {
	var realP string
	isAbs := filepath.IsAbs(path)
	// Implementation to use the absolute path as long as it is. Otherwise use the constructed.
	if isAbs {
		realP = path
	} else {
		realP = filepath.Join(repo.GotTree, path)
	}
	// User blob content.
	content, err := os.ReadFile(realP)
	if err != nil {
		return nil, err
	}
	//Create base blob object. At least the content must be filled out.
	blob := Blob{
		Repo:        repo,
		Hash:        "",
		FileContent: content,
		Path:        realP,
		Commit:      nil,
	}
	// Create  possible hash build the base object.
	possibleHash, err := CreatePossibleObjectFromData(repo, blob, BlobHeaderName)
	if err != nil {
		panic(err)
	}
	blob.Hash = possibleHash
	return &blob, nil
}
