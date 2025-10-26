package protolizer

import (
	"bytes"
	"fmt"
	"io"
)

func TagEncode(fieldNumber int32, wireType WireType) (*bytes.Buffer, error) {
	if fieldNumber < 1 {
		return nil, fmt.Errorf("field number must be positive")
	}
	if wireType > 5 {
		return nil, fmt.Errorf("invalid wire type")
	}

	tag := (int64(fieldNumber) << 3) | int64(wireType)
	return UvarintEncode(uint64(tag)), nil
}

func TagInlineEncode(fieldNumber int32, wireType WireType, buffer *bytes.Buffer) error {
	if fieldNumber < 1 {
		return fmt.Errorf("field number must be positive")
	}
	if wireType > 5 {
		return fmt.Errorf("invalid wire type")
	}

	tag := (int64(fieldNumber) << 3) | int64(wireType)
	UvarintInlineEncode(uint64(tag), buffer)
	return nil
}

func TagDecode(data *bytes.Buffer) (int32, WireType, error) {
	tag, err := UvarintDecode(data)
	if err != nil {
		return 0, 0, err
	}

	fieldNumber := int32(tag >> 3)
	wireType := WireType(tag & 0x7)

	if fieldNumber < 1 {
		return 0, 0, fmt.Errorf("invalid field number")
	}

	return fieldNumber, wireType, nil
}

func TagPeek(data *bytes.Buffer) (int32, WireType, func(), error) {
	if data.Len() == 0 {
		return 0, 0, nil, io.EOF
	}
	tag, err := UvarintPeek(data)
	if err != nil {
		return 0, 0, nil, err
	}

	fieldNumber := int32(tag >> 3)
	wireType := WireType(tag & 0x7)

	if fieldNumber < 1 {
		return 0, 0, nil, fmt.Errorf("invalid field number")
	}

	return fieldNumber, wireType, func() { TagDecode(data) }, nil
}
