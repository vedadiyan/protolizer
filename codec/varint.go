package codec

import "fmt"

func EncodeVarint(value int64) []byte {
	return EncodeUvarint(uint64(value))
}

func EncodeUvarint(value uint64) []byte {
	var result []byte
	for value >= 0x80 {
		result = append(result, byte(value)|0x80)
		value >>= 7
	}
	result = append(result, byte(value))
	return result
}

func EncodeVarintLongForm(value int64, extraBytes int) []byte {
	if extraBytes <= 0 {
		return EncodeVarint(value)
	}

	normal := EncodeVarint(value)
	if len(normal) == 0 {
		return normal
	}

	result := make([]byte, len(normal)-1, len(normal)+extraBytes)
	copy(result, normal[:len(normal)-1])

	for i := range result {
		result[i] |= 0x80
	}

	for range extraBytes {
		result = append(result, 0x80)
	}

	result = append(result, normal[len(normal)-1])
	return result
}

func DecodeVarint(data []byte, offset int) (int64, int, error) {
	value, consumed, err := DecodeUvarint(data, offset)
	return int64(value), consumed, err
}

func DecodeUvarint(data []byte, offset int) (uint64, int, error) {
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

func VarintWireType() WireType {
	return WireTypeVarint
}
