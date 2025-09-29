package protolizer

import "bytes"

type (
	Reflected interface {
		Encode(*Field, *bytes.Buffer) error
		Decode(*Field, *bytes.Buffer) error
		New() Reflected
		Type() Type
	}
)

func FastMarshal(v Reflected) ([]byte, error) {
	typ := v.Type()

	buffer := alloc(0)
	defer dealloc(buffer)
	for _, fields := range typ.Fields {
		if err := v.Encode(fields, buffer); err != nil {
			return nil, err
		}
	}

	return bytes.Clone(buffer.Bytes()), nil
}

func FastUnmarshal(v Reflected, date []byte) error {
	typ := v.Type()

	buffer := alloc(0)
	defer dealloc(buffer)
	buffer.Write(date)
	for _, fields := range typ.Fields {
		if err := v.Decode(fields, buffer); err != nil {
			return err
		}
	}

	return nil
}
