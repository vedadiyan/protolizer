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
	buffer := Alloc(0)
	defer Dealloc(buffer)
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
		Dealloc(bytes)
		if err != nil {
			return nil, err
		}
	}
	return bytes.Clone(buffer.Bytes()), nil
}

func SignedNumberEncoder(v int64, field *Field) (*bytes.Buffer, error) {
	switch field.Tags.Protobuf.WireType {
	case WireTypeI32:
		{
			return Fixed32Encode(int32(v)), nil
		}
	case WireTypeI64:
		{
			return Fixed64Encode(v), nil
		}
	default:
		{
			return ZigzagEncode(v), nil
		}
	}
}

func signedNumberEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	switch wireType {
	case WireTypeI32:
		{
			return Fixed32Encode(int32(v.Int())), nil
		}
	case WireTypeI64:
		{
			return Fixed64Encode(int64(v.Int())), nil
		}
	default:
		{
			return ZigzagEncode(v.Int()), nil
		}
	}
}

func UnsignedNumberEncoder(v uint64, field *Field) (*bytes.Buffer, error) {
	switch field.Tags.Protobuf.WireType {
	case WireTypeI32:
		{
			return Fixed32Encode(int32(v)), nil
		}
	case WireTypeI64:
		{
			return Fixed64Encode(int64(v)), nil
		}
	default:
		{
			return UvarintEncode(v), nil
		}
	}
}

func unsignedNumberEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	switch wireType {
	case WireTypeI32:
		{
			return Fixed32Encode(int32(v.Uint())), nil
		}
	case WireTypeI64:
		{
			return Fixed64Encode(int64(v.Uint())), nil
		}
	default:
		{
			return UvarintEncode(v.Uint()), nil
		}
	}
}

func FloatEncoder(v float32, field *Field) (*bytes.Buffer, error) {
	return Float32Encode(v), nil
}

func floatEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return Float32Encode(float32(v.Float())), nil
}

func DoubleEncoder(v float64, field *Field) (*bytes.Buffer, error) {
	return Float46Encode(v), nil
}

func doubleEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return Float46Encode(v.Float()), nil
}

func BooleanEncoder(v bool, field *Field) (*bytes.Buffer, error) {
	return BoolEncode(v), nil
}

func booleanEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return BoolEncode(v.Bool()), nil
}

func StringEncoder(v string, field *Field) (*bytes.Buffer, error) {
	return StringEncode(v), nil
}

func stringEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	return StringEncode(v.String()), nil
}

func arrayEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	if field.Index == reflect.Uint8 {
		return BytesEncode(v.Bytes()), nil
	}

	switch wireType {
	case WireTypeVarint, WireTypeI32, WireTypeI64:
		{
			buffer := Alloc(0)
			defer Dealloc(buffer)
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
				Dealloc(value)
				if err != nil {
					return nil, err
				}
			}
			return BytesEncode(buffer.Bytes()), nil
		}
	default:
		{
			buffer := Alloc(0)
			tag, err := TagEncode(int32(field.Tags.Protobuf.FieldNum), WireTypeLen)
			if err != nil {
				return nil, err
			}
			defer Dealloc(tag)
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
				Dealloc(value)
				if err != nil {
					return nil, err
				}
			}
			return buffer, nil
		}
	}
}

func mapEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	buffer := Alloc(0)
	mapRange := v.MapRange()
	tag, err := TagEncode(int32(field.Tags.Protobuf.FieldNum), WireTypeLen)
	if err != nil {
		return nil, err
	}
	defer Dealloc(tag)
	for mapRange.Next() {
		if buffer.Len() != 0 {
			_, _ = buffer.Write(tag.Bytes())
		}
		entry := Alloc(0)
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
		Dealloc(keyBytes)
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
		Dealloc(valueBytes)
		if err != nil {
			return nil, err
		}
		encodedEntry := BytesEncode(entry.Bytes())
		_, err = encodedEntry.WriteTo(buffer)
		Dealloc(encodedEntry)
		if err != nil {
			return nil, err
		}
		Dealloc(entry)
	}
	return buffer, nil
}

func structEncoder(v reflect.Value, field *Field, wireType WireType) (*bytes.Buffer, error) {
	encodedStruct, err := marshal(v)
	if err != nil {
		return nil, err
	}
	return BytesEncode(encodedStruct), nil
}

func Unmarshal(bytes []byte, v any) error {
	reflected := reflect.ValueOf(v)
	buffer := Alloc(0)
	buffer.Write(bytes)
	defer Dealloc(buffer)
	return unmarshal(buffer, reflected)
}

func unmarshal(bytes *bytes.Buffer, v reflect.Value) error {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	typ := CaptureType(v.Type())
	for bytes.Len() != 0 {
		fieldNum, _, err := TagDecode(bytes)
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

func SignedNumberDecoder(field *Field, bytes *bytes.Buffer) (int64, error) {
	switch field.Tags.Protobuf.WireType {
	case WireTypeI32:
		{
			value, err := Fixed32Decode(bytes)
			if err != nil {
				return 0, err
			}
			return int64(value), nil
		}
	case WireTypeI64:
		{
			return Fixed64Decode(bytes)
		}
	default:
		{
			return ZigzagDecode(bytes)
		}
	}
}

func signedNumberDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	switch wireType {
	case WireTypeI32:
		{
			value, err := Fixed32Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(int64(value))
			return nil
		}
	case WireTypeI64:
		{
			value, err := Fixed64Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(value)
			return nil
		}
	default:
		{
			value, err := ZigzagDecode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(int64(value))
			return nil
		}
	}
}

func UnsignedNumberDecoder(field *Field, bytes *bytes.Buffer) (uint64, error) {
	switch field.Tags.Protobuf.WireType {
	case WireTypeI32:
		{
			value, err := Fixed32Decode(bytes)
			if err != nil {
				return 0, err
			}
			return uint64(value), nil
		}
	case WireTypeI64:
		{
			value, err := Fixed64Decode(bytes)
			if err != nil {
				return 0, err
			}
			return uint64(value), nil
		}
	default:
		{
			return UvarintDecode(bytes)
		}
	}
}

func unsignedNumberDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	switch wireType {
	case WireTypeI32:
		{
			value, err := Fixed32Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetInt(int64(value))
			return nil
		}
	case WireTypeI64:
		{
			value, err := Fixed64Decode(bytes)
			if err != nil {
				return err
			}
			elem.SetUint(uint64(value))
			return nil
		}
	default:
		{
			value, err := UvarintDecode(bytes)
			if err != nil {
				return err
			}
			elem.SetUint(value)
			return nil
		}
	}
}

func FloatDecoder(field *Field, bytes *bytes.Buffer) (float32, error) {
	return Float32Decode(bytes)
}

func floatDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := Float32Decode(bytes)
	if err != nil {
		return err
	}
	elem.SetFloat(float64(value))
	return nil
}

func DoubleDecoder(field *Field, bytes *bytes.Buffer) (float64, error) {
	return Float64Decode(bytes)
}

func doubleDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := Float64Decode(bytes)
	if err != nil {
		return err
	}
	elem.SetFloat(value)
	return nil
}

func BooleanDecoder(field *Field, bytes *bytes.Buffer) (bool, error) {
	return BoolDecode(bytes)
}

func booleanDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := BoolDecode(bytes)
	if err != nil {
		return err
	}
	elem.SetBool(value)
	return nil
}

func StringDecoder(field *Field, bytes *bytes.Buffer) (string, error) {
	return StringDecode(bytes)
}

func stringDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := StringDecode(bytes)
	if err != nil {
		return err
	}
	elem.SetString(value)
	return nil
}

func arrayDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	k := field.Index
	if k == reflect.Uint8 {
		value, err := BytesDecode(bytes)
		if err != nil {
			return err
		}
		v.SetBytes(value)
		return nil
	}
	tmp := reflect.New(v.Type().Elem())
	tmp = tmp.Elem()
	switch wireType {
	case WireTypeVarint, WireTypeI32, WireTypeI64:
		{
			value, err := BytesDecode(bytes)
			if err != nil {
				return err
			}
			buffer := Alloc(0)
			defer Dealloc(buffer)
			_, _ = buffer.Write(value)
			for buffer.Len() != 0 {
				elem, addr := dereference(tmp)
				err := _decoders[elem.Kind()](elem, nil, buffer, wireType)
				if err != nil {
					return err
				}
				v.Set(reflect.Append(v, addr))
			}
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
	value, err := BytesDecode(bytes)
	if err != nil {
		return err
	}

	buffer := Alloc(0)
	defer Dealloc(buffer)
	_, _ = buffer.Write(value)

	typ := v.Type()
	keyType := typ.Key()
	valueType := typ.Elem()
	if v.IsZero() {
		v.Set(reflect.MakeMap(reflect.MapOf(keyType, valueType)))
	}
	_, keyWireType, err := TagDecode(buffer)
	if err != nil {
		return err
	}
	key := reflect.New(keyType).Elem()
	err = _decoders[key.Kind()](key, nil, buffer, keyWireType)
	if err != nil {
		return err
	}

	_, valueWireType, err := TagDecode(buffer)
	if err != nil {
		return err
	}
	val := reflect.New(valueType).Elem()
	elem, addr := dereference(val)
	err = _decoders[elem.Kind()](elem, nil, buffer, valueWireType)
	if err != nil {
		return err
	}
	v.SetMapIndex(key, addr)
	return nil
}

func structDecoder(v reflect.Value, field *Field, bytes *bytes.Buffer, wireType WireType) error {
	elem, _ := dereference(v)
	value, err := BytesDecode(bytes)
	if err != nil {
		return err
	}
	buffer := Alloc(0)
	defer Dealloc(buffer)
	_, _ = buffer.Write(value)
	if err := unmarshal(buffer, elem); err != nil {
		return err
	}
	return nil
}
