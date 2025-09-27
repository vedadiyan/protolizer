package protolizer

import "bytes"

func boolEncode(value bool) *bytes.Buffer {
	if value {
		return varintEncode(1)
	}
	return varintEncode(0)
}

func boolDecode(data *bytes.Buffer, offset int) (bool, int, error) {
	value, consumed, err := varintDecode(data, offset)
	if err != nil {
		return false, 0, err
	}
	return value != 0, consumed, nil
}
