package protolizer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func encodeFloat32(value float32) []byte {
	buf := _buffer.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		_buffer.Put(buf)
	}()
	buf.Grow(4)
	out := buf.AvailableBuffer()[:4]
	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(out, bits)
	return out
}

func encodeFloat64(value float64) []byte {
	buf := _buffer.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		_buffer.Put(buf)
	}()
	buf.Grow(8)
	out := buf.AvailableBuffer()[:8]
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(out, bits)
	return out
}

func decodeFloat32(data []byte, offset int) (float32, int, error) {
	if len(data) < offset+4 {
		return 0, 0, fmt.Errorf("insufficient bytes for float32")
	}
	bits := binary.LittleEndian.Uint32(data[offset : offset+4])
	value := math.Float32frombits(bits)
	return value, 4, nil
}

func decodeFloat64(data []byte, offset int) (float64, int, error) {
	if len(data) < offset+8 {
		return 0, 0, fmt.Errorf("insufficient bytes for float64")
	}
	bits := binary.LittleEndian.Uint64(data[offset : offset+8])
	value := math.Float64frombits(bits)
	return value, 8, nil
}
