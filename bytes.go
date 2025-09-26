package protolizer

import (
	"bytes"
	"fmt"
)

func encodeBytes(value []byte) *bytes.Buffer {
	memory := alloc(0)
	length := encodeVarint(int64(len(value)))
	length.WriteTo(memory)
	dealloc(length)
	memory.Write(value)
	return memory
}

func encodeString(value string) *bytes.Buffer {
	return encodeBytes([]byte(value))
}

func decodeBytes(data []byte, offset int) ([]byte, int, error) {
	length, lengthSize, err := decodeVarint(data, offset)
	if err != nil {
		return nil, 0, err
	}

	if length < 0 {
		return nil, 0, fmt.Errorf("negative length")
	}

	start := offset + lengthSize
	end := start + int(length)

	if len(data) < end {
		return nil, 0, fmt.Errorf("insufficient bytes for length-prefixed data")
	}

	value := make([]byte, length)
	copy(value, data[start:end])
	return value, lengthSize + int(length), nil
}

func decodeString(data []byte, offset int) (string, int, error) {
	bytes, consumed, err := decodeBytes(data, offset)
	if err != nil {
		return "", 0, err
	}
	return string(bytes), consumed, nil
}
