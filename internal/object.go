package internal

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	// "fmt"

	// "fmt"
	"io"

	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type GotObject interface {
	Commit | Tree
}


const (
	tagName            = "object"
	newLine           = '\n'
	tab                = '\t'
	commitHeaderName = string("commit")
	treeHeaderName   = string("tree")
	blobHeaderName   = string("blob")
)

var (
	ErrorParsingObject       = errors.New("error parsing object")
	ErrorIsNotObject         = errors.New("the pointer isn't an object")
	ErrorIncorrectOBjectType = errors.New("incorrect object type")
)

func Serialize[K GotObject](c *K) ([]byte, error) {
	var out bytes.Buffer
	t := reflect.TypeOf(*c)
	v := reflect.ValueOf(*c)

	if t.Kind() != reflect.Struct {
		return nil, ErrorIsNotObject
	}
	for index := range t.NumField() {
		out.Write([]byte(t.Field(index).Tag.Get(tagName)))
		out.WriteByte(tab)
		out.Write([]byte(v.Field(index).String()))
		if t.NumField()-1 > index {
			out.WriteByte(newLine)
		}
	}
	return out.Bytes(), nil
}

func Deserialize[K GotObject](c *K, b []byte) error {
	t := reflect.TypeOf(*c)
	v := reflect.ValueOf(c).Elem()
	m := make(map[string]interface{})

	lines := strings.Split(string(b), string(newLine))
	for _, line := range lines {
		elements := strings.Split(line, string(tab))
		if len(elements) < 2 {
			return ErrorParsingObject
		}
		m[elements[0]] = elements[1]
	}
	for i := range t.NumField() {
		// fieldTagName := t.Field(i).Tag.Get(tagName)
		field := v.FieldByName(t.Field(i).Name)
		if !field.CanSet() {
			return ErrorParsingObject
		}
		field.Set(reflect.ValueOf(m[t.Field(i).Tag.Get(tagName)]))
	}
	return nil
}

func Compress(b []byte, c *bytes.Buffer) {
	w := zlib.NewWriter(c)
	w.Write(b)
	w.Close()
}

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

func ReadObject(repo *GotRepository, header string, hash string) ([]byte, error) {
	sizePos := len(header) + 1
	dataStartPos := len(header) + 3
	content, err := os.ReadFile(filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
	if err != nil {
		return nil, err
	}
	var bb bytes.Buffer
	Decompress(content, &bb)
	if bytes.Compare(bb.Bytes()[0:len(header)], []byte(header)) > 0 {
		return nil, ErrorIncorrectOBjectType
	}
	if bb.Bytes()[len(header)] != 0x20 {
		return nil, errors.New("malformed object")
	}
	data := bb.Bytes()[dataStartPos : int(bb.Bytes()[sizePos])+dataStartPos]
	return data, nil

}

func RemoveObjectFrom(repo *GotRepository, hash string) error {
	return os.Remove(filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
}

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

func BuildObject(header string, data []byte) []byte {
	buf := make([]byte, 0)
	//Build object.
	buf = append(buf, []byte(header)...)
	buf = append(buf, 0x20)
	buf = append(buf, byte(len(data)))
	buf = append(buf, 0x00)
	buf = append(buf, data...)
	return buf
}

func HashToPath(repo * GotRepository,hash string)(string, error){
	if len(hash) != sha1.Size * 2{
		return "", errors.New("inconsistent object id")
	} 
	return filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]), nil
}

func CreatePossibleObjectFromData(repo *GotRepository, objD []byte, header string) (string, error){
	//1. Build the object
	rawObj := BuildObject(header, objD)
	//2. Derive the has
	hash := CreateSha1(rawObj)
	return string(hash), nil
}

func WriteObject(repo *GotRepository, objD []byte, header string) (string, error) {

	//1. Build the object
	rawObj := BuildObject(header, objD)
	//2. Derive the has
	hash := CreateSha1(rawObj)
	//3. Compress
	var bb bytes.Buffer
	Compress(rawObj, &bb)
	//4. Write on disk
	TryCreateFolderIn(filepath.Join(repo.GotDir, gotRepositoryDirObjects), string(hash[:2]))
	objPath, err := HashToPath(repo, string(hash))
	if err != nil{
		return "", err
	}
	file, err := os.OpenFile(objPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.Write(bb.Bytes())
	return string(hash), nil
}
