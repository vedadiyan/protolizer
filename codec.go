package protolizer

import (
	"bytes"
	"fmt"
	"reflect"
)

type (
	Encoder func(reflect.Value, *Field, WireType) (*bytes.Buffer, error)
)

var (
	_encoders map[reflect.Kind]Encoder
)

func init() {
	_encoders = make(map[reflect.Kind]Encoder)
	_encoders[reflect.Int] = signedNumberEncoder
	_encoders[reflect.Int16] = signedNumberEncoder
	_encoders[reflect.Int32] = signedNumberEncoder
	_encoders[reflect.Int64] = signedNumberEncoder
	_encoders[reflect.Int8] = signedNumberEncoder
	_encoders[reflect.Uint] = unsignedNumberEncoder
	_encoders[reflect.Uint16] = unsignedNumberEncoder
	_encoders[reflect.Uint32] = unsignedNumberEncoder
	_encoders[reflect.Uint64] = unsignedNumberEncoder
	_encoders[reflect.Uint8] = unsignedNumberEncoder
	_encoders[reflect.Float32] = floatEncoder
	_encoders[reflect.Float64] = doubleEncoder
	_encoders[reflect.Bool] = booleanEncoder
	_encoders[reflect.String] = stringEncoder
	_encoders[reflect.Array] = arrayEncoder
	_encoders[reflect.Slice] = arrayEncoder
	_encoders[reflect.Map] = mapEncoder
	_encoders[reflect.Struct] = structEncoder
}

func Marshal(v any) ([]byte, error) {
	return marshal(reflect.ValueOf(v))
}

func marshal(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	typ := CaptureType(v.Type())
	buffer := alloc(0)
	defer dealloc(buffer)
	for _, field := range typ.Fields {
		v := v.FieldByIndex(field.FieldIndex)
		if v.IsZero() {
			continue
		}
		if field.IsPointer {
			v = v.Elem()
		}
		_, _ = buffer.Write(field.Tag)
		bytes, err := _encoders[field.Kind](v, field, field.Tags.Protobuf.WireType)
		if err != nil {
			return nil, err
		}
		_, err = bytes.WriteTo(buffer)
		dealloc(bytes)
		if err != nil {
			return nil, err
		}
	}
	return bytes.Clone(buffer.Bytes()), nil
}

func signedNumberEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	switch wireType {
	case WireTypeI32:
		{
			return fixed32Encode(int32(v.Int())), nil
		}
	case WireTypeI64:
		{
			return fixed64Encode(int64(v.Int())), nil
		}
	default:
		{
			return zigzagEncode(v.Int()), nil
		}
	}
}

func unsignedNumberEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	switch wireType {
	case WireTypeI32:
		{
			return fixed32Encode(int32(v.Uint())), nil
		}
	case WireTypeI64:
		{
			return fixed64Encode(int64(v.Uint())), nil
		}
	default:
		{
			return uvarintEncode(v.Uint()), nil
		}
	}
}

func floatEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return float32Encode(float32(v.Float())), nil
}

func doubleEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return float46Encode(v.Float()), nil
}

func booleanEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return boolEncode(v.Bool()), nil
}

func stringEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return stringEncode(v.String()), nil
}

func arrayEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	if field.Index == reflect.Uint8 {
		return bytesEncode(v.Bytes()), nil
	}

	switch wireType {
	case WireTypeVarint, WireTypeI32, WireTypeI64:
		{
			buffer := alloc(0)
			defer dealloc(buffer)
			for i := range v.Len() {
				v := v.Index(i)
				if v.Kind() == reflect.Pointer {
					v = v.Elem()
				}
				value, err := _encoders[v.Kind()](v, field, wireType)
				if err != nil {
					return nil, err
				}
				_, err = value.WriteTo(buffer)
				dealloc(value)
				if err != nil {
					return nil, err
				}
			}
			return bytesEncode(buffer.Bytes()), nil
		}
	default:
		{
			buffer := alloc(0)
			tag, err := tagEncode(int32(field.Tags.Protobuf.FieldNum), WireTypeLen)
			if err != nil {
				return nil, err
			}
			defer dealloc(tag)
			for i := range v.Len() {
				if buffer.Len() != 0 {
					_, _ = buffer.Write(tag.Bytes())
				}
				v := v.Index(i)
				if v.Kind() == reflect.Pointer {
					v = v.Elem()
				}
				value, err := _encoders[v.Kind()](v, nil, wireType)
				if err != nil {
					return nil, err
				}
				_, err = value.WriteTo(buffer)
				dealloc(value)
				if err != nil {
					return nil, err
				}
			}
			return buffer, nil
		}
	}
}

func mapEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	buffer := alloc(0)
	mapRange := v.MapRange()
	tag, err := tagEncode(int32(field.Tags.Protobuf.FieldNum), WireTypeLen)
	if err != nil {
		return nil, err
	}
	defer dealloc(tag)
	for mapRange.Next() {
		if buffer.Len() != 0 {
			_, _ = buffer.Write(tag.Bytes())
		}
		entry := alloc(0)
		key := mapRange.Key()
		if key.Kind() == reflect.Pointer {
			key = key.Elem()
		}
		_, _ = entry.Write(field.KeyTag)
		keyBytes, err := _encoders[key.Kind()](key, nil, field.Tags.MapKey)
		if err != nil {
			return nil, err
		}
		_, err = keyBytes.WriteTo(entry)
		dealloc(keyBytes)
		if err != nil {
			return nil, err
		}
		value := mapRange.Value()
		_, _ = entry.Write(field.ValueTag)
		if value.Kind() == reflect.Pointer {
			value = value.Elem()
		}
		valueBytes, err := _encoders[value.Kind()](value, nil, field.Tags.MapValue)
		if err != nil {
			return nil, err
		}
		_, err = valueBytes.WriteTo(entry)
		dealloc(valueBytes)
		if err != nil {
			return nil, err
		}
		encodedEntry := bytesEncode(entry.Bytes())
		_, err = encodedEntry.WriteTo(buffer)
		dealloc(encodedEntry)
		if err != nil {
			return nil, err
		}
		dealloc(entry)
	}
	return buffer, nil
}

func structEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	encodedStruct, err := marshal(v)
	if err != nil {
		return nil, err
	}
	return bytesEncode(encodedStruct), nil
}

func Unmarshal(bytes []byte, v any) error {
	reflected := reflect.ValueOf(v)
	buffer := alloc(0)
	buffer.Write(bytes)
	defer dealloc(buffer)
	return unmarshal(buffer, reflected)
}

func unmarshal(bytes *bytes.Buffer, v reflect.Value) error {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	typ := CaptureType(v.Type())
	pos := 0
	for pos < bytes.Len() {
		fieldNum, _, consumed, err := tagDecode(bytes, pos)
		if err != nil {
			return err
		}
		pos += consumed
		field, ok := typ.FieldsIndexer[int(fieldNum)]
		if !ok {
			continue
		}
		v := v.FieldByIndex(field.FieldIndex)
		consumed, err = decodeValue(v, field, field.Kind, bytes, field.Tags.Protobuf.WireType, pos)
		if err != nil {
			return err
		}
		pos = consumed
	}
	return nil
}

func decodeValue(v reflect.Value, field *Field, kind reflect.Kind, bytes *bytes.Buffer, wireType WireType, pos int) (int, error) {
	switch kind {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		{
			elem, _ := dereference(v)
			if wireType == WireTypeI32 {
				value, consumed, err := fixed32Decode(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetInt(int64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := fixed64Decode(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetInt(value)
				return pos + consumed, nil
			}
			value, consumed, err := zigzagDecode(bytes, pos)
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
				value, consumed, err := fixed32Decode(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetUint(uint64(value))
				return pos + consumed, nil
			}
			if wireType == WireTypeI64 {
				value, consumed, err := fixed64Decode(bytes, pos)
				if err != nil {
					return pos, err
				}
				elem.SetUint(uint64(value))
				return pos + consumed, nil
			}
			value, consumed, err := uvarintDecode(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetUint(value)
			return pos + consumed, nil
		}
	case reflect.Float32:
		{
			elem, _ := dereference(v)
			value, consumed, err := float32Decode(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetFloat(float64(value))
			return pos + consumed, nil
		}
	case reflect.Float64:
		{
			elem, _ := dereference(v)
			value, consumed, err := float64Decode(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetFloat(value)
			return pos + consumed, nil
		}
	case reflect.Bool:
		{
			elem, _ := dereference(v)
			value, consumed, err := boolDecode(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetBool(value)
			return pos + consumed, nil
		}
	case reflect.String:
		{
			elem, _ := dereference(v)
			value, consumed, err := stringDecode(bytes, pos)
			if err != nil {
				return pos, err
			}
			elem.SetString(value)
			return pos + consumed, nil
		}
	case reflect.Array, reflect.Slice:
		{
			k := field.Index
			if k == reflect.Uint8 {
				value, consumed, err := bytesDecode(bytes, pos)
				if err != nil {
					return pos, err
				}
				v.SetBytes(append([]byte{}, value.Bytes()...))
				dealloc(value)
				return pos + consumed, nil
			}
			tmp := reflect.New(v.Type().Elem())
			tmp = tmp.Elem()
			switch wireType {
			case WireTypeVarint, WireTypeI32, WireTypeI64:
				{
					value, consumed, err := bytesDecode(bytes, pos)
					if err != nil {
						return pos, err
					}
					innerPos := 0
					for innerPos < value.Len() {
						elem, addr := dereference(tmp)
						consumed, err := decodeValue(elem, nil, elem.Kind(), value, wireType, innerPos)
						if err != nil {
							return pos, err
						}
						innerPos = consumed
						v.Set(reflect.Append(v, addr))
					}
					dealloc(value)
					return pos + consumed, nil
				}
			default:
				{
					elem, addr := dereference(tmp)
					consumed, err := decodeValue(elem, nil, elem.Kind(), bytes, wireType, pos)
					if err != nil {
						return pos, err
					}
					v.Set(reflect.Append(v, addr))
					return consumed, nil
				}
			}
		}
	case reflect.Map:
		{
			value, c, err := bytesDecode(bytes, pos)
			if err != nil {
				return pos, err
			}
			keyType := v.Type().Key()
			valueType := v.Type().Elem()
			if v.IsZero() {
				v.Set(reflect.MakeMap(reflect.MapOf(keyType, valueType)))
			}
			innerPos := 0
			_, keyWireType, consumed, err := tagDecode(value, innerPos)
			if err != nil {
				return pos, err
			}
			innerPos += consumed
			key := reflect.New(keyType).Elem()
			consumed, err = decodeValue(key, nil, key.Kind(), value, keyWireType, innerPos)
			if err != nil {
				return pos, err
			}
			innerPos = consumed

			_, valueWireType, consumed, err := tagDecode(value, innerPos)
			if err != nil {
				return pos, err
			}
			innerPos += consumed
			val := reflect.New(valueType).Elem()
			elem, addr := dereference(val)
			_, err = decodeValue(elem, nil, elem.Kind(), value, valueWireType, innerPos)
			if err != nil {
				return pos, err
			}
			v.SetMapIndex(key, addr)
			dealloc(value)
			return pos + c, nil
		}
	case reflect.Struct:
		{
			elem, _ := dereference(v)
			value, c, err := bytesDecode(bytes, pos)
			if err != nil {
				return pos, err
			}
			if err := unmarshal(value, elem); err != nil {
				return c, err
			}
			dealloc(value)
			return pos + c, nil
		}
	}
	return pos, fmt.Errorf("unexpected type %v", kind)
}
