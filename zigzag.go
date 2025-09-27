package protolizer

import "bytes"

func zigzagEncode(value int64) *bytes.Buffer {
	encoded := uint64((value << 1) ^ (value >> 63))
	return uvarintEncode(encoded)
}

func zigzagDecode(data *bytes.Buffer, offset int) (int64, int, error) {
	encoded, consumed, err := uvarintDecode(data, offset)
	if err != nil {
		return 0, 0, err
	}
	value := int64((encoded >> 1) ^ uint64((int64(encoded&1)<<63)>>63))
	return value, consumed, nil
}
