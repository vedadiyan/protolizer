package codecs

import (
	"bytes"

	aloc "github.com/vedadiyan/protolizer/memory"
	"github.com/vedadiyan/protolizer/metadata"
	"github.com/vedadiyan/protolizer/pdk"
)

type (
	Reflected interface {
		Encode(metadata.Field, *bytes.Buffer) error
		Decode(metadata.Field, *bytes.Buffer) error
		New() Reflected
		Type() metadata.Type
		IsZero(metadata.Field) bool
	}
	Static struct{}
)

func (*Static) Marshal(v Reflected) ([]byte, error) {
	typ := v.Type()

	buffer := aloc.Alloc(0)
	defer aloc.Dealloc(buffer)
	for _, field := range typ.Fields {
		if v.IsZero(field) {
			continue
		}
		err := pdk.TagInlineEncode(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType, buffer)
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

	buffer := aloc.Alloc(0)
	for _, field := range typ.Fields {
		if v.IsZero(field) {
			continue
		}
		err := pdk.TagInlineEncode(int32(field.Tags.Protobuf.FieldNum), field.Tags.Protobuf.WireType, buffer)
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

	buffer := aloc.Alloc(0)
	defer aloc.Dealloc(buffer)
	buffer.Write(data)

	for buffer.Len() != 0 {
		fieldNumber, _, err := pdk.TagDecode(buffer)
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
	l, err := pdk.UvarintDecode(data)
	if err != nil {
		return err
	}
	end := data.Len() - int(l)
	for data.Len() != end {
		fieldNumber, _, err := pdk.TagDecode(data)
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
