package protolizer

import (
	"fmt"
	"reflect"

	"github.com/vedadiyan/protolizer/metadata"
)

func ExportType[T any]() ([]byte, error) {
	t := metadata.CaptureTypeFor[T]()
	return DynamicCodec().Marshal(t)
}

func ImportType(bytes []byte) (*metadata.Type, error) {
	t := new(metadata.Type)
	if err := DynamicCodec().Unmarshal(bytes, t); err != nil {
		return nil, err
	}
	return t, nil
}

func exportModule(t reflect.Type) (*metadata.Module, error) {
	module := new(metadata.Module)
	module.Types = make(map[string]metadata.Type)
	typ := metadata.CaptureType(t)
	if typ == nil {
		return nil, fmt.Errorf("type not found")
	}
	module.Types[metadata.TypeName(t)] = *typ
	for i := range t.NumField() {
		fieldType := t.Field(i).Type
		if fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			modules, err := exportModule(fieldType)
			if err != nil {
				return nil, err
			}
			for key, value := range modules.Types {
				module.Types[key] = value
			}
			continue
		}
	}
	return module, nil
}

func ExportModule[T any]() ([]byte, error) {
	modules, err := exportModule(reflect.TypeFor[T]())
	if err != nil {
		return nil, err
	}
	return DynamicCodec().Marshal(modules)
}

func ImportModule(bytes []byte) (*metadata.Module, error) {
	module := new(metadata.Module)
	err := DynamicCodec().Unmarshal(bytes, module)
	if err != nil {
		return nil, err
	}
	return module, nil
}
