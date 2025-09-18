package main

func EncodeBool(value bool) []byte {
	if value {
		return EncodeVarint(1)
	}
	return EncodeVarint(0)
}

func DecodeBool(data []byte, offset int) (bool, int, error) {
	value, consumed, err := DecodeVarint(data, offset)
	if err != nil {
		return false, 0, err
	}
	return value != 0, consumed, nil
}
