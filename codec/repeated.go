package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

func EncodeRepeatedVarint(values []int64) [][]byte {
	var result [][]byte
	for _, value := range values {
		result = append(result, EncodeVarint(value))
	}
	return result
}

func EncodePackedVarint(values []int64) []byte {
	var packed []byte
	for _, value := range values {
		packed = append(packed, EncodeVarint(value)...)
	}
	return EncodeBytes(packed)
}

func EncodePackedFixed32(values []int32) []byte {
	packed := make([]byte, 0, len(values)*4)
	for _, value := range values {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(value))
		packed = append(packed, buf...)
	}
	return EncodeBytes(packed)
}

func EncodePackedFixed64(values []int64) []byte {
	packed := make([]byte, 0, len(values)*8)
	for _, value := range values {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(value))
		packed = append(packed, buf...)
	}
	return EncodeBytes(packed)
}

func EncodePackedFloat32(values []float32) []byte {
	packed := make([]byte, 0, len(values)*4)
	for _, value := range values {
		buf := make([]byte, 4)
		bits := math.Float32bits(value)
		binary.LittleEndian.PutUint32(buf, bits)
		packed = append(packed, buf...)
	}
	return EncodeBytes(packed)
}

func EncodePackedFloat64(values []float64) []byte {
	packed := make([]byte, 0, len(values)*8)
	for _, value := range values {
		buf := make([]byte, 8)
		bits := math.Float64bits(value)
		binary.LittleEndian.PutUint64(buf, bits)
		packed = append(packed, buf...)
	}
	return EncodeBytes(packed)
}

func DecodePackedVarint(data []byte, offset int) ([]int64, int, error) {
	packedData, consumed, err := DecodeBytes(data, offset)
	if err != nil {
		return nil, 0, err
	}

	var values []int64
	pos := 0
	for pos < len(packedData) {
		value, size, err := DecodeVarint(packedData, pos)
		if err != nil {
			return nil, 0, err
		}
		values = append(values, value)
		pos += size
	}

	return values, consumed, nil
}

func DecodePackedFixed32(data []byte, offset int) ([]int32, int, error) {
	packedData, consumed, err := DecodeBytes(data, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(packedData)%4 != 0 {
		return nil, 0, fmt.Errorf("packed fixed32 data length not multiple of 4")
	}

	var values []int32
	for i := 0; i < len(packedData); i += 4 {
		value := binary.LittleEndian.Uint32(packedData[i : i+4])
		values = append(values, int32(value))
	}

	return values, consumed, nil
}

func RepeatedWireType() WireType {
	return WireTypeVarint
}

func RepeatedPackedWireType() WireType {
	return WireTypeLen
}
