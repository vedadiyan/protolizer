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

func WithMapWireTypes(key WireType, value WireType) CodecOption {
	return func(eo *CodecOptions) {
		eo.MapKeyWireType = key
		eo.MapValueWireType = value
	}
}

func EncodeField(n int32, wireType WireType, v []byte) ([]byte, error) {
	tagBytes, err := EncodeTag(n, wireType)
	if err != nil {
		return nil, err
	}
	return append(tagBytes, v...), nil
}

func Encode(v reflect.Value, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...CodecOption) ([]byte, error) {
	tag, err := EncodeTag(int32(fieldNumber), wireType)
	if err != nil {
		return nil, err
	}
	if v.IsZero() {
		return nil, nil
	}
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	bytes, err := RawEncode(v, kind, fieldNumber, wireType, opts...)
	if err != nil {
		return nil, err
	}
	return append(tag, bytes...), nil
}

func RawEncode(v reflect.Value, kind reflect.Kind, fieldNumber int, wireType WireType, opts ...CodecOption) ([]byte, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			if wireType == WireTypeI32 {
				return EncodeFixed32(int32(v.Int())), nil
			}
			if wireType == WireTypeI64 {
				return EncodeFixed64(int64(v.Int())), nil
			}
			return EncodeVarint(v.Int()), nil
		}
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		{
			if wireType == WireTypeI32 {
				return EncodeFixed32(int32(v.Uint())), nil
			}
			if wireType == WireTypeI64 {
				return EncodeFixed64(int64(v.Uint())), nil
			}
			return EncodeUvarint(v.Uint()), nil
		}
	case reflect.Float32:
		{
			return EncodeFloat32(float32(v.Float())), nil
		}
	case reflect.Float64:
		{
			return EncodeFloat64(v.Float()), nil
		}
	case reflect.Bool:
		{
			return EncodeBool(v.Bool()), nil
		}
	case reflect.String:
		{
			return EncodeString(v.String()), nil
		}
	case reflect.Array, reflect.Slice:
		{
			k := v.Type().Elem().Kind()
			if k == reflect.Uint8 {
				return EncodeBytes(v.Bytes()), nil
			}
			var data []byte
			for i := 0; i < v.Len(); i++ {
				v := v.Index(i)
				tag, err := EncodeTag(int32(fieldNumber), wireType)
				if err != nil {
					return nil, err
				}
				bytes, err := RawEncode(v, v.Kind(), fieldNumber, wireType)
				if err != nil {
					return nil, err
				}
				data = append(data, append(tag, bytes...)...)
			}
			return data, nil
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
				key := mapRange.Key()
				keyTag, err := EncodeTag(1, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				keyBytes, err := RawEncode(key, key.Kind(), fieldNumber, encodeOptions.MapKeyWireType)
				if err != nil {
					return nil, err
				}
				value := mapRange.Value()
				valueTag, err := EncodeTag(2, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				valueBytes, err := RawEncode(value, value.Kind(), fieldNumber, encodeOptions.MapValueWireType)
				if err != nil {
					return nil, err
				}
				data = append(data, append(append(keyTag, keyBytes...), append(valueTag, valueBytes...)...)...)
			}
			return EncodeBytes(data), nil
		}
	case reflect.Struct:
		{
			typ := RegisterType(v.Type())
			var data []byte
			for _, i := range typ.Fields {
				bytes, err := i.Encode(v.FieldByIndex(i.Index))
				if err != nil {
					return nil, err
				}
				data = append(data, bytes...)
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("")
}

func Decode(v reflect.Value, expectedFieldNumber int, kind reflect.Kind, bytes []byte, pos int, opts ...CodecOption) (int, error) {
	fieldNum, wireType, consumed, err := DecodeTag(bytes, pos)
	if err != nil {
		return pos, err
	}
	if fieldNum != int32(expectedFieldNumber) {
		return 0, nil
	}
	return RawDecode(v, kind, bytes, wireType, pos+consumed, opts...)

}

func RawDecode(v reflect.Value, kind reflect.Kind, bytes []byte, wireType WireType, pos int, opts ...CodecOption) (int, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			if v.Kind() == reflect.Pointer {
				v.Set(reflect.New(v.Type().Elem()))
				v = v.Elem()
			}
			if wireType == WireTypeI32 {
				value, consumed, err := DecodeFixed32(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetInt(int64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := DecodeFixed64(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetInt(value)
				return pos + consumed, nil
			}
			value, consumed, err := DecodeVarint(bytes, pos)
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
				value, consumed, err := DecodeFixed32(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetUint(uint64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := DecodeFixed64(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetUint(uint64(value))
				return pos + consumed, nil
			}
			value, consumed, err := DecodeUvarint(bytes, pos)
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
			value, consumed, err := DecodeFloat32(bytes, pos)
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
			value, consumed, err := DecodeFloat64(bytes, pos)
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
			value, consumed, err := DecodeBool(bytes, pos)
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
			value, consumed, err := DecodeString(bytes, pos)
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
				value, consumed, err := DecodeBytes(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetBytes(value)
				return pos + consumed, nil
			}
		}
	case reflect.Map:
		{
			value, consumed, err := DecodeBytes(bytes, pos)
			if err != nil {
				return pos, err
			}
			keyType := v.Type().Key()
			valueType := v.Type().Elem()
			v.Set(reflect.MakeMap(reflect.MapOf(keyType, valueType)))

			_ = valueType
			_pos := 0
			for _pos < len(value) {
				_, keyWireType, consumed, err := DecodeTag(value, _pos)
				if err != nil {
					return pos, err
				}
				_pos += consumed
				_key := reflect.New(keyType).Elem()
				consumed, err = RawDecode(_key, _key.Kind(), value, keyWireType, _pos)
				if err != nil {
					return pos, err
				}
				_pos = consumed
				test := _key.Int()
				_ = test

				_, valueWireType, consumed, err := DecodeTag(value, _pos)
				if err != nil {
					return pos, err
				}
				_pos += consumed
				_value := reflect.New(valueType).Elem()
				consumed, err = RawDecode(_value, _value.Kind(), value, valueWireType, _pos)
				if err != nil {
					return pos, err
				}
				_pos = consumed
				v.SetMapIndex(_key, _value)
			}
			return pos + consumed, nil
		}
	case reflect.Struct:
		{
			val := reflect.ValueOf(v.Interface())
			if val.Kind() == reflect.Pointer {
				val = v.Elem()
			}
			typ := RegisterType(val.Type())

			for _, i := range typ.Fields {
				consumed, err := i.Decode(bytes, val.FieldByIndex(i.Index), pos)
				if err != nil {
					return consumed, err
				}
				if consumed != 0 {
					pos = consumed
				}
				if pos >= len(bytes) {
					break
				}
			}
			return pos, nil
		}
	}
	return pos, fmt.Errorf("")
}
