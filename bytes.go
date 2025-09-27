package protolizer

import (
	"bytes"
	"fmt"
)

func bytesEncode(value []byte) *bytes.Buffer {
	memory := alloc(0)
	uvarint(uint64(len(value)), memory)
	memory.Write(value)
	return memory
}

func stringEncode(value string) *bytes.Buffer {
	return bytesEncode([]byte(value))
}

func bytesDecode(data *bytes.Buffer, offset int) (*bytes.Buffer, int, error) {
	length, lengthSize, err := varintDecode(data, offset)
	if err != nil {
		return nil, 0, err
	}

	if length < 0 {
		return nil, 0, fmt.Errorf("negative length")
	}

	start := offset + lengthSize
	end := start + int(length)

	if data.Len() < end {
		return nil, 0, fmt.Errorf("insufficient bytes for length-prefixed data")
	}
	buffer := alloc(0)
	buffer.Write(data.Bytes()[start:end])
	return buffer, lengthSize + int(length), nil
}

func stringDecode(data *bytes.Buffer, offset int) (string, int, error) {
	bytes, consumed, err := bytesDecode(data, offset)
	defer dealloc(bytes)
	if err != nil {
		return "", 0, err
	}
	return bytes.String(), consumed, nil
}
