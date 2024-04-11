package internal

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
)

// "bytes"
// "reflect"
// "strings"

type Commit struct {
	Hash        string `object:"hash"`
	Author      string `object:"author"`
	Committer   string `object:"committer"`
	Tree        string `object:"tree"`
	Date        string `object:"date"`
	Description string `object:"description"`
	Parent      string `object:"parent"`
}

// Turn Commit instance into array of bytes.
func (c Commit) Serialize() []byte {
	var out bytes.Buffer
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)

	if t.Kind() != reflect.Struct {
		panic(ErrorIsNotObject)
	}
	for index := range t.NumField() {
		out.Write([]byte(t.Field(index).Tag.Get(tagName)))
		out.WriteByte(tab)
		out.Write([]byte(v.Field(index).String()))
		if t.NumField()-1 > index {
			out.WriteByte(newLine)
		}
	}
	return out.Bytes()
}

// Convert an array of byte to a Commit instance.
func (c Commit) Deserialize(d []byte) Commit {
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(&c).Elem()
	m := make(map[string]interface{})

	lines := strings.Split(string(d), string(newLine))
	for _, line := range lines {

		elements := strings.Split(line, string(tab))
		if len(elements) < 2 {
			panic(ErrorParsingObject)
		}
		m[elements[0]] = elements[1]
	}
	for i := range t.NumField() {
		field := v.FieldByName(t.Field(i).Name)
		if !field.CanSet() {
			panic(ErrorParsingObject)
		}
		field.Set(reflect.ValueOf(m[t.Field(i).Tag.Get(tagName)]))
	}
	return v.Interface().(Commit)
}

func (c Commit) Location() string {
	if c.Hash == "" {
		panic("hash hasn't been calculated yet.")
	}
	return filepath.Join(c.Hash[:2], c.Hash[:2])
}


// func CreateCommit() * Commit{}