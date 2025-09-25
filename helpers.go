package protolizer

import "reflect"

func dereference(v *reflect.Value) (*reflect.Value, *reflect.Value) {
	if v.Kind() == reflect.Pointer {
		v.Set(reflect.New(v.Type().Elem()))
		elem := v.Elem()
		return &elem, v
	}
	return v, v
}
