package util

import "reflect"

func GetElemenType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	return t
}
