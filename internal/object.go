package internal

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
)

type GotObject interface {
	Commit
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
	ErrorParsingObject = errors.New("Error parsing object")
	ErrorIsNotObject   = errors.New("The pointer isn't an object")
)

func Serialize[K GotObject](c *K) ([]byte, error) {
	var out bytes.Buffer
	t := reflect.TypeOf(*c)
	v := reflect.ValueOf(*c)
	if t.Kind() != reflect.Struct {
		return nil, ErrorIsNotObject
	}
	for index := range t.NumField() {
		field := t.Field(index)
		out.Write([]byte(field.Tag.Get(tagName)))
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
		fieldTagName := t.Field(i).Tag.Get(tagName)
		field := v.FieldByName(t.Field(i).Name)
		if !field.CanSet() {
			return ErrorParsingObject
		}
		field.Set(reflect.ValueOf(m[fieldTagName]))
	}
	return nil
}

func WriteObject[K GotObject](c *K) {
	t := reflect.TypeOf(*c)
	// fmt.Println(t)
	switch t.Name() {
	case "Commit":

	}

}
