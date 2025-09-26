package protolizer

import (
	"fmt"
	"reflect"
)

type (
	encodeOptions struct {
		MapKeyWireType   WireType
		MapValueWireType WireType
	}
	encodeOption func(*encodeOptions)
)

func withMapWireTypes(key WireType, value WireType) encodeOption {
	return func(eo *encodeOptions) {
		eo.MapKeyWireType = key
		eo.MapValueWireType = value
	}
}

func Marshal(v any) ([]byte, error) {
	reflected := reflect.ValueOf(v)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	typ := CaptureType(reflected.Type())
	memory := alloc(0)
	defer dealloc(memory)
	for _, i := range typ.Fields {
		var opts []encodeOption
		v := reflected.FieldByIndex(i.FieldIndex)
		if v.IsZero() {
			continue
		}
		if v.Kind() == reflect.Map {
			opts = append(opts, withMapWireTypes(i.Tags.MapKey, i.Tags.MapValue))
		}
		w := i.Tags.Protobuf.WireType
		if i.Kind == reflect.Slice {
			w = WireTypeLen
		}
		tag, err := encodeTag(int32(i.Tags.Protobuf.FieldNum), w)
		if err != nil {
			return nil, err
		}
		memory.Write(tag)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		bytes, err := encodeValue(&v, i.Kind, i.Tags.Protobuf.FieldNum, i.Tags.Protobuf.WireType, opts...)
		if err != nil {
			return nil, err
		}
		memory.Write(bytes)
	}
	return append([]byte{}, memory.Bytes()...), nil
}

func encodeValue(v *reflect.Value, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...encodeOption) ([]byte, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			if wireType == WireTypeI32 {
				return encodeFixed32(int32(v.Int())), nil
			}
			if wireType == WireTypeI64 {
				return encodeFixed64(int64(v.Int())), nil
			}
			return encodeVarint(v.Int()), nil
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		{
			if wireType == WireTypeI32 {
				return encodeFixed32(int32(v.Uint())), nil
			}
			if wireType == WireTypeI64 {
				return encodeFixed64(int64(v.Uint())), nil
			}
			return encodeUvarint(v.Uint()), nil
		}
	case reflect.Float32:
		{
			return encodeFloat32(float32(v.Float())), nil
		}
	case reflect.Float64:
		{
			return encodeFloat64(v.Float()), nil
		}
	case reflect.Bool:
		{
			return encodeBool(v.Bool()), nil
		}
	case reflect.String:
		{
			return encodeString(v.String()), nil
		}
	case reflect.Array, reflect.Slice:
		{
			k := v.Type().Elem().Kind()
			if k == reflect.Uint8 {
				return encodeBytes(v.Bytes()), nil
			}

			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					data := alloc(0)
					defer dealloc(data)
					for i := 0; i < v.Len(); i++ {
						v := v.Index(i)
						if v.Kind() == reflect.Pointer {
							v = v.Elem()
						}
						bytes, err := encodeValue(&v, v.Kind(), fieldNumber, wireType)
						if err != nil {
							return nil, err
						}
						data.Write(bytes)
					}
					return encodeBytes(data.Bytes()), nil
				}
			default:
				{
					data := alloc(0)
					defer dealloc(data)
					for i := 0; i < v.Len(); i++ {
						if data.Len() != 0 {
							tag, err := encodeTag(int32(fieldNumber), WireTypeLen)
							if err != nil {
								return nil, err
							}
							data.Write(tag)
						}
						v := v.Index(i)
						if v.Kind() == reflect.Pointer {
							v = v.Elem()
						}
						bytes, err := encodeValue(&v, v.Kind(), fieldNumber, wireType)
						if err != nil {
							return nil, err
						}
						data.Write(bytes)
					}
					return data.Bytes(), nil
				}
			}

		}
	case reflect.Map:
		{
			encodeOptions := new(encodeOptions)
			for _, opt := range opts {
				opt(encodeOptions)
			}
			data := alloc(0)
			defer dealloc(data)
			mapRange := v.MapRange()
			for mapRange.Next() {
				if data.Len() != 0 {
					tag, err := encodeTag(int32(fieldNumber), WireTypeLen)
					if err != nil {
						return nil, err
					}
					data.Write(tag)
				}
				keyValue := alloc(0)
				key := mapRange.Key()
				if key.Kind() == reflect.Pointer {
					key = key.Elem()
				}
				keyTag, err := encodeTag(1, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				keyValue.Write(keyTag)
				keyBytes, err := encodeValue(&key, key.Kind(), fieldNumber, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				keyValue.Write(keyBytes)
				value := mapRange.Value()
				valueTag, err := encodeTag(2, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				keyValue.Write(valueTag)
				if value.Kind() == reflect.Pointer {
					value = value.Elem()
				}
				valueBytes, err := encodeValue(&value, value.Kind(), fieldNumber, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				keyValue.Write(valueBytes)
				entry := encodeBytes(keyValue.Bytes())
				dealloc(keyValue)
				data.Write(entry)
			}
			return data.Bytes(), nil
		}
	case reflect.Struct:
		{

			data, err := Marshal(v.Interface())
			if err != nil {
				return nil, err
			}
			out := encodeBytes(data)
			return out, nil
		}
	}
	return nil, fmt.Errorf("unexpected type %v", kind)
}

func Unmarshal(bytes []byte, v any) error {
	reflected := reflect.ValueOf(v)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}

	typ := CaptureType(reflected.Type())
	pos := 0
	for pos < len(bytes) {
		fieldNum, _, consumed, err := decodeTag(bytes, pos)
		if err != nil {
			return err
		}
		pos += consumed
		field, ok := typ.FieldsIndexer[int(fieldNum)]
		if !ok {
			continue
		}
		v2 := reflected.FieldByIndex(field.FieldIndex)
		consumed, err = decodeValue(&v2, field.Kind, bytes, field.Tags.Protobuf.WireType, pos)
		if err != nil {
			return err
		}
		pos = consumed
	}
	return nil
}

func decodeValue(v *reflect.Value, kind reflect.Kind, bytes []byte, wireType WireType, pos int) (int, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			elem, _ := dereference(v)
			if wireType == WireTypeI32 {
				value, consumed, err := decodeFixed32(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetInt(int64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := decodeFixed64(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetInt(value)
				return pos + consumed, nil
			}
			value, consumed, err := decodeVarint(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetInt(int64(value))
			return pos + consumed, nil
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		{
			elem, _ := dereference(v)
			if wireType == WireTypeI32 {
				value, consumed, err := decodeFixed32(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetUint(uint64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := decodeFixed64(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetUint(uint64(value))
				return pos + consumed, nil
			}
			value, consumed, err := decodeUvarint(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetUint(value)
			return pos + consumed, nil
		}
	case reflect.Float32:
		{
			elem, _ := dereference(v)
			value, consumed, err := decodeFloat32(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetFloat(float64(value))
			return pos + consumed, nil
		}
	case reflect.Float64:
		{
			elem, _ := dereference(v)
			value, consumed, err := decodeFloat64(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetFloat(value)
			return pos + consumed, nil
		}
	case reflect.Bool:
		{
			elem, _ := dereference(v)
			value, consumed, err := decodeBool(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetBool(value)
			return pos + consumed, nil
		}
	case reflect.String:
		{
			elem, _ := dereference(v)
			value, consumed, err := decodeString(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetString(value)
			return pos + consumed, nil
		}
	case reflect.Array, reflect.Slice:
		{
			k := v.Type().Elem().Kind()
			if k == reflect.Uint8 {
				value, consumed, err := decodeBytes(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetBytes(value)
				return pos + consumed, nil
			}
			if v.IsZero() {
				v.Set(reflect.MakeSlice(v.Type(), 0, 0))
			}
			tmp := reflect.New(v.Type().Elem())
			tmp = tmp.Elem()
			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					value, consumed, err := decodeBytes(bytes, pos)
					if err != nil {
						return pos, err
					}
					innerPos := 0
					for innerPos < len(value) {
						elem, addr := dereference(&tmp)
						consumed, err := decodeValue(elem, elem.Kind(), value, wireType, innerPos)
						if err != nil {
							return pos, err
						}
						innerPos = consumed
						v.Set(reflect.Append(*v, *addr))
					}
					return pos + consumed, nil
				}
			default:
				{
					elem, addr := dereference(&tmp)
					consumed, err := decodeValue(elem, elem.Kind(), bytes, wireType, pos)
					if err != nil {
						return pos, err
					}
					v.Set(reflect.Append(*v, *addr))
					return consumed, nil
				}
			}
		}
	case reflect.Map:
		{
			value, c, err := decodeBytes(bytes, pos)
			if err != nil {
				return pos, err
			}
			keyType := v.Type().Key()
			valueType := v.Type().Elem()
			if v.IsZero() {
				v.Set(reflect.MakeMap(reflect.MapOf(keyType, valueType)))
			}
			innerPos := 0
			_, keyWireType, consumed, err := decodeTag(value, innerPos)
			if err != nil {
				return pos, err
			}
			innerPos += consumed
			key := reflect.New(keyType).Elem()
			consumed, err = decodeValue(&key, key.Kind(), value, keyWireType, innerPos)
			if err != nil {
				return pos, err
			}
			innerPos = consumed

			_, valueWireType, consumed, err := decodeTag(value, innerPos)
			if err != nil {
				return pos, err
			}
			innerPos += consumed
			val := reflect.New(valueType).Elem()
			elem, addr := dereference(&val)
			_, err = decodeValue(elem, elem.Kind(), value, valueWireType, innerPos)
			if err != nil {
				return pos, err
			}
			v.SetMapIndex(key, *addr)
			return pos + c, nil
		}
	case reflect.Struct:
		{
			elem, _ := dereference(v)
			value, c, err := decodeBytes(bytes, pos)
			if err != nil {
				return pos, err
			}
			if err := Unmarshal(value, elem.Addr().Interface()); err != nil {
				return c, err
			}
			return pos + c, nil
		}
	}
	return pos, fmt.Errorf("unexpected type %v", kind)
}
