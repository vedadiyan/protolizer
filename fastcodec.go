package protolizer

import "bytes"

type (
	Reflected interface {
		Encode(*Field, *bytes.Buffer) error
		Decode(*Field, *bytes.Buffer) error
		New() Reflected
		Type() Type
		IsZero(*Field) bool
	}
)

func FastMarshal(v Reflected) ([]byte, error) {
	typ := v.Type()

	buffer := Alloc(0)
	defer Dealloc(buffer)
	for _, field := range typ.Fields {
		if v.IsZero(field) {
			continue
		}
		tag, err := TagEncode(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType)
		if err != nil {
			return nil, err
		}
		tag.WriteTo(buffer)
		Dealloc(tag)
		if err := v.Encode(field, buffer); err != nil {
			return nil, err
		}
	}

	return bytes.Clone(buffer.Bytes()), nil
}

func FastUnmarshal(v Reflected, data []byte) error {
	typ := v.Type()

	buffer := Alloc(0)
	defer Dealloc(buffer)
	buffer.Write(data)

	for buffer.Len() != 0 {
		fieldNumber, _, err := TagDecode(buffer)
		if err != nil {
			return err
		}
		field := typ.FieldsIndexer[int(fieldNumber)]
		if err := v.Decode(field, buffer); err != nil {
			return err
		}
	}
	return nil
}
