package protolizer

import (
	"bytes"
	"fmt"
)

func VarintEncode(value int64) *bytes.Buffer {
	return UvarintEncode(uint64(value))
}

func UvarintEncode(value uint64) *bytes.Buffer {
	memory := Alloc(0)
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

func VarintDecode(data *bytes.Buffer) (int64, error) {
	value, err := UvarintDecode(data)
	return int64(value), err
}

func UvarintDecode(data *bytes.Buffer) (uint64, error) {
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

func UvarintPeek(data *bytes.Buffer) (uint64, error) {
	var result uint64
	var shift uint
	bytes := data.Bytes()

	for i := 0; i < len(bytes); i++ {
		b := bytes[i]
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
