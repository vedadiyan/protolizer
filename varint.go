package protolizer

import (
	"bytes"
	"fmt"
)

func encodeVarint(value int64) *bytes.Buffer {
	return encodeUvarint(uint64(value))
}

func encodeUvarint(value uint64) *bytes.Buffer {
	memory := alloc(8)
	for value >= 0x80 {
		memory.WriteByte(byte(value) | 0x80)
		value >>= 7
	}
	memory.WriteByte(byte(value))
	return memory
}

func decodeVarint(data []byte, offset int) (int64, int, error) {
	value, consumed, err := decodeUvarint(data, offset)
	return int64(value), consumed, err
}

func decodeUvarint(data []byte, offset int) (uint64, int, error) {
	var result uint64
	var shift uint
	pos := offset

	for pos < len(data) {
		b := data[pos]
		if shift == 63 && b > 1 {
			return 0, 0, fmt.Errorf("varint overflows uint64")
		}
		result |= uint64(b&0x7f) << shift
		pos++

		if b&0x80 == 0 {
			return result, pos - offset, nil
		}

		shift += 7
	}
	return 0, 0, fmt.Errorf("truncated varint")
}
