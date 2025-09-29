package protolizer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func Fixed32Encode(value int32) *bytes.Buffer {
	memory := Alloc(4)
	buf := memory.AvailableBuffer()[:4]
	binary.LittleEndian.PutUint32(buf, uint32(value))
	memory.Write(buf)
	return memory
}

func Fixed64Encode(value int64) *bytes.Buffer {
	memory := Alloc(8)
	buf := memory.AvailableBuffer()[:8]
	binary.LittleEndian.PutUint64(buf, uint64(value))
	memory.Write(buf)
	return memory
}

func Fixed32Decode(data *bytes.Buffer) (int32, error) {
	if data.Len() < 4 {
		return 0, fmt.Errorf("insufficient bytes for fixed32")
	}
	value := binary.LittleEndian.Uint32(data.Next(4))
	return int32(value), nil
}

func Fixed64Decode(data *bytes.Buffer) (int64, error) {
	if data.Len() < 8 {
		return 0, fmt.Errorf("insufficient bytes for fixed64")
	}
	value := binary.LittleEndian.Uint64(data.Next(8))
	return int64(value), nil
}
