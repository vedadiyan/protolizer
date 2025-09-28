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

func bytesDecode(data *bytes.Buffer) ([]byte, error) {
	length, err := varintDecode(data)
	if err != nil {
		return nil, err
	}

	if length < 0 {
		return nil, fmt.Errorf("negative length")
	}

	if data.Len() < int(length) {
		return nil, fmt.Errorf("insufficient bytes for length-prefixed data")
	}
	return data.Next(int(length)), nil
}

func stringDecode(data *bytes.Buffer) (string, error) {
	bytes, err := bytesDecode(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
