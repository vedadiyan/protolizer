package codec

import (
	"fmt"
	"reflect"
)

type (
	CodecOptions struct {
		MapKeyWireType   WireType
		MapValueWireType WireType
	}
	CodecOption func(*CodecOptions)
)

func withMapWireTypes(key WireType, value WireType) CodecOption {
	return func(eo *CodecOptions) {
		eo.MapKeyWireType = key
		eo.MapValueWireType = value
	}
}

func Marshal(v any) ([]byte, error) {
	reflected := reflect.ValueOf(v)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	typ := RegisterType(reflected.Type())
	out := make([]byte, 0)
	for _, i := range typ.Fields {
		var opts []CodecOption
		v := reflected.FieldByIndex(i.Index)
		if v.Kind() == reflect.Map {
			opts = append(opts, withMapWireTypes(i.Tags.MapKey, i.Tags.MapValue))
		}
		bytes, err := encode(v, i.Kind, i.Tags.Protobuf.FieldNum, i.Tags.Protobuf.WireType, opts...)
		if err != nil {
			return nil, err
		}
		out = append(out, bytes...)
	}
	return out, nil
}

func encode(v reflect.Value, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...CodecOption) ([]byte, error) {
	w := wireType
	if kind == reflect.Slice {
		w = WireTypeLen
	}
	tag, err := encodeTag(int32(fieldNumber), w)
	if err != nil {
		return nil, err
	}
	if v.IsZero() {
		return nil, nil
	}
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	bytes, err := encodeRaw(v, kind, fieldNumber, wireType, opts...)
	if err != nil {
		return nil, err
	}
	return append(tag, bytes...), nil
}

func encodeRaw(v reflect.Value, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...CodecOption) ([]byte, error) {
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
						bytes, err := encodeRaw(v, v.Kind(), fieldNumber, wireType)
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
						bytes, err := encodeRaw(v, v.Kind(), fieldNumber, wireType)
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
			encodeOptions := new(CodecOptions)
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
				keyBytes, err := encodeRaw(key, key.Kind(), fieldNumber, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				value := mapRange.Value()
				valueTag, err := encodeTag(2, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				valueBytes, err := encodeRaw(value, value.Kind(), fieldNumber, encodeOptions.MapValueWireType)
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
	return nil, fmt.Errorf("")
}

func Unmarshal(bytes []byte, v any) error {
	reflected := reflect.ValueOf(v)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	typ := RegisterType(reflected.Type())
	pos := 0
	for pos < len(bytes) {
		fieldNum, wireType, consumed, err := decodeTag(bytes, pos)
		if err != nil {
			return err
		}
		_ = wireType
		pos += consumed
		for _, i := range typ.Fields {
			if i.Tags.Protobuf.FieldNum == int(fieldNum) {
				v := reflected.FieldByIndex(i.Index)
				var opts []CodecOption
				if v.Kind() == reflect.Map {
					opts = append(opts, withMapWireTypes(i.Tags.MapKey, i.Tags.MapValue))
				}
				consumed, err := decodeRaw(v, i.Kind, bytes, i.Tags.Protobuf.WireType, pos, opts...)
				if err != nil {
					return err
				}
				pos = consumed
				break
			}
		}
	}
	return nil
}

func decode(v reflect.Value, expectedFieldNumber int, kind reflect.Kind, bytes []byte, pos int, opts ...CodecOption) (int, error) {
	fieldNum, wireType, consumed, err := decodeTag(bytes, pos)
	if err != nil {
		return pos, err
	}
	if fieldNum != int32(expectedFieldNumber) {
		return 0, nil
	}
	return decodeRaw(v, kind, bytes, wireType, pos+consumed, opts...)

}

func decodeRaw(v reflect.Value, kind reflect.Kind, bytes []byte, wireType WireType, pos int, opts ...CodecOption) (int, error) {
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
			_v := reflect.New(v.Type().Elem())
			_v = _v.Elem()
			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					value, consumed, err := decodeBytes(bytes, pos)
					if err != nil {
						return pos, err
					}
					_pos := 0
					for _pos < len(value) {
						consumed, err := decodeRaw(_v, _v.Kind(), value, wireType, _pos)
						if err != nil {
							return pos, err
						}
						_pos = consumed
						v.Set(reflect.Append(v, _v))
					}
					return pos + consumed, nil
				}
			default:
				{
					consumed, err := decodeRaw(_v, _v.Kind(), bytes, wireType, pos)
					if err != nil {
						return pos, err
					}
					v.Set(reflect.Append(v, _v))
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
			_pos := 0
			_, keyWireType, consumed, err := decodeTag(value, _pos)
			if err != nil {
				return pos, err
			}
			_pos += consumed
			_key := reflect.New(keyType).Elem()
			consumed, err = decodeRaw(_key, _key.Kind(), value, keyWireType, _pos)
			if err != nil {
				return pos, err
			}
			_pos = consumed

			_, valueWireType, consumed, err := decodeTag(value, _pos)
			if err != nil {
				return pos, err
			}
			_pos += consumed
			_value := reflect.New(valueType).Elem()
			_, err = decodeRaw(_value, _value.Kind(), value, valueWireType, _pos)
			if err != nil {
				return pos, err
			}
			v.SetMapIndex(_key, _value)
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
	return pos, fmt.Errorf("")
}
