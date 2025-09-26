package protolizer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func encodeFixed32(value int32) *bytes.Buffer {
	memory := alloc(4)
	buf := memory.AvailableBuffer()[:4]
	binary.LittleEndian.PutUint32(buf, uint32(value))
	memory.Write(buf)
	return memory
}

func encodeFixed64(value int64) *bytes.Buffer {
	memory := alloc(8)
	buf := memory.AvailableBuffer()[:8]
	binary.LittleEndian.PutUint64(buf, uint64(value))
	memory.Write(buf)
	return memory
}

func decodeFixed32(data []byte, offset int) (int32, int, error) {
	if len(data) < offset+4 {
		return 0, 0, fmt.Errorf("insufficient bytes for fixed32")
	}
	value := binary.LittleEndian.Uint32(data[offset : offset+4])
	return int32(value), 4, nil
}

func decodeFixed64(data []byte, offset int) (int64, int, error) {
	if len(data) < offset+8 {
		return 0, 0, fmt.Errorf("insufficient bytes for fixed64")
	}
	value := binary.LittleEndian.Uint64(data[offset : offset+8])
	return int64(value), 8, nil
}
