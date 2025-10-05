package protolizer

import "bytes"

func ZigzagEncode(value int64) *bytes.Buffer {
	encoded := uint64((value << 1) ^ (value >> 63))
	return UvarintEncode(encoded)
}

func ZigzagInlineEncode(value int64, buffer *bytes.Buffer) {
	encoded := uint64((value << 1) ^ (value >> 63))
	UvarintInlineEncode(encoded, buffer)
}

func ZigzagDecode(data *bytes.Buffer) (int64, error) {
	encoded, err := UvarintDecode(data)
	if err != nil {
		return 0, err
	}
	value := int64((encoded >> 1) ^ uint64((int64(encoded&1)<<63)>>63))
	return value, nil
}
