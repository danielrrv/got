package internal

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"

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
	tagName  = "object"
	NEW_LINE = '\n'
	TAB      = '\t'
	COMMIT   = string("commit")
	TREE     = string("tree")
	BLOB     = string("blob")
)

var (
	ErrorParsingObject = errors.New("error parsing object")
	ErrorIsNotObject   = errors.New("the pointer isn't an object")
)

func Serialize[K GotObject](c *K) ([]byte, error) {
	var out bytes.Buffer;t := reflect.TypeOf(*c);v := reflect.ValueOf(*c)
	
	if t.Kind() != reflect.Struct {
		return nil, ErrorIsNotObject
	}
	for index := range t.NumField() {
		out.Write([]byte( t.Field(index).Tag.Get(tagName)))
		out.WriteByte(TAB)
		out.Write([]byte(v.Field(index).String()))
		if t.NumField()-1 > index {
			out.WriteByte(NEW_LINE)
		}
	}
	return out.Bytes(), nil
}

func Deserialize[K GotObject](c *K, b []byte) error {

	t := reflect.TypeOf(*c)
	v := reflect.ValueOf(c).Elem()
	m := make(map[string]interface{})

	lines := strings.Split(string(b), string(NEW_LINE))
	for _, line := range lines {
		elements := strings.Split(line, string(TAB))
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

func Decompress(b []byte, c *bytes.Buffer) error {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	bb, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	c.Write(bb)
	r.Close()
	return nil
}

func ReadObject[K GotObject](repo *GotRepository, ob *K, header string, hash string) error {
	sizePos := len(header) + 1
	dataStartPos := len(header) + 3
	content, err := os.ReadFile(filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
	if err != nil {
		return err
	}
	var bb bytes.Buffer
	Decompress(content, &bb)
	if bytes.Compare(bb.Bytes()[0:len(header)], []byte(header)) > 0 {
		return errors.New("incorrect object type")
	}
	if bb.Bytes()[len(header)] != 0x20 {
		return errors.New("malformed object")
	}
	data := bb.Bytes()[dataStartPos : int(bb.Bytes()[sizePos])+dataStartPos]
	fmt.Println(string(data))
	return nil

}

func RemoveObjectFrom(repo *GotRepository, hash string) error {
	return os.Remove(filepath.Join(repo.GotDir, gotRepositoryDirObjects, hash[:2], hash[2:]))
}

func WriteObject[K GotObject](repo *GotRepository, ob *K, header string) (string, error) {
	
	var bb bytes.Buffer
	hash := make([]byte, sha1.Size*2)
	buf := make([]byte, 0)

	//Serilize the object to bytes.
	b, err := Serialize(ob)
	if err != nil {
		return "", err
	}
	//Hash the name and hex data.
	hasher := sha1.New()
	_, err = hasher.Write(b)
	if err != nil {
		return "", err
	}
	hex.Encode(hash, hasher.Sum(nil))

	//Build object.
	buf = append(buf, []byte(header)...)
	buf = append(buf, 0x20)
	buf = append(buf, byte(len(b)))
	buf = append(buf, 0x00)
	buf = append(buf, b...)

	//Compress
	Compress(buf, &bb)
	//Populate the file.
	TryCreateFolderIn(filepath.Join(repo.GotDir, gotRepositoryDirObjects), string(hash[:2]))
	file, err := os.OpenFile(filepath.Join(repo.GotDir, gotRepositoryDirObjects, string(hash[:2]), string(hash[2:])), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	file.Write(bb.Bytes())
	return string(hash), nil
}
