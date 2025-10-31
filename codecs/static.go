package codecs

import (
	"bytes"

	p "github.com/vedadiyan/protolizer"
)

type (
	Reflected interface {
		Encode(*p.Field, *bytes.Buffer) error
		Decode(*p.Field, *bytes.Buffer) error
		New() Reflected
		Type() p.Type
		IsZero(*p.Field) bool
	}
	Static struct{}
)

func (*Static) Marshal(v Reflected) ([]byte, error) {
	typ := v.Type()

	buffer := p.Alloc(0)
	defer p.Dealloc(buffer)
	for _, field := range typ.Fields {
		if v.IsZero(field) {
			continue
		}
		err := p.TagInlineEncode(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType, buffer)
		if err != nil {
			return nil, err
		}
		if err := v.Encode(field, buffer); err != nil {
			return nil, err
		}
	}

	return bytes.Clone(buffer.Bytes()), nil
}

func (*Static) InlineMarshal(v Reflected) (*bytes.Buffer, error) {
	typ := v.Type()

	buffer := p.Alloc(0)
	for _, field := range typ.Fields {
		if v.IsZero(field) {
			continue
		}
		err := p.TagInlineEncode(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType, buffer)
		if err != nil {
			return nil, err
		}
		if err := v.Encode(field, buffer); err != nil {
			return nil, err
		}
	}

	return buffer, nil
}

func (*Static) Unmarshal(data []byte, v Reflected) error {
	typ := v.Type()

	buffer := p.Alloc(0)
	defer p.Dealloc(buffer)
	buffer.Write(data)

	for buffer.Len() != 0 {
		fieldNumber, _, err := p.TagDecode(buffer)
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

func (*Static) UnmarshalFromBuffer(v Reflected, data *bytes.Buffer) error {
	typ := v.Type()
	l, err := p.UvarintDecode(data)
	if err != nil {
		return err
	}
	end := data.Len() - int(l)
	for data.Len() != end {
		fieldNumber, _, err := p.TagDecode(data)
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
