package main

func EncodeZigzag(value int64) []byte {
	encoded := ZigzagEncode(value)
	return EncodeUvarint(encoded)
}

func DecodeZigzag(data []byte, offset int) (int64, int, error) {
	encoded, consumed, err := DecodeUvarint(data, offset)
	if err != nil {
		return 0, 0, err
	}
	value := ZigzagDecode(encoded)
	return value, consumed, nil
}

func ZigzagEncode(value int64) uint64 {
	return uint64((value << 1) ^ (value >> 63))
}

func ZigzagDecode(value uint64) int64 {
	return int64((value >> 1) ^ uint64((int64(value&1)<<63)>>63))
}
