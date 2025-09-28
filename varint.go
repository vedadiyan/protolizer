package protolizer

import (
	"bytes"
	"fmt"
)

func varintEncode(value int64) *bytes.Buffer {
	return uvarintEncode(uint64(value))
}

func uvarintEncode(value uint64) *bytes.Buffer {
	memory := alloc(0)
	uvarint(value, memory)
	return memory
}

func uvarint(value uint64, buffer *bytes.Buffer) {
	for value >= 0x80 {
		buffer.WriteByte(byte(value) | 0x80)
		value >>= 7
	}
	buffer.WriteByte(byte(value))
}

func varintDecode(data *bytes.Buffer) (int64, error) {
	value, err := uvarintDecode(data)
	return int64(value), err
}

func uvarintDecode(data *bytes.Buffer) (uint64, error) {
	var result uint64
	var shift uint

	for data.Len() != 0 {
		b, _ := data.ReadByte()
		if shift == 63 && b > 1 {
			return 0, fmt.Errorf("varint overflows uint64")
		}
		result |= uint64(b&0x7f) << shift

		if b&0x80 == 0 {
			return result, nil
		}

		shift += 7
	}
	return 0, fmt.Errorf("truncated varint")
}
