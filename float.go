package protolizer

import (
	"encoding/binary"
	"fmt"
	"math"
)

func encodeFloat32(value float32) []byte {
	buf := make([]byte, 4)
	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(buf, bits)
	return buf
}

func encodeFloat64(value float64) []byte {
	buf := make([]byte, 8)
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(buf, bits)
	return buf
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
