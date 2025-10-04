package protolizer

import (
	"bytes"
	"fmt"
)

func BytesEncode(value []byte) *bytes.Buffer {
	memory := Alloc(0)
	uvarint(uint64(len(value)), memory)
	memory.Write(value)
	return memory
}

func BufferEncode(value *bytes.Buffer) *bytes.Buffer {
	memory := Alloc(0)
	uvarint(uint64(value.Len()), memory)
	value.WriteTo(memory)
	return memory
}

func StringEncode(value string) *bytes.Buffer {
	return BytesEncode([]byte(value))
}

func BytesDecode(data *bytes.Buffer) ([]byte, error) {
	length, err := VarintDecode(data)
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

func StringDecode(data *bytes.Buffer) (string, error) {
	bytes, err := BytesDecode(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
