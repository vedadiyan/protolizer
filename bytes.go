package protolizer

import (
	"bytes"
	"fmt"
)

func encodeBytes(value []byte) []byte {
	out := _buffer.Get().(*bytes.Buffer)
	defer func() {
		out.Reset()
		_buffer.Put(out)
	}()
	length := encodeVarint(int64(len(value)))
	out.Write(length)
	out.Write(value)
	return out.Bytes()
}

func encodeString(value string) []byte {
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

	return data[start:end], lengthSize + int(length), nil
}

func decodeString(data []byte, offset int) (string, int, error) {
	bytes, consumed, err := decodeBytes(data, offset)
	if err != nil {
		return "", 0, err
	}
	return string(bytes), consumed, nil
}
