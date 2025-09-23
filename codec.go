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
	out := make([]byte, 0)
	for _, i := range typ.Fields {
		var opts []encodeOption
		v := reflected.FieldByIndex(i.FieldIndex)
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
		if v.IsZero() {
			continue
		}
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		bytes, err := encodeValue(v, i.Kind, i.Tags.Protobuf.FieldNum, i.Tags.Protobuf.WireType, opts...)
		if err != nil {
			return nil, err
		}
		out = append(out, append(tag, bytes...)...)
	}
	return out, nil
}

func encodeValue(v reflect.Value, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...encodeOption) ([]byte, error) {
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
			var data []byte
			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					for i := 0; i < v.Len(); i++ {
						v := v.Index(i)
						bytes, err := encodeValue(v, v.Kind(), fieldNumber, wireType)
						if err != nil {
							return nil, err
						}
						data = append(data, bytes...)
					}
					return encodeBytes(data), nil
				}
			default:
				{
					for i := 0; i < v.Len(); i++ {
						if len(data) != 0 {
							tag, err := encodeTag(int32(fieldNumber), WireTypeLen)
							if err != nil {
								return nil, err
							}
							data = append(data, tag...)
						}
						v := v.Index(i)
						bytes, err := encodeValue(v, v.Kind(), fieldNumber, wireType)
						if err != nil {
							return nil, err
						}
						data = append(data, bytes...)
					}
					return data, nil
				}
			}

		}
	case reflect.Map:
		{
			encodeOptions := new(encodeOptions)
			for _, opt := range opts {
				opt(encodeOptions)
			}
			var data []byte
			mapRange := v.MapRange()
			for mapRange.Next() {
				if len(data) != 0 {
					tag, err := encodeTag(int32(fieldNumber), WireTypeLen)
					if err != nil {
						return nil, err
					}
					data = append(data, tag...)
				}
				key := mapRange.Key()
				keyTag, err := encodeTag(1, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				keyBytes, err := encodeValue(key, key.Kind(), fieldNumber, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				value := mapRange.Value()
				valueTag, err := encodeTag(2, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				valueBytes, err := encodeValue(value, value.Kind(), fieldNumber, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}

				keyEntry := append(keyTag, keyBytes...)
				valueEntry := append(valueTag, valueBytes...)
				entry := encodeBytes(append(keyEntry, valueEntry...))
				data = append(data, entry...)
			}
			return data, nil
		}
	case reflect.Struct:
		{
			return Marshal(v.Interface())
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
		v := reflected.FieldByIndex(field.FieldIndex)
		consumed, err = decodeValue(v, field.Kind, bytes, field.Tags.Protobuf.WireType, pos)
		if err != nil {
			return err
		}
		pos = consumed
	}
	return nil
}

func decodeValue(v reflect.Value, kind reflect.Kind, bytes []byte, wireType WireType, pos int) (int, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			if wireType == WireTypeI32 {
				value, consumed, err := decodeFixed32(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetInt(int64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := decodeFixed64(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetInt(value)
				return pos + consumed, nil
			}
			value, consumed, err := decodeVarint(bytes, pos)
			if err != nil {
				return pos, err
			}
			v.SetInt(int64(value))
			return pos + consumed, nil
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			if wireType == WireTypeI32 {
				value, consumed, err := decodeFixed32(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetUint(uint64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := decodeFixed64(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetUint(uint64(value))
				return pos + consumed, nil
			}
			value, consumed, err := decodeUvarint(bytes, pos)
			if err != nil {
				return pos, err
			}
			v.SetUint(value)
			return pos + consumed, nil
		}
	case reflect.Float32:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			value, consumed, err := decodeFloat32(bytes, pos)
			if err != nil {
				return pos, err
			}
			v.SetFloat(float64(value))
			return pos + consumed, nil
		}
	case reflect.Float64:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			value, consumed, err := decodeFloat64(bytes, pos)
			if err != nil {
				return pos, err
			}
			v.SetFloat(value)
			return pos + consumed, nil
		}
	case reflect.Bool:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			value, consumed, err := decodeBool(bytes, pos)
			if err != nil {
				return pos, err
			}
			v.SetBool(value)
			return pos + consumed, nil
		}
	case reflect.String:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			value, consumed, err := decodeString(bytes, pos)
			if err != nil {
				return pos, err
			}
			v.SetString(value)
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
						consumed, err := decodeValue(tmp, tmp.Kind(), value, wireType, innerPos)
						if err != nil {
							return pos, err
						}
						innerPos = consumed
						v.Set(reflect.Append(v, tmp))
					}
					return pos + consumed, nil
				}
			default:
				{
					consumed, err := decodeValue(tmp, tmp.Kind(), bytes, wireType, pos)
					if err != nil {
						return pos, err
					}
					v.Set(reflect.Append(v, tmp))
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
			consumed, err = decodeValue(key, key.Kind(), value, keyWireType, innerPos)
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
			_, err = decodeValue(val, val.Kind(), value, valueWireType, innerPos)
			if err != nil {
				return pos, err
			}
			v.SetMapIndex(key, val)
			return pos + c, nil
		}
	case reflect.Struct:
		{
			value, c, err := decodeBytes(bytes, pos)
			if err != nil {
				return pos, err
			}

			if err := Unmarshal(value, v.Interface()); err != nil {
				return c, err
			}
			return c, nil
		}
	}
	return pos, fmt.Errorf("unexpected type %v", kind)
}

func UnmarshalAnonymous(typeName string, bytes []byte) (map[string]any, error) {
	typ := CaptureTypeByName(typeName)
	out := make(map[string]any)
	pos := 0
	for pos < len(bytes) {
		fieldNum, _, consumed, err := decodeTag(bytes, pos)
		if err != nil {
			return nil, err
		}
		pos += consumed
		field, ok := typ.FieldsIndexer[int(fieldNum)]
		if !ok {
			continue
		}
		value, consumed, err := decodeValueAnonymous(field, bytes, field.Tags.Protobuf.WireType, pos)
		if err != nil {
			return nil, err
		}
		pos = consumed
		val, ok := out[field.Name]
		if !ok {
			out[field.Name] = value
			continue
		}
		switch t := val.(type) {
		case []any:
			{
				t = append(t, value)
				out[field.Name] = t
			}
		case map[string]any:
			{
				tmp, ok := value.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("expected map[string]any but got %T", value)
				}
				for key, value := range tmp {
					t[key] = value
				}
				out[field.Name] = t
			}
		default:
			{
				slice := make([]any, 0)
				slice = append(slice, t)
				slice = append(slice, value)
				out[field.Name] = slice
			}
		}
	}
	return out, nil
}

func decodeValueAnonymous(field *Field, bytes []byte, wireType WireType, pos int) (any, int, error) {
	switch field.Kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{

			if wireType == WireTypeI32 {
				value, consumed, err := decodeFixed32(bytes, pos)
				if err != nil {
					return nil, pos, err
				}
				return float64(value), pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := decodeFixed64(bytes, pos)
				if err != nil {
					return nil, pos, err
				}
				return float64(value), pos + consumed, nil
			}
			value, consumed, err := decodeVarint(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			return float64(value), pos + consumed, nil
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		{
			if wireType == WireTypeI32 {
				value, consumed, err := decodeFixed32(bytes, pos)
				if err != nil {
					return nil, pos, err
				}
				return float64(value), pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := decodeFixed64(bytes, pos)
				if err != nil {
					return nil, pos, err
				}
				return float64(value), pos + consumed, nil
			}
			value, consumed, err := decodeUvarint(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			return float64(value), pos + consumed, nil
		}
	case reflect.Float32:
		{
			value, consumed, err := decodeFloat32(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			return float64(value), pos + consumed, nil
		}
	case reflect.Float64:
		{
			value, consumed, err := decodeFloat64(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			return float64(value), pos + consumed, nil
		}
	case reflect.Bool:
		{
			value, consumed, err := decodeBool(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			return value, pos + consumed, nil
		}
	case reflect.String:
		{
			value, consumed, err := decodeString(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			return value, pos + consumed, nil
		}
	case reflect.Array, reflect.Slice:
		{
			if field.Index == reflect.Uint8 {
				value, consumed, err := decodeBytes(bytes, pos)
				if err != nil {
					return nil, pos, err
				}
				return value, pos + consumed, nil
			}
			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					value, consumed, err := decodeBytes(bytes, pos)
					if err != nil {
						return nil, pos, err
					}
					innerPos := 0
					out := make([]float64, 0)
					for innerPos < len(value) {
						value, consumed, err := decodeValueAnonymous(field, value, wireType, innerPos)
						if err != nil {
							return nil, pos, err
						}
						innerPos = consumed
						out = append(out, value.(float64))
					}
					return out, pos + consumed, nil
				}
			default:
				{
					value, consumed, err := decodeValueAnonymous(field, bytes, wireType, pos)
					if err != nil {
						return nil, pos, err
					}
					return value, consumed, nil
				}
			}
		}
	case reflect.Map:
		{
			value, c, err := decodeBytes(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			innerPos := 0
			_, keyWireType, consumed, err := decodeTag(value, innerPos)
			if err != nil {
				return nil, pos, err
			}
			innerPos += consumed
			key, consumed, err := decodeValueAnonymous(&Field{Kind: field.Key}, value, keyWireType, innerPos)
			if err != nil {
				return nil, pos, err
			}
			innerPos = consumed

			_, valueWireType, consumed, err := decodeTag(value, innerPos)
			if err != nil {
				return nil, pos, err
			}
			innerPos += consumed
			v, _, err := decodeValueAnonymous(&Field{Kind: field.Index}, value, valueWireType, innerPos)
			if err != nil {
				return nil, pos, err
			}
			return map[string]any{fmt.Sprintf("%v", key): v}, pos + c, nil
		}
	case reflect.Struct:
		{
			value, c, err := decodeBytes(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			v, err := UnmarshalAnonymous(field.TypeName, value)
			if err != nil {
				return nil, pos, err
			}
			return v, pos + c, nil
		}
	}
	return nil, pos, fmt.Errorf("unexpected type %v", field)
}
