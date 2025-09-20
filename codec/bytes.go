package codec

import "fmt"

func EncodeBytes(value []byte) []byte {
	length := EncodeVarint(int64(len(value)))
	return append(length, value...)
}

func EncodeBytesLongForm(value []byte, extraBytes int) []byte {
	length := EncodeVarintLongForm(int64(len(value)), extraBytes)
	return append(length, value...)
}

func EncodeString(value string) []byte {
	return EncodeBytes([]byte(value))
}

func DecodeBytes(data []byte, offset int) ([]byte, int, error) {
	length, lengthSize, err := DecodeVarint(data, offset)
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

func DecodeString(data []byte, offset int) (string, int, error) {
	bytes, consumed, err := DecodeBytes(data, offset)
	if err != nil {
		return "", 0, err
	}
	return string(bytes), consumed, nil
}

func BytesWireType() WireType {
	return WireTypeLen
}

func StringWireType() WireType {
	return WireTypeLen
}
