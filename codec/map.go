package codec

import "fmt"

func EncodeStringInt64Map(mapData map[string]int64) ([][]byte, error) {
	var entries [][]byte

	for key, value := range mapData {
		keyBytes, err := EncodeField(1, WireTypeLen, EncodeString(key))
		if err != nil {
			return nil, err
		}
		valueBytes, err := EncodeField(2, WireTypeVarint, EncodeVarint(value))
		if err != nil {
			return nil, err
		}
		entryMessage := append(keyBytes, valueBytes...)
		entries = append(entries, EncodeMessage(entryMessage))
	}

	return entries, nil
}

func DecodeStringInt64MapEntry(data []byte, offset int) (string, int64, int, error) {
	entryData, consumed, err := DecodeMessage(data, offset)
	if err != nil {
		return "", 0, 0, err
	}

	pos := 0
	var key string
	var value int64

	for pos < len(entryData) {
		fieldNum, wireType, headerSize, err := DecodeTag(entryData, pos)
		if err != nil {
			return "", 0, 0, err
		}
		pos += headerSize

		switch fieldNum {
		case 1:
			if wireType != WireTypeLen {
				return "", 0, 0, fmt.Errorf("map key must be string")
			}
			key, headerSize, err = DecodeString(entryData, pos)
			if err != nil {
				return "", 0, 0, err
			}
			pos += headerSize

		case 2:
			if wireType != WireTypeVarint {
				return "", 0, 0, fmt.Errorf("map value must be varint")
			}
			value, headerSize, err = DecodeVarint(entryData, pos)
			if err != nil {
				return "", 0, 0, err
			}
			pos += headerSize
		}
	}

	return key, value, consumed, nil
}
