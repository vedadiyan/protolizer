package protolizer

import "bytes"

func boolEncode(value bool) *bytes.Buffer {
	if value {
		return varintEncode(1)
	}
	return varintEncode(0)
}

func boolDecode(data *bytes.Buffer) (bool, error) {
	value, err := varintDecode(data)
	if err != nil {
		return false, err
	}
	return value != 0, nil
}
