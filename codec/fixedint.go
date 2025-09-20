package codec

import (
	"encoding/binary"
	"fmt"
)

func EncodeFixed32(value int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(value))
	return buf
}

func EncodeFixed64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

func DecodeFixed32(data []byte, offset int) (int32, int, error) {
	if len(data) < offset+4 {
		return 0, 0, fmt.Errorf("insufficient bytes for fixed32")
	}
	value := binary.LittleEndian.Uint32(data[offset : offset+4])
	return int32(value), 4, nil
}

func DecodeFixed64(data []byte, offset int) (int64, int, error) {
	if len(data) < offset+8 {
		return 0, 0, fmt.Errorf("insufficient bytes for fixed64")
	}
	value := binary.LittleEndian.Uint64(data[offset : offset+8])
	return int64(value), 8, nil
}

func Fixed32WireType() WireType {
	return WireTypeI32
}

func Fixed64WireType() WireType {
	return WireTypeI64
}
