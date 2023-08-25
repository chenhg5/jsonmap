package jsonmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/modern-go/reflect2"
	"github.com/tidwall/gjson"
)

const (
	jsonMapTagKey = "jsonmap"
)

type JSONMap struct {
	Val   interface{}
	cache map[uintptr]map[int]map[string]string
}

func Wrap(val interface{}) *JSONMap {
	return &JSONMap{Val: val, cache: make(map[uintptr]map[int]map[string]string)}
}

func (mapper *JSONMap) MarshalJSON() ([]byte, error) {
	destVal := reflect.ValueOf(mapper.Val)
	typeKey := reflect2.RTypeOf(mapper.Val)
	if destVal.Kind() == reflect.Slice {
		var buf bytes.Buffer
		err := buf.WriteByte('[')
		if err != nil {
			return nil, err
		}
		for i := 0; i < destVal.Len(); i++ {
			indexByte, err := mapper.marshalStruct(typeKey, destVal.Index(i).Elem())
			if err != nil {
				return nil, err
			}
			if i == destVal.Len()-1 {
				_, err = buf.Write(indexByte)
			} else {
				_, err = buf.Write(append(indexByte, byte(',')))
			}
			if err != nil {
				return nil, err
			}
		}
		err = buf.WriteByte(']')
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	if destVal.Kind() == reflect.Ptr && destVal.Elem().Kind() == reflect.Struct {
		return mapper.marshalStruct(typeKey, destVal.Elem())
	}
	if destVal.Kind() == reflect.Struct {
		return mapper.marshalStruct(typeKey, destVal)
	}
	return nil, fmt.Errorf("not struct pointer or slice")
}

func (mapper *JSONMap) UnmarshalJSON(data []byte) error {
	destVal := reflect.ValueOf(mapper.Val).Elem()
	typeKey := reflect2.RTypeOf(mapper.Val)

	if destVal.Kind() == reflect.Slice {
		sliceType := destVal.Type().Elem()
		arr := gjson.ParseBytes(data).Array()
		for i := 0; i < len(arr); i++ {
			tmpValPtr := reflect.New(sliceType.Elem())
			err := mapper.unmarshalStruct(typeKey, arr[i], tmpValPtr.Elem())
			if err != nil {
				return err
			}
			destVal.Set(reflect.Append(destVal, tmpValPtr))
		}
	} else if destVal.Kind() == reflect.Struct {
		return mapper.unmarshalStruct(typeKey, gjson.ParseBytes(data), destVal.Addr().Elem())
	}

	return nil
}

func convertValueToFieldType(value interface{}, fieldType reflect.Type) (reflect.Value, error) {
	floatValue, ok := value.(float64)
	if !ok {
		return reflect.ValueOf(value), nil
	}
	return reflect.ValueOf(floatValue).Convert(fieldType), nil
}

func (mapper *JSONMap) unmarshalStruct(typeKey uintptr, result gjson.Result, destVal reflect.Value) error {
	destType := destVal.Type()
	for j := 0; j < destType.NumField(); j++ {
		field := destType.Field(j)
		jsonFieldName := field.Tag.Get("json")
		jsonFieldData := result.Get(jsonFieldName).Raw
		if jsonFieldData == "" {
			continue
		}
		fieldValue := destVal.Field(j)
		if _, ok := mapper.cache[typeKey]; !ok {
			mapper.cache[typeKey] = make(map[int]map[string]string)
		}
		jsonmap, ok := mapper.cache[typeKey][j]
		if ok {
			jsonFieldData, ok = jsonmap[result.Get(jsonFieldName).String()]
			if !ok {
				return fmt.Errorf("not allowed value of field %s", jsonFieldName)
			}
		} else {
			jsonMapTag, exist := field.Tag.Lookup(jsonMapTagKey)
			if exist {
				if jsonMapTag == "" {
					tmpMapper := Wrap(fieldValue.Addr().Interface())
					err := json.Unmarshal([]byte(jsonFieldData), tmpMapper)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.ValueOf(tmpMapper.Val).Elem())
					continue
				} else {
					pairs := strings.Split(jsonMapTag, ";")
					jsonmap = make(map[string]string)
					for _, pair := range pairs {
						kv := strings.Split(pair, ":")
						jsonmap[kv[1]] = kv[0]
					}
					jsonFieldData, ok = jsonmap[result.Get(jsonFieldName).String()]
					if !ok {
						return fmt.Errorf("not allowed value of field %s", jsonFieldName)
					}
					mapper.cache[typeKey][j] = jsonmap
				}
			}
		}
		err := json.Unmarshal([]byte(jsonFieldData), fieldValue.Addr().Interface())
		if err != nil {
			return err
		}
		convertedValue, err := convertValueToFieldType(fieldValue.Interface(), destVal.Field(j).Type())
		if err != nil {
			return err
		}
		fieldValue.Set(convertedValue)
	}
	return nil
}

func (mapper *JSONMap) marshalStruct(typeKey uintptr, destVal reflect.Value) ([]byte, error) {

	var buf bytes.Buffer

	buf.WriteByte('{')
	destType := destVal.Type()

	for j := 0; j < destType.NumField(); j++ {
		field := destType.Field(j)
		jsonFieldName := field.Tag.Get("json")
		var fieldByte []byte
		var err error
		if _, ok := mapper.cache[typeKey]; !ok {
			mapper.cache[typeKey] = make(map[int]map[string]string)
		}
		jsonmap, ok := mapper.cache[typeKey][j]
		var val interface{}
		if !ok {
			jsonMapTag, exist := field.Tag.Lookup(jsonMapTagKey)
			if !exist {
				val = destVal.Field(j).Interface()
			} else if jsonMapTag == "" {
				val = Wrap(destVal.Field(j).Interface())
			} else {
				pairs := strings.Split(jsonMapTag, ";")
				jsonmap = make(map[string]string)
				for _, pair := range pairs {
					kv := strings.Split(pair, ":")
					jsonmap[kv[0]] = kv[1]
				}
				val = jsonmap[fmt.Sprintf("%v", destVal.Field(j).Interface())]
				mapper.cache[typeKey][j] = jsonmap
			}
		} else {
			val = jsonmap[fmt.Sprintf("%v", destVal.Field(j).Interface())]
		}
		fieldByte, err = json.Marshal(map[string]interface{}{jsonFieldName: val})
		if err != nil {
			return nil, err
		}
		if len(fieldByte) > 0 {
			fieldByte = fieldByte[1 : len(fieldByte)-1]
		}
		if j == destType.NumField()-1 {
			_, err = buf.Write(fieldByte)
		} else {
			_, err = buf.Write(append(fieldByte, byte(',')))
		}
		if err != nil {
			return nil, err
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, Wrap(v))
}

func Marshal(v any) ([]byte, error) {
	return json.Marshal(Wrap(v))
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(Wrap(v), prefix, indent)
}
