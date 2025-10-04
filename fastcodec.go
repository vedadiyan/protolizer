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
		err := TagEncodeToBuffer(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType, buffer)
		if err != nil {
			return nil, err
		}
		if err := v.Encode(field, buffer); err != nil {
			return nil, err
		}
	}

	return bytes.Clone(buffer.Bytes()), nil
}

func FastMarshalToBuffer(v Reflected) (*bytes.Buffer, error) {
	typ := v.Type()

	buffer := Alloc(0)
	for _, field := range typ.Fields {
		if v.IsZero(field) {
			continue
		}
		err := TagEncodeToBuffer(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType, buffer)
		if err != nil {
			return nil, err
		}
		if err := v.Encode(field, buffer); err != nil {
			return nil, err
		}
	}

	return buffer, nil
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

func FastUnmarshalFromBuffer(v Reflected, data *bytes.Buffer) error {
	typ := v.Type()
	l, err := UvarintDecode(data)
	if err != nil {
		return err
	}
	end := data.Len() - int(l)
	for data.Len() != end {
		fieldNumber, _, err := TagDecode(data)
		if err != nil {
			return err
		}
		field := typ.FieldsIndexer[int(fieldNumber)]
		if err := v.Decode(field, data); err != nil {
			return err
		}
	}
	return nil
}
