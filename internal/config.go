package internal

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type Config interface {
	Unmarshal(d []byte, v interface{})
	Marshal()
}

type UserConfig struct {
	Email string `property:"email"`
	Name  string `property:"name"`
	Owner bool   `property:"owner"`
}

type CoreConfig struct {
	Bare     bool `property:"bare"`
	Filemode bool `property:"filemode"`
}

type GotConfig struct {
	User     UserConfig `property:"user"`
	Bare     bool       `property:"bare"`
	Branch   string     `property:"branch"`
	Core     CoreConfig `property:"core"`
	MaxCache int        `property:"max_cache"`
}



// Encode the g interface into buffer of string java properties-ish.
// user.email=some@email.com
// branch=false
// maxBlobSize=13000
// core.branch.default=main
//
// The encode allows to encode struct member of types string, bool, int and other struct.
//
// The encode does not support slices and arrays types. It can be extended in the future with commas sepator.
// TODO: separator might be passed by user.
func Marshal(g interface{}, ret *bytes.Buffer) error {
	return marshal(g, "", ret)
}
func marshal(g interface{}, parent string, ret *bytes.Buffer) error {
	t := reflect.TypeOf(g)
	v := reflect.ValueOf(g)
	for index := range t.NumField() {
		key := formatKey(parent, t, index)
		switch v.Field(index).Kind() {
		case reflect.String:
			ret.Write([]byte(key))
			ret.Write([]byte{'='})
			ret.Write([]byte(v.Field(index).String()))
			ret.Write([]byte{'\n'})
		case reflect.Bool:
			ret.Write([]byte(key))
			ret.Write([]byte{'='})
			ret.Write([]byte(strconv.FormatBool(v.Field(index).Bool())))
			ret.Write([]byte{'\n'})
		case reflect.Int:
			ret.Write([]byte(key))
			ret.Write([]byte{'='})
			ret.Write([]byte(strconv.FormatInt(v.Field(index).Int(), 10)))
			ret.Write([]byte{'\n'})
		case reflect.Struct:
			err := marshal(v.Field(index).Interface(), key, ret)
			if err != nil {
				return err
			}
		case reflect.Array, reflect.Slice:
			return errors.New("slice and array are not support")
		default:
			return errors.New("unexpected type")
		}
	}
	return nil
}

// Decode bytes into interface object.
func Unmarshal(d []byte, g interface{}) error {
	return unmarshal(parse(d), reflect.ValueOf(g).Elem().Type(), reflect.ValueOf(g).Elem(), "")
}
func unmarshal(m map[string]string, t reflect.Type, v reflect.Value, parent string) error {
	for index := range t.NumField() {
		fieldType := t.Field(index)
		fieldValue := v.Field(index)
		key := formatKey(parent, t, index)
		switch fieldType.Type.Kind() {

		case reflect.String:
			if value, ok := m[key]; ok {
				if !fieldValue.CanSet() {
					return ErrorParsingObject
				}
				fieldValue.Set(reflect.ValueOf(value))

			}
		case reflect.Bool:
			if value, ok := m[key]; ok {
				if !fieldValue.CanSet() {
					panic(ErrorParsingObject)
				}
				val, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				fieldValue.Set(reflect.ValueOf(val))
			}
		case reflect.Int:
			if value, ok := m[key]; ok {
				if !fieldValue.CanSet() {
					return ErrorParsingObject
				}
				val, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return err
				}
				fieldValue.Set(reflect.ValueOf(int(val)))
			}
		case reflect.Struct:

			unmarshal(m, fieldValue.Type(), fieldValue, key)
		default:
			return errors.New("unimplemented")
		}
	}
	return nil
}

// Format the key of the property.
func formatKey(parent string, t reflect.Type, index int) string {
	var key string
	if len(parent) > 0 {
		key = parent + "." + t.Field(index).Tag.Get("property")
	} else {
		key = t.Field(index).Tag.Get("property")
	}
	return key
}


func parse(d []byte) map[string]string {
	m := make(map[string]string)
	lines := strings.Split(string(d), string(newLine))
	for _, line := range lines {
		elements := strings.Split(line, string([]byte{'='}))
		if len(elements) >= 2 {
			key := elements[0]
			value := elements[1]
			m[key] = value
			continue
		}
	}
	return m
}
