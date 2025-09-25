package protolizer

import (
	"fmt"
	"reflect"
)

func Read(typeName string, bytes []byte) (map[string]any, error) {
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
				t = append(t, value.([]any)...)
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
		case map[float64]any:
			{
				tmp, ok := value.(map[float64]any)
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
				return nil, fmt.Errorf("unexpected type %T", val)
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
						value, consumed, err := decodeValueAnonymous(&Field{Kind: field.Index, TypeName: field.IndexType}, value, wireType, innerPos)
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
					value, consumed, err := decodeValueAnonymous(&Field{Kind: field.Index, TypeName: field.IndexType}, bytes, wireType, pos)
					if err != nil {
						return nil, pos, err
					}
					return []any{value}, consumed, nil
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
			key, consumed, err := decodeValueAnonymous(&Field{Kind: field.Key, TypeName: field.KeyType}, value, keyWireType, innerPos)
			if err != nil {
				return nil, pos, err
			}
			innerPos = consumed

			_, valueWireType, consumed, err := decodeTag(value, innerPos)
			if err != nil {
				return nil, pos, err
			}
			innerPos += consumed
			v, _, err := decodeValueAnonymous(&Field{Kind: field.Index, TypeName: field.IndexType}, value, valueWireType, innerPos)
			if err != nil {
				return nil, pos, err
			}
			if keyWireType == WireTypeVarint || keyWireType == WireTypeI32 || keyWireType == WireTypeI64 {
				return map[float64]any{key.(float64): v}, pos + c, nil
			}
			return map[any]any{key: v}, pos + c, nil
		}
	case reflect.Struct:
		{
			value, c, err := decodeBytes(bytes, pos)
			if err != nil {
				return nil, pos, err
			}
			v, err := Read(field.TypeName, value)
			if err != nil {
				return nil, pos, err
			}
			return v, pos + c, nil
		}
	}
	return nil, pos, fmt.Errorf("unexpected type %v", field)
}

func Write(typeName string, v map[string]any) ([]byte, error) {
	typ := CaptureTypeByName(typeName)
	out := make([]byte, 0)
	for _, i := range typ.Fields {
		var opts []encodeOption
		value, ok := v[i.Name]
		if !ok {
			continue
		}
		v := reflect.ValueOf(value)
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
		bytes, err := encodeValueAnonymous(&v, i, i.Kind, i.Tags.Protobuf.FieldNum, i.Tags.Protobuf.WireType, opts...)
		if err != nil {
			return nil, err
		}
		out = append(out, append(tag, bytes...)...)
	}
	return out, nil
}

func encodeValueAnonymous(v *reflect.Value, field *Field, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...encodeOption) ([]byte, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			if wireType == WireTypeI32 {
				return encodeFixed32(int32(v.Float())), nil
			}
			if wireType == WireTypeI64 {
				return encodeFixed64(int64(v.Float())), nil
			}
			return encodeVarint(int64(v.Float())), nil
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		{
			if wireType == WireTypeI32 {
				return encodeFixed32(int32(v.Float())), nil
			}
			if wireType == WireTypeI64 {
				return encodeFixed64(int64(v.Float())), nil
			}
			return encodeUvarint(uint64(v.Float())), nil
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
			k := field.Index
			if k == reflect.Uint8 {
				return encodeBytes(v.Bytes()), nil
			}
			var data []byte
			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					for i := 0; i < v.Len(); i++ {
						v := v.Index(i)
						if v.Kind() == reflect.Pointer {
							v = v.Elem()
						}
						bytes, err := encodeValueAnonymous(&v, field, field.Index, fieldNumber, wireType)
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
						if v.Kind() == reflect.Pointer {
							v = v.Elem()
						}
						bytes, err := encodeValueAnonymous(&v, field, field.Index, fieldNumber, wireType)
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
				if key.Kind() == reflect.Pointer {
					key = key.Elem()
				}
				keyTag, err := encodeTag(1, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				keyBytes, err := encodeValueAnonymous(&key, field, field.Key, fieldNumber, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				value := mapRange.Value()
				valueTag, err := encodeTag(2, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				if value.Kind() == reflect.Pointer {
					value = value.Elem()
				}
				valueBytes, err := encodeValueAnonymous(&value, field, field.Index, fieldNumber, encodeOptions.MapValueWireType)
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

			typeName := field.TypeName
			if len(field.IndexType) != 0 {
				typeName = field.IndexType
			}
			data, err := Write(typeName, v.Interface().(map[string]any))
			if err != nil {
				return nil, err
			}
			out := encodeBytes(data)
			return out, nil

		}
	case reflect.Interface:
		{
			elem := v.Elem()
			return encodeValueAnonymous(&elem, field, field.Kind, fieldNumber, wireType)
		}
	}
	return nil, fmt.Errorf("unexpected type %v", kind)
}
