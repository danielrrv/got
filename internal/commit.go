package internal

import (
	"bytes"
	"reflect"
	"strings"
)

// "bytes"
// "reflect"
// "strings"

type Commit struct {
	Author      string `object:"author"`
	Committer   string `object:"committer"`
	Tree        string `object:"tree"`
	Date        string `object:"date"`
	Description string `object:"description"`
	Parent      string `object:"parent"`
}

func (c Commit) Serialize() ([]byte) {
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
			// fieldTagName := t.Field(i).Tag.Get(tagName)
			field := v.FieldByName(t.Field(i).Name)
			if !field.CanSet() {
				panic(ErrorParsingObject)
			}
			field.Set(reflect.ValueOf(m[t.Field(i).Tag.Get(tagName)]))
		}
		return v.Interface().(Commit)
}
