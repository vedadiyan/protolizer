package protolizer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func float32Encode(value float32) *bytes.Buffer {
	memory := alloc(4)
	buf := memory.AvailableBuffer()[:4]
	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(buf, bits)
	memory.Write(buf)
	return memory
}

func float46Encode(value float64) *bytes.Buffer {
	memory := alloc(8)
	buf := memory.AvailableBuffer()[:8]
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(buf, bits)
	memory.Write(buf)
	return memory
}

func float32Decode(data *bytes.Buffer, offset int) (float32, int, error) {
	if data.Len() < offset+4 {
		return 0, 0, fmt.Errorf("insufficient bytes for float32")
	}
	bits := binary.LittleEndian.Uint32(data.Bytes()[offset : offset+4])
	value := math.Float32frombits(bits)
	return value, 4, nil
}

func float64Decode(data *bytes.Buffer, offset int) (float64, int, error) {
	if data.Len() < offset+8 {
		return 0, 0, fmt.Errorf("insufficient bytes for float64")
	}
	bits := binary.LittleEndian.Uint64(data.Bytes()[offset : offset+8])
	value := math.Float64frombits(bits)
	return value, 8, nil
}
