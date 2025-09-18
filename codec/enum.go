package main

func EncodeEnum(value int32) []byte {
	return EncodeVarint(int64(value))
}

func DecodeEnum(data []byte, offset int) (int32, int, error) {
	value, consumed, err := DecodeVarint(data, offset)
	if err != nil {
		return 0, 0, err
	}
	return int32(value), consumed, nil
}

func EnumWireType() WireType {
	return WireTypeVarint
}
