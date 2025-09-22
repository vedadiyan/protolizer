package codec

import "fmt"

func encodeTag(fieldNumber int32, wireType WireType) ([]byte, error) {
	if fieldNumber < 1 {
		return nil, fmt.Errorf("field number must be positive")
	}
	if wireType > 5 {
		return nil, fmt.Errorf("invalid wire type")
	}

	tag := (int64(fieldNumber) << 3) | int64(wireType)
	return encodeVarint(tag), nil
}

func decodeTag(data []byte, offset int) (int32, WireType, int, error) {
	tag, consumed, err := decodeVarint(data, offset)
	if err != nil {
		return 0, 0, 0, err
	}

	fieldNumber := int32(tag >> 3)
	wireType := WireType(tag & 0x7)

	if fieldNumber < 1 {
		return 0, 0, 0, fmt.Errorf("invalid field number")
	}

	return fieldNumber, wireType, consumed, nil
}
