package main

func EncodeField(n int32, wireType WireType, v []byte) ([]byte, error) {
	tagBytes, err := EncodeTag(n, wireType)
	if err != nil {
		return nil, err
	}
	return append(tagBytes, v...), nil
}
