package protolizer

import "bytes"

func BoolEncode(value bool) *bytes.Buffer {
	if value {
		return VarintEncode(1)
	}
	return VarintEncode(0)
}

func BoolInlineEncode(value bool, buffer *bytes.Buffer) {
	if value {
		UvarintInlineEncode(1, buffer)
		return
	}
	UvarintInlineEncode(0, buffer)
}

func BoolDecode(data *bytes.Buffer) (bool, error) {
	value, err := VarintDecode(data)
	if err != nil {
		return false, err
	}
	return value != 0, nil
}
