package protolizer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func Float32Encode(value float32) *bytes.Buffer {
	memory := Alloc(4)
	buf := memory.AvailableBuffer()[:4]
	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(buf, bits)
	memory.Write(buf)
	return memory
}

func Float32InlineEncode(value float32, buffer *bytes.Buffer) {
	buffer.Grow(4)
	buf := buffer.AvailableBuffer()[:4]
	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(buf, bits)
	buffer.Write(buf)
}

func Float64Encode(value float64) *bytes.Buffer {
	memory := Alloc(8)
	buf := memory.AvailableBuffer()[:8]
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(buf, bits)
	memory.Write(buf)
	return memory
}

func Float64InlineEncode(value float64, buffer *bytes.Buffer) {
	buffer.Grow(8)
	buf := buffer.AvailableBuffer()[:8]
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(buf, bits)
	buffer.Write(buf)
}

func Float32Decode(data *bytes.Buffer) (float32, error) {
	if data.Len() < 4 {
		return 0, fmt.Errorf("insufficient bytes for float32")
	}
	bits := binary.LittleEndian.Uint32(data.Next(4))
	value := math.Float32frombits(bits)
	return value, nil
}

func Float64Decode(data *bytes.Buffer) (float64, error) {
	if data.Len() < 8 {
		return 0, fmt.Errorf("insufficient bytes for float64")
	}
	bits := binary.LittleEndian.Uint64(data.Next(8))
	value := math.Float64frombits(bits)
	return value, nil
}
