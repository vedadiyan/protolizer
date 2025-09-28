package protolizer

import (
	"bytes"
	"reflect"
)

type (
	Encoder func(reflect.Value, *Field, WireType) (*bytes.Buffer, error)
	Decoder func(reflect.Value, *Field, *bytes.Buffer, WireType) error
)

var (
	_encoders map[reflect.Kind]Encoder
	_decoders map[reflect.Kind]Decoder
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
	_decoders = make(map[reflect.Kind]Decoder)
	_decoders[reflect.Int] = signedNumberDecoder
	_decoders[reflect.Int16] = signedNumberDecoder
	_decoders[reflect.Int32] = signedNumberDecoder
	_decoders[reflect.Int64] = signedNumberDecoder
	_decoders[reflect.Int8] = signedNumberDecoder
	_decoders[reflect.Uint] = unsignedNumberDecoder
	_decoders[reflect.Uint16] = unsignedNumberDecoder
	_decoders[reflect.Uint32] = unsignedNumberDecoder
	_decoders[reflect.Uint64] = unsignedNumberDecoder
	_decoders[reflect.Uint8] = unsignedNumberDecoder
	_decoders[reflect.Float32] = floatDecoder
	_decoders[reflect.Float64] = doubleDecoder
	_decoders[reflect.Bool] = booleanDecoder
	_decoders[reflect.String] = stringDecoder
	_decoders[reflect.Array] = arrayDecoder
	_decoders[reflect.Slice] = arrayDecoder
	_decoders[reflect.Map] = mapDecoder
	_decoders[reflect.Struct] = structDecoder

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
		fieldNum, _, err := tagDecode(bytes)
		if err != nil {
			return err
		}
		field, ok := typ.FieldsIndexer[int(fieldNum)]
		if !ok {
			continue
		}
		v := v.FieldByIndex(field.FieldIndex)
		err = _decoders[field.Kind](v, field, bytes, field.Tags.Protobuf.WireType)
		if err != nil {
			return err
		}
	}
	return nil
}

func signedNumberDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	switch wireType {
	case WireTypeI32:
		{
			value, err := fixed32Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(int64(value))
			return nil
		}
	case WireTypeI64:
		{
			value, err := fixed64Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(value)
			return nil
		}
	default:
		{
			value, err := zigzagDecode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(int64(value))
			return nil
		}
	}
}

func unsignedNumberDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	switch wireType {
	case WireTypeI32:
		{
			value, err := fixed32Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(int64(value))
			return nil
		}
	case WireTypeI64:
		{
			value, err := fixed64Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetUint(uint64(value))
			return nil
		}
	default:
		{
			value, err := uvarintDecode(bytes)
			if err != nil {
				return err
			}
			elem.SetUint(value)
			return nil
		}
	}
}

func floatDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := float32Decode(bytes)
	if err != nil {
		return err
	}
	elem.SetFloat(float64(value))
	return nil
}

func doubleDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := float64Decode(bytes)
	if err != nil {
		return err
	}
	elem.SetFloat(value)
	return nil
}

func booleanDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := boolDecode(bytes)
	if err != nil {
		return err
	}
	elem.SetBool(value)
	return nil
}

func stringDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := stringDecode(bytes)
	if err != nil {
		return err
	}
	elem.SetString(value)
	return nil
}

func arrayDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	k := field.Index
	if k == reflect.Uint8 {
		value, err := bytesDecode(bytes)
		if err != nil {
			return err
		}
		v.SetBytes(append([]byte{}, value.Bytes()...))
		dealloc(value)
		return nil
	}
	tmp := reflect.New(v.Type().Elem())
	tmp = tmp.Elem()
	switch wireType {
	case WireTypeVarint, WireTypeI32, WireTypeI64:
		{
			value, err := bytesDecode(bytes)
			if err != nil {
				return err
			}
			innerPos := 0
			for innerPos < value.Len() {
				elem, addr := dereference(tmp)
				err := _decoders[elem.Kind()](elem, nil, value, wireType)
				if err != nil {
					return err
				}
				v.Set(reflect.Append(v, addr))
			}
			dealloc(value)
			return nil
		}
	default:
		{
			elem, addr := dereference(tmp)
			err := _decoders[elem.Kind()](elem, nil, bytes, wireType)
			if err != nil {
				return err
			}
			v.Set(reflect.Append(v, addr))
			return nil
		}
	}
}

func mapDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	value, err := bytesDecode(bytes)
	if err != nil {
		return err
	}
	keyType := v.Type().Key()
	valueType := v.Type().Elem()
	if v.IsZero() {
		v.Set(reflect.MakeMap(reflect.MapOf(keyType, valueType)))
	}
	_, keyWireType, err := tagDecode(value)
	if err != nil {
		return err
	}
	key := reflect.New(keyType).Elem()
	err = _decoders[key.Kind()](key, nil, value, keyWireType)
	if err != nil {
		return err
	}

	_, valueWireType, err := tagDecode(value)
	if err != nil {
		return err
	}
	val := reflect.New(valueType).Elem()
	elem, addr := dereference(val)
	err = _decoders[elem.Kind()](elem, nil, value, valueWireType)
	if err != nil {
		return err
	}
	v.SetMapIndex(key, addr)
	dealloc(value)
	return nil
}

func structDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := bytesDecode(bytes)
	if err != nil {
		return err
	}
	if err := unmarshal(value, elem); err != nil {
		return err
	}
	dealloc(value)
	return nil
}
