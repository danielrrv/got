package internal

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)



const (
	newLine          = '\n'
	tab              = '\t'
	CommitHeaderName = string("commit")
	TreeHeaderName   = string("tree")
	BlobHeaderName   = string("blob")
)

var (
	ErrorParsingObject       = errors.New("error parsing object")
	ErrorIsNotObject         = errors.New("the pointer isn't an object")
	ErrorIncorrectOBjectType = errors.New("incorrect object type")
	ErrorMalformedObject     = errors.New("malformed object")
)


type GotObject interface {
	// Implementation to transform struct to bytes.
	Serialize() []byte
	// // Implementation to get the object location on got/objects folders.
	// Location() string
}

type Object struct {
	// Type of object. blob, tree, commit or tag.
	Header []byte
	// Seperator of the header.
	Pad1   byte //0x20
	// Size of the Data
	Size   uint32
	// Separator of the Data
	Pad2   byte //0x00
	// Actual Object data.
	Data   []byte
}

// Zlib compress data.
func Compress(b []byte, c *bytes.Buffer) {
	w := zlib.NewWriter(c)
	w.Write(b)
	w.Close()
}

// Zlib uncompress data.
func Decompress(b []byte, c *bytes.Buffer) {

	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	bb, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	c.Write(bb)
	r.Close()
}

// Read any got object given the hash/object id and header(commit, tree, tags, blob)
func ReadObject(repo *GotRepository, header string, hash string) ([]byte, error) {
	//decompress(header[unbound size uint8]|0x20[uint8 x 1]|size[uint32 x 1]|0x00[uint8 x 1]|data[unbound size uint8])
	sizePos := len(header) + 1
	// dataStartPos := len(header) + 3	
	content, err := os.ReadFile(filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
	if err != nil {
		return nil, err
	}
	// Decompress the object.
	var bb bytes.Buffer
	Decompress(content, &bb)
	if !bytes.Equal(bb.Bytes()[0:len(header)], []byte(header)){
		return nil, ErrorIncorrectOBjectType
	}
	//Implementation to validate the correctness at this point.
	if bb.Bytes()[len(header)] != 0x20 {
		return nil, ErrorMalformedObject
	}
	//Size of data is uint32
	sizeOfData := Bit32FromBytes(bb.Bytes()[sizePos:sizePos + 4])
	// after :sizePos + 4, data comes.
	data := bb.Bytes()[sizePos+5: sizePos + 5 + int(sizeOfData)]
	return data, nil

}

// Remove object given the objectId.
func RemoveObjectFrom(repo *GotRepository, hash string) error {
	return os.Remove(filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
}

// Create sha1 hash from data. TODO: Open to other hasher.
func CreateSha1(data []byte) []byte {
	hash := make([]byte, sha1.Size*2)
	hasher := sha1.New()
	_, err := hasher.Write(data)
	if err != nil {
		panic(err)
	}
	hex.Encode(hash, hasher.Sum(nil))
	return hash
}

// Base method to abstract serialization of any GotObject.
func newObject(header string, g GotObject) *Object {
	data := g.Serialize()
	return &Object{
		Header: []byte(header),
		Pad1:   0x20,
		Size:   uint32(len(data)),
		Pad2:   0x00,
		Data:   data,
	}
}

// Build object from data.
func BuildObject(header string, g GotObject) []byte {
	obj := newObject(header, g)
	packet :=AllocatePacket(0)
	// header[unbound size uint8]|0x20[uint8 x 1]|size[uint32 x 1]|0x00[uint8 x 1]|data[unbound size uint8]
	packet.Set(obj.Header, []byte{obj.Pad1}, Bit32(obj.Size).Bytes(), []byte{obj.Pad2}, obj.Data) 
	return packet.buff
}

// obtain the object's path given the hash.
func HashToPath(repo *GotRepository, hash string) (string, error) {
	if len(hash) != sha1.Size*2 {
		return "", errors.New("inconsistent object id")
	}
	return filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]), nil
}

// Create in-memory the object with its hash given the data and object type.
func CreatePossibleObjectFromData(repo *GotRepository, g GotObject, header string) (string, error) {
	//1. Build the object
	rawObj := BuildObject(header, g)
	//2. Derive the has
	hash := CreateSha1(rawObj)
	return string(hash), nil
}

// [Persist] the object in disk given the data. CratePossibleObject must have generated the same hash. Use cautionsly.
func WriteObject(repo *GotRepository, g GotObject, header string) (string, error) {
	// fmt.Println("Writing a "+ header + " at "+ g.Location())
	//1. Build the object
	rawObj := BuildObject(header, g)
	//2. Derive the has
	hash := CreateSha1(rawObj)
	//3. Compress
	var bb bytes.Buffer
	Compress(rawObj, &bb)
	//4. Object parent folder not created.
	if dir := filepath.Join(repo.GotDir, gotRepositoryDirObjects, string(hash[:2])); !pathExist(dir, true) {
		os.Mkdir(dir, fs.ModePerm|0644)
	}
	//Create final path from the hash.
	objPath, err := HashToPath(repo, string(hash))
	if err != nil {
		return "", err
	}
	//5. Write the data in the object. Corruption may happen.
	file, err := os.OpenFile(objPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.Write(bb.Bytes())
	return string(hash), nil
}
