package codec

func EncodeMessage(content []byte) []byte {
	return EncodeBytes(content)
}

func DecodeMessage(data []byte, offset int) ([]byte, int, error) {
	return DecodeBytes(data, offset)
}
