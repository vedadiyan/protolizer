package codec

func encodeBool(value bool) []byte {
	if value {
		return encodeVarint(1)
	}
	return encodeVarint(0)
}

func decodeBool(data []byte, offset int) (bool, int, error) {
	value, consumed, err := decodeVarint(data, offset)
	if err != nil {
		return false, 0, err
	}
	return value != 0, consumed, nil
}
