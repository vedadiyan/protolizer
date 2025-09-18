// Package protoscope provides direct encoding and decoding for Protobuf wire format
// using standard Go types, implementing the Protoscope specification functionality.
package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

// WireType represents Protobuf wire types
type WireType uint8

const (
	WireTypeVarint WireType = 0
	WireTypeI64    WireType = 1
	WireTypeLen    WireType = 2
	WireTypeSGroup WireType = 3
	WireTypeEGroup WireType = 4
	WireTypeI32    WireType = 5
)

// ProtoscopeError represents encoding/decoding errors
type ProtoscopeError struct {
	Message string
}

func (e *ProtoscopeError) Error() string {
	return e.Message
}

// =============================================================================
// VARINT ENCODER/DECODER
// =============================================================================

type VarintCodec struct{}

// EncodeVarint encodes a signed integer as varint
func (v *VarintCodec) EncodeVarint(value int64) []byte {
	return v.EncodeUvarint(uint64(value))
}

// EncodeUvarint encodes an unsigned integer as varint
func (v *VarintCodec) EncodeUvarint(value uint64) []byte {
	var result []byte
	for value >= 0x80 {
		result = append(result, byte(value)|0x80)
		value >>= 7
	}
	result = append(result, byte(value))
	return result
}

// EncodeVarintLongForm encodes varint with extra padding bytes
func (v *VarintCodec) EncodeVarintLongForm(value int64, extraBytes int) []byte {
	if extraBytes <= 0 {
		return v.EncodeVarint(value)
	}

	normal := v.EncodeVarint(value)
	if len(normal) == 0 {
		return normal
	}

	// Add padding by extending the encoding
	result := make([]byte, len(normal)-1, len(normal)+extraBytes)
	copy(result, normal[:len(normal)-1])

	// Make all but the last byte continue
	for i := range result {
		result[i] |= 0x80
	}

	// Add padding bytes
	for i := 0; i < extraBytes; i++ {
		result = append(result, 0x80)
	}

	// Add final byte
	result = append(result, normal[len(normal)-1])
	return result
}

// DecodeVarint decodes varint from bytes, returns (value, bytesConsumed, error)
func (v *VarintCodec) DecodeVarint(data []byte, offset int) (int64, int, error) {
	value, consumed, err := v.DecodeUvarint(data, offset)
	return int64(value), consumed, err
}

// DecodeUvarint decodes unsigned varint from bytes
func (v *VarintCodec) DecodeUvarint(data []byte, offset int) (uint64, int, error) {
	var result uint64
	var shift uint
	pos := offset

	for pos < len(data) {
		b := data[pos]
		if shift == 63 && b > 1 {
			return 0, 0, &ProtoscopeError{"varint overflows uint64"}
		}
		result |= uint64(b&0x7f) << shift
		pos++

		if b&0x80 == 0 {
			return result, pos - offset, nil
		}

		shift += 7
	}

	return 0, 0, &ProtoscopeError{"truncated varint"}
}

// =============================================================================
// ZIGZAG ENCODER/DECODER
// =============================================================================

type ZigzagCodec struct {
	varint *VarintCodec
}

func NewZigzagCodec() *ZigzagCodec {
	return &ZigzagCodec{varint: &VarintCodec{}}
}

// EncodeZigzag encodes signed integer using zigzag encoding
func (z *ZigzagCodec) EncodeZigzag(value int64) []byte {
	encoded := z.ZigzagEncode(value)
	return z.varint.EncodeUvarint(encoded)
}

// DecodeZigzag decodes zigzag encoded integer
func (z *ZigzagCodec) DecodeZigzag(data []byte, offset int) (int64, int, error) {
	encoded, consumed, err := z.varint.DecodeUvarint(data, offset)
	if err != nil {
		return 0, 0, err
	}
	value := z.ZigzagDecode(encoded)
	return value, consumed, nil
}

// ZigzagEncode applies zigzag encoding to signed integer
func (z *ZigzagCodec) ZigzagEncode(value int64) uint64 {
	return uint64((value << 1) ^ (value >> 63))
}

// ZigzagDecode applies zigzag decoding to get signed integer
func (z *ZigzagCodec) ZigzagDecode(value uint64) int64 {
	return int64((value >> 1) ^ uint64((int64(value&1)<<63)>>63))
}

// =============================================================================
// FIXED WIDTH INTEGER ENCODER/DECODER
// =============================================================================

type FixedIntCodec struct{}

// EncodeFixed32 encodes 32-bit integer in little-endian
func (f *FixedIntCodec) EncodeFixed32(value int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(value))
	return buf
}

// EncodeFixed64 encodes 64-bit integer in little-endian
func (f *FixedIntCodec) EncodeFixed64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

// DecodeFixed32 decodes 32-bit integer from little-endian bytes
func (f *FixedIntCodec) DecodeFixed32(data []byte, offset int) (int32, int, error) {
	if len(data) < offset+4 {
		return 0, 0, &ProtoscopeError{"insufficient bytes for fixed32"}
	}
	value := binary.LittleEndian.Uint32(data[offset : offset+4])
	return int32(value), 4, nil
}

// DecodeFixed64 decodes 64-bit integer from little-endian bytes
func (f *FixedIntCodec) DecodeFixed64(data []byte, offset int) (int64, int, error) {
	if len(data) < offset+8 {
		return 0, 0, &ProtoscopeError{"insufficient bytes for fixed64"}
	}
	value := binary.LittleEndian.Uint64(data[offset : offset+8])
	return int64(value), 8, nil
}

// =============================================================================
// FLOAT ENCODER/DECODER
// =============================================================================

type FloatCodec struct{}

// EncodeFloat32 encodes 32-bit float in IEEE 754 format
func (f *FloatCodec) EncodeFloat32(value float32) []byte {
	buf := make([]byte, 4)
	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(buf, bits)
	return buf
}

// EncodeFloat64 encodes 64-bit float in IEEE 754 format
func (f *FloatCodec) EncodeFloat64(value float64) []byte {
	buf := make([]byte, 8)
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(buf, bits)
	return buf
}

// DecodeFloat32 decodes 32-bit float from IEEE 754 bytes
func (f *FloatCodec) DecodeFloat32(data []byte, offset int) (float32, int, error) {
	if len(data) < offset+4 {
		return 0, 0, &ProtoscopeError{"insufficient bytes for float32"}
	}
	bits := binary.LittleEndian.Uint32(data[offset : offset+4])
	value := math.Float32frombits(bits)
	return value, 4, nil
}

// DecodeFloat64 decodes 64-bit float from IEEE 754 bytes
func (f *FloatCodec) DecodeFloat64(data []byte, offset int) (float64, int, error) {
	if len(data) < offset+8 {
		return 0, 0, &ProtoscopeError{"insufficient bytes for float64"}
	}
	bits := binary.LittleEndian.Uint64(data[offset : offset+8])
	value := math.Float64frombits(bits)
	return value, 8, nil
}

// =============================================================================
// BOOLEAN ENCODER/DECODER
// =============================================================================

type BoolCodec struct {
	varint *VarintCodec
}

func NewBoolCodec() *BoolCodec {
	return &BoolCodec{varint: &VarintCodec{}}
}

// EncodeBool encodes boolean as varint (0 or 1)
func (b *BoolCodec) EncodeBool(value bool) []byte {
	if value {
		return b.varint.EncodeVarint(1)
	}
	return b.varint.EncodeVarint(0)
}

// DecodeBool decodes boolean from varint
func (b *BoolCodec) DecodeBool(data []byte, offset int) (bool, int, error) {
	value, consumed, err := b.varint.DecodeVarint(data, offset)
	if err != nil {
		return false, 0, err
	}
	return value != 0, consumed, nil
}

// =============================================================================
// STRING/BYTES ENCODER/DECODER
// =============================================================================

type BytesCodec struct {
	varint *VarintCodec
}

func NewBytesCodec() *BytesCodec {
	return &BytesCodec{varint: &VarintCodec{}}
}

// EncodeBytes encodes byte slice with length prefix
func (b *BytesCodec) EncodeBytes(value []byte) []byte {
	length := b.varint.EncodeVarint(int64(len(value)))
	return append(length, value...)
}

// EncodeBytesLongForm encodes bytes with padded length prefix
func (b *BytesCodec) EncodeBytesLongForm(value []byte, extraBytes int) []byte {
	length := b.varint.EncodeVarintLongForm(int64(len(value)), extraBytes)
	return append(length, value...)
}

// EncodeString encodes string with length prefix
func (b *BytesCodec) EncodeString(value string) []byte {
	return b.EncodeBytes([]byte(value))
}

// DecodeBytes decodes length-prefixed byte slice
func (b *BytesCodec) DecodeBytes(data []byte, offset int) ([]byte, int, error) {
	length, lengthSize, err := b.varint.DecodeVarint(data, offset)
	if err != nil {
		return nil, 0, err
	}

	if length < 0 {
		return nil, 0, &ProtoscopeError{"negative length"}
	}

	start := offset + lengthSize
	end := start + int(length)

	if len(data) < end {
		return nil, 0, &ProtoscopeError{"insufficient bytes for length-prefixed data"}
	}

	value := make([]byte, length)
	copy(value, data[start:end])
	return value, lengthSize + int(length), nil
}

// DecodeString decodes length-prefixed string
func (b *BytesCodec) DecodeString(data []byte, offset int) (string, int, error) {
	bytes, consumed, err := b.DecodeBytes(data, offset)
	if err != nil {
		return "", 0, err
	}
	return string(bytes), consumed, nil
}

// =============================================================================
// TAG ENCODER/DECODER
// =============================================================================

type TagCodec struct {
	varint *VarintCodec
}

func NewTagCodec() *TagCodec {
	return &TagCodec{varint: &VarintCodec{}}
}

// EncodeTag encodes field number and wire type into tag
func (t *TagCodec) EncodeTag(fieldNumber int32, wireType WireType) []byte {
	if fieldNumber < 1 {
		panic("field number must be positive")
	}
	if wireType > 5 {
		panic("invalid wire type")
	}

	tag := (int64(fieldNumber) << 3) | int64(wireType)
	return t.varint.EncodeVarint(tag)
}

// DecodeTag decodes tag into field number and wire type
func (t *TagCodec) DecodeTag(data []byte, offset int) (int32, WireType, int, error) {
	tag, consumed, err := t.varint.DecodeVarint(data, offset)
	if err != nil {
		return 0, 0, 0, err
	}

	fieldNumber := int32(tag >> 3)
	wireType := WireType(tag & 0x7)

	if fieldNumber < 1 {
		return 0, 0, 0, &ProtoscopeError{"invalid field number"}
	}

	return fieldNumber, wireType, consumed, nil
}

// =============================================================================
// MESSAGE ENCODER/DECODER
// =============================================================================

type MessageCodec struct {
	varint *VarintCodec
	tag    *TagCodec
	bytes  *BytesCodec
}

func NewMessageCodec() *MessageCodec {
	varint := &VarintCodec{}
	return &MessageCodec{
		varint: varint,
		tag:    NewTagCodec(),
		bytes:  NewBytesCodec(),
	}
}

// EncodeMessage encodes message content with length prefix
func (m *MessageCodec) EncodeMessage(content []byte) []byte {
	return m.bytes.EncodeBytes(content)
}

// DecodeMessage decodes length-prefixed message
func (m *MessageCodec) DecodeMessage(data []byte, offset int) ([]byte, int, error) {
	return m.bytes.DecodeBytes(data, offset)
}

// =============================================================================
// FIELD VALUE ENCODER/DECODER
// =============================================================================

type FieldValue interface {
	EncodeValue() []byte
	WireType() WireType
}

// Varint field value
type VarintValue struct {
	Value int64
	codec *VarintCodec
}

func NewVarintValue(value int64) *VarintValue {
	return &VarintValue{Value: value, codec: &VarintCodec{}}
}

func (v *VarintValue) EncodeValue() []byte {
	return v.codec.EncodeVarint(v.Value)
}

func (v *VarintValue) WireType() WireType {
	return WireTypeVarint
}

// Fixed32 field value
type Fixed32Value struct {
	Value int32
	codec *FixedIntCodec
}

func NewFixed32Value(value int32) *Fixed32Value {
	return &Fixed32Value{Value: value, codec: &FixedIntCodec{}}
}

func (f *Fixed32Value) EncodeValue() []byte {
	return f.codec.EncodeFixed32(f.Value)
}

func (f *Fixed32Value) WireType() WireType {
	return WireTypeI32
}

// Fixed64 field value
type Fixed64Value struct {
	Value int64
	codec *FixedIntCodec
}

func NewFixed64Value(value int64) *Fixed64Value {
	return &Fixed64Value{Value: value, codec: &FixedIntCodec{}}
}

func (f *Fixed64Value) EncodeValue() []byte {
	return f.codec.EncodeFixed64(f.Value)
}

func (f *Fixed64Value) WireType() WireType {
	return WireTypeI64
}

// Bytes field value
type BytesValue struct {
	Value []byte
	codec *BytesCodec
}

func NewBytesValue(value []byte) *BytesValue {
	return &BytesValue{Value: value, codec: NewBytesCodec()}
}

func (b *BytesValue) EncodeValue() []byte {
	return b.codec.EncodeBytes(b.Value)
}

func (b *BytesValue) WireType() WireType {
	return WireTypeLen
}

// String field value
type StringValue struct {
	Value string
	codec *BytesCodec
}

func NewStringValue(value string) *StringValue {
	return &StringValue{Value: value, codec: NewBytesCodec()}
}

func (s *StringValue) EncodeValue() []byte {
	return s.codec.EncodeString(s.Value)
}

func (s *StringValue) WireType() WireType {
	return WireTypeLen
}

// =============================================================================
// FIELD ENCODER/DECODER
// =============================================================================

type Field struct {
	Number int32
	Value  FieldValue
}

type FieldCodec struct {
	tag *TagCodec
}

func NewFieldCodec() *FieldCodec {
	return &FieldCodec{tag: NewTagCodec()}
}

// EncodeField encodes complete field (tag + value)
func (f *FieldCodec) EncodeField(field *Field) []byte {
	tagBytes := f.tag.EncodeTag(field.Number, field.Value.WireType())
	valueBytes := field.Value.EncodeValue()
	return append(tagBytes, valueBytes...)
}

// =============================================================================
// DECODER FOR READING FIELDS
// =============================================================================

type FieldDecoder struct {
	tag    *TagCodec
	varint *VarintCodec
	fixed  *FixedIntCodec
	float  *FloatCodec
	bytes  *BytesCodec
	bool   *BoolCodec
	zigzag *ZigzagCodec
}

func NewFieldDecoder() *FieldDecoder {
	return &FieldDecoder{
		tag:    NewTagCodec(),
		varint: &VarintCodec{},
		fixed:  &FixedIntCodec{},
		float:  &FloatCodec{},
		bytes:  NewBytesCodec(),
		bool:   NewBoolCodec(),
		zigzag: NewZigzagCodec(),
	}
}

// DecodeFieldHeader decodes field tag
func (d *FieldDecoder) DecodeFieldHeader(data []byte, offset int) (int32, WireType, int, error) {
	return d.tag.DecodeTag(data, offset)
}

// DecodeVarintField decodes varint field value
func (d *FieldDecoder) DecodeVarintField(data []byte, offset int) (int64, int, error) {
	return d.varint.DecodeVarint(data, offset)
}

// DecodeBytesField decodes bytes field value
func (d *FieldDecoder) DecodeBytesField(data []byte, offset int) ([]byte, int, error) {
	return d.bytes.DecodeBytes(data, offset)
}

// DecodeStringField decodes string field value
func (d *FieldDecoder) DecodeStringField(data []byte, offset int) (string, int, error) {
	return d.bytes.DecodeString(data, offset)
}

// DecodeFixed32Field decodes fixed32 field value
func (d *FieldDecoder) DecodeFixed32Field(data []byte, offset int) (int32, int, error) {
	return d.fixed.DecodeFixed32(data, offset)
}

// DecodeFixed64Field decodes fixed64 field value
func (d *FieldDecoder) DecodeFixed64Field(data []byte, offset int) (int64, int, error) {
	return d.fixed.DecodeFixed64(data, offset)
}

// =============================================================================
// REPEATED FIELD ENCODER/DECODER
// =============================================================================

type RepeatedCodec struct {
	varint *VarintCodec
	bytes  *BytesCodec
}

func NewRepeatedCodec() *RepeatedCodec {
	return &RepeatedCodec{
		varint: &VarintCodec{},
		bytes:  NewBytesCodec(),
	}
}

// EncodeRepeatedVarint encodes repeated varint as separate fields
func (r *RepeatedCodec) EncodeRepeatedVarint(values []int64) [][]byte {
	var result [][]byte
	for _, value := range values {
		result = append(result, r.varint.EncodeVarint(value))
	}
	return result
}

// EncodePackedVarint encodes repeated varint as packed field
func (r *RepeatedCodec) EncodePackedVarint(values []int64) []byte {
	var packed []byte
	for _, value := range values {
		packed = append(packed, r.varint.EncodeVarint(value)...)
	}
	return r.bytes.EncodeBytes(packed)
}

// EncodePackedFixed32 encodes repeated fixed32 as packed field
func (r *RepeatedCodec) EncodePackedFixed32(values []int32) []byte {
	packed := make([]byte, 0, len(values)*4)
	for _, value := range values {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(value))
		packed = append(packed, buf...)
	}
	return r.bytes.EncodeBytes(packed)
}

// EncodePackedFixed64 encodes repeated fixed64 as packed field
func (r *RepeatedCodec) EncodePackedFixed64(values []int64) []byte {
	packed := make([]byte, 0, len(values)*8)
	for _, value := range values {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(value))
		packed = append(packed, buf...)
	}
	return r.bytes.EncodeBytes(packed)
}

// EncodePackedFloat32 encodes repeated float32 as packed field
func (r *RepeatedCodec) EncodePackedFloat32(values []float32) []byte {
	packed := make([]byte, 0, len(values)*4)
	for _, value := range values {
		buf := make([]byte, 4)
		bits := math.Float32bits(value)
		binary.LittleEndian.PutUint32(buf, bits)
		packed = append(packed, buf...)
	}
	return r.bytes.EncodeBytes(packed)
}

// EncodePackedFloat64 encodes repeated float64 as packed field
func (r *RepeatedCodec) EncodePackedFloat64(values []float64) []byte {
	packed := make([]byte, 0, len(values)*8)
	for _, value := range values {
		buf := make([]byte, 8)
		bits := math.Float64bits(value)
		binary.LittleEndian.PutUint64(buf, bits)
		packed = append(packed, buf...)
	}
	return r.bytes.EncodeBytes(packed)
}

// DecodePackedVarint decodes packed varint field
func (r *RepeatedCodec) DecodePackedVarint(data []byte, offset int) ([]int64, int, error) {
	packedData, consumed, err := r.bytes.DecodeBytes(data, offset)
	if err != nil {
		return nil, 0, err
	}

	var values []int64
	pos := 0
	for pos < len(packedData) {
		value, size, err := r.varint.DecodeVarint(packedData, pos)
		if err != nil {
			return nil, 0, err
		}
		values = append(values, value)
		pos += size
	}

	return values, consumed, nil
}

// DecodePackedFixed32 decodes packed fixed32 field
func (r *RepeatedCodec) DecodePackedFixed32(data []byte, offset int) ([]int32, int, error) {
	packedData, consumed, err := r.bytes.DecodeBytes(data, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(packedData)%4 != 0 {
		return nil, 0, &ProtoscopeError{"packed fixed32 data length not multiple of 4"}
	}

	var values []int32
	for i := 0; i < len(packedData); i += 4 {
		value := binary.LittleEndian.Uint32(packedData[i : i+4])
		values = append(values, int32(value))
	}

	return values, consumed, nil
}

// =============================================================================
// ENUM ENCODER/DECODER
// =============================================================================

type EnumCodec struct {
	varint *VarintCodec
}

func NewEnumCodec() *EnumCodec {
	return &EnumCodec{varint: &VarintCodec{}}
}

// EncodeEnum encodes enum value as varint
func (e *EnumCodec) EncodeEnum(value int32) []byte {
	return e.varint.EncodeVarint(int64(value))
}

// DecodeEnum decodes enum value from varint
func (e *EnumCodec) DecodeEnum(data []byte, offset int) (int32, int, error) {
	value, consumed, err := e.varint.DecodeVarint(data, offset)
	if err != nil {
		return 0, 0, err
	}
	return int32(value), consumed, nil
}

// =============================================================================
// MAP ENCODER/DECODER
// =============================================================================

type MapCodec struct {
	field   *FieldCodec
	message *MessageCodec
	tag     *TagCodec
}

func NewMapCodec() *MapCodec {
	return &MapCodec{
		field:   NewFieldCodec(),
		message: NewMessageCodec(),
		tag:     NewTagCodec(),
	}
}

// EncodeStringInt64Map encodes map[string]int64 as repeated message entries
func (m *MapCodec) EncodeStringInt64Map(mapData map[string]int64) [][]byte {
	var entries [][]byte

	for key, value := range mapData {
		// Each map entry is a message with field 1 (key) and field 2 (value)
		keyField := &Field{Number: 1, Value: NewStringValue(key)}
		valueField := &Field{Number: 2, Value: NewVarintValue(value)}

		keyBytes := m.field.EncodeField(keyField)
		valueBytes := m.field.EncodeField(valueField)

		entryMessage := append(keyBytes, valueBytes...)
		entries = append(entries, m.message.EncodeMessage(entryMessage))
	}

	return entries
}

// EncodeInt32StringMap encodes map[int32]string as repeated message entries
func (m *MapCodec) EncodeInt32StringMap(mapData map[int32]string) [][]byte {
	var entries [][]byte

	for key, value := range mapData {
		keyField := &Field{Number: 1, Value: NewVarintValue(int64(key))}
		valueField := &Field{Number: 2, Value: NewStringValue(value)}

		keyBytes := m.field.EncodeField(keyField)
		valueBytes := m.field.EncodeField(valueField)

		entryMessage := append(keyBytes, valueBytes...)
		entries = append(entries, m.message.EncodeMessage(entryMessage))
	}

	return entries
}

// DecodeStringInt64MapEntry decodes single map entry for map[string]int64
func (m *MapCodec) DecodeStringInt64MapEntry(data []byte, offset int) (string, int64, int, error) {
	entryData, consumed, err := m.message.DecodeMessage(data, offset)
	if err != nil {
		return "", 0, 0, err
	}

	decoder := NewFieldDecoder()
	pos := 0
	var key string
	var value int64

	for pos < len(entryData) {
		fieldNum, wireType, headerSize, err := decoder.DecodeFieldHeader(entryData, pos)
		if err != nil {
			return "", 0, 0, err
		}
		pos += headerSize

		switch fieldNum {
		case 1: // Key field
			if wireType != WireTypeLen {
				return "", 0, 0, &ProtoscopeError{"map key must be string"}
			}
			key, headerSize, err = decoder.DecodeStringField(entryData, pos)
			if err != nil {
				return "", 0, 0, err
			}
			pos += headerSize

		case 2: // Value field
			if wireType != WireTypeVarint {
				return "", 0, 0, &ProtoscopeError{"map value must be varint"}
			}
			value, headerSize, err = decoder.DecodeVarintField(entryData, pos)
			if err != nil {
				return "", 0, 0, err
			}
			pos += headerSize
		}
	}

	return key, value, consumed, nil
}

// =============================================================================
// HIGH-LEVEL FIELD VALUES FOR REPEATED/MAP/ENUM
// =============================================================================

// RepeatedVarintValue for repeated varint fields
type RepeatedVarintValue struct {
	Values []int64
	Packed bool
	codec  *RepeatedCodec
}

func NewRepeatedVarintValue(values []int64, packed bool) *RepeatedVarintValue {
	return &RepeatedVarintValue{
		Values: values,
		Packed: packed,
		codec:  NewRepeatedCodec(),
	}
}

func (r *RepeatedVarintValue) EncodeValue() []byte {
	if r.Packed {
		return r.codec.EncodePackedVarint(r.Values)
	}
	// For non-packed, this would need special handling at the field level
	// since each value becomes a separate field
	panic("non-packed repeated fields require field-level handling")
}

func (r *RepeatedVarintValue) WireType() WireType {
	if r.Packed {
		return WireTypeLen
	}
	return WireTypeVarint
}

// EnumValue for enum fields
type EnumValue struct {
	Value int32
	codec *EnumCodec
}

func NewEnumValue(value int32) *EnumValue {
	return &EnumValue{Value: value, codec: NewEnumCodec()}
}

func (e *EnumValue) EncodeValue() []byte {
	return e.codec.EncodeEnum(e.Value)
}

func (e *EnumValue) WireType() WireType {
	return WireTypeVarint
}

// MapStringInt64Value for map[string]int64 fields
type MapStringInt64Value struct {
	Value map[string]int64
	codec *MapCodec
}

func NewMapStringInt64Value(value map[string]int64) *MapStringInt64Value {
	return &MapStringInt64Value{Value: value, codec: NewMapCodec()}
}

func (m *MapStringInt64Value) EncodeValue() []byte {
	// Maps require special handling at field level since each entry is a separate field
	panic("map fields require field-level handling")
}

func (m *MapStringInt64Value) WireType() WireType {
	return WireTypeLen
}

// =============================================================================
// EXAMPLE USAGE
// =============================================================================

func main() {
	// Example 1: Simple message with basic fields
	fieldCodec := NewFieldCodec()

	field1 := &Field{Number: 1, Value: NewVarintValue(150)}
	field2 := &Field{Number: 2, Value: NewStringValue("testing")}
	field3 := &Field{Number: 3, Value: NewEnumValue(42)}

	message := append(fieldCodec.EncodeField(field1), fieldCodec.EncodeField(field2)...)
	message = append(message, fieldCodec.EncodeField(field3)...)

	fmt.Printf("Basic message: %x\n", message)

	// Example 2: Packed repeated field
	repeatedCodec := NewRepeatedCodec()
	packedInts := repeatedCodec.EncodePackedVarint([]int64{1, 2, 3, 4, 5})

	packedField := &Field{Number: 4, Value: NewBytesValue(packedInts[len(repeatedCodec.varint.EncodeVarint(int64(len(packedInts)-len(repeatedCodec.varint.EncodeVarint(int64(len(packedInts))))))):])}

	fmt.Printf("Packed field: %x\n", fieldCodec.EncodeField(packedField))

	// Example 3: Map field
	mapCodec := NewMapCodec()
	mapData := map[string]int64{"key1": 100, "key2": 200}
	mapEntries := mapCodec.EncodeStringInt64Map(mapData)

	fmt.Printf("Map entries count: %d\n", len(mapEntries))
	for i, entry := range mapEntries {
		fmt.Printf("Map entry %d: %x\n", i, entry)
	}

	// Example 4: Decode packed repeated field
	testPacked := repeatedCodec.EncodePackedVarint([]int64{10, 20, 30})
	values, _, err := repeatedCodec.DecodePackedVarint(testPacked, 0)
	if err != nil {
		fmt.Printf("Error decoding packed: %v\n", err)
	} else {
		fmt.Printf("Decoded packed values: %v\n", values)
	}

	// Example 5: Decode map entry
	if len(mapEntries) > 0 {
		key, value, _, err := mapCodec.DecodeStringInt64MapEntry(mapEntries[0], 0)
		if err != nil {
			fmt.Printf("Error decoding map entry: %v\n", err)
		} else {
			fmt.Printf("Decoded map entry: %s -> %d\n", key, value)
		}
	}

	// Example 6: Different packed types
	packedFloats := repeatedCodec.EncodePackedFloat32([]float32{1.5, 2.5, 3.5})
	packedFixed := repeatedCodec.EncodePackedFixed32([]int32{100, 200, 300})

	fmt.Printf("Packed floats: %x\n", packedFloats)
	fmt.Printf("Packed fixed32: %x\n", packedFixed)

	// Decode them back
	floatValues, _, err := repeatedCodec.DecodePackedFixed32(packedFixed, 0)
	if err != nil {
		fmt.Printf("Error decoding fixed32: %v\n", err)
	} else {
		fmt.Printf("Decoded fixed32 values: %v\n", floatValues)
	}

	// Example 7: Enum field test
	fmt.Println("\n--- Enum Tests ---")
	enumCodec := NewEnumCodec()

	// Test different enum values
	enumValues := []int32{0, 1, 255, -1}
	for _, val := range enumValues {
		encoded := enumCodec.EncodeEnum(val)
		decoded, _, err := enumCodec.DecodeEnum(encoded, 0)
		if err != nil {
			fmt.Printf("Error decoding enum %d: %v\n", val, err)
		} else {
			fmt.Printf("Enum %d -> %x -> %d\n", val, encoded, decoded)
		}
	}

	// Example 8: Repeated struct (nested messages)
	fmt.Println("\n--- Repeated Struct Tests ---")

	// Create a "Person" message with id and name
	createPersonMessage := func(id int64, name string) []byte {
		idField := &Field{Number: 1, Value: NewVarintValue(id)}
		nameField := &Field{Number: 2, Value: NewStringValue(name)}

		idBytes := fieldCodec.EncodeField(idField)
		nameBytes := fieldCodec.EncodeField(nameField)

		personData := append(idBytes, nameBytes...)
		return personData
	}

	// Create repeated struct field (repeated Person messages)
	person1 := createPersonMessage(1, "Alice")
	person2 := createPersonMessage(2, "Bob")
	person3 := createPersonMessage(3, "Charlie")

	fmt.Printf("Person 1 message: %x\n", person1)
	fmt.Printf("Person 2 message: %x\n", person2)
	fmt.Printf("Person 3 message: %x\n", person3)

	// Encode as repeated message fields (field 5)
	messageCodec := NewMessageCodec()
	repeatedPersonField1 := &Field{Number: 5, Value: NewBytesValue(person1)}
	repeatedPersonField2 := &Field{Number: 5, Value: NewBytesValue(person2)}
	repeatedPersonField3 := &Field{Number: 5, Value: NewBytesValue(person3)}

	repeatedPersons := append(fieldCodec.EncodeField(repeatedPersonField1), fieldCodec.EncodeField(repeatedPersonField2)...)
	repeatedPersons = append(repeatedPersons, fieldCodec.EncodeField(repeatedPersonField3)...)

	fmt.Printf("Repeated persons field: %x\n", repeatedPersons)

	// Example 9: Map with enum values
	fmt.Println("\n--- Map with Enum Values Tests ---")

	// Create map[string]enum where enum values are status codes
	statusMap := map[string]int32{
		"pending":  0,
		"active":   1,
		"inactive": 2,
		"deleted":  99,
	}

	// Encode map entries where value is enum
	var enumMapEntries [][]byte
	for key, enumValue := range statusMap {
		keyField := &Field{Number: 1, Value: NewStringValue(key)}
		// Encode enum as varint value
		enumField := &Field{Number: 2, Value: NewEnumValue(enumValue)}

		keyBytes := fieldCodec.EncodeField(keyField)
		valueBytes := fieldCodec.EncodeField(enumField)

		entryMessage := append(keyBytes, valueBytes...)
		encodedEntry := messageCodec.EncodeMessage(entryMessage)
		enumMapEntries = append(enumMapEntries, encodedEntry)

		fmt.Printf("Map entry '%s' -> %d: %x\n", key, enumValue, encodedEntry)
	}

	// Decode one of the enum map entries
	if len(enumMapEntries) > 0 {
		entryData, _, err := messageCodec.DecodeMessage(enumMapEntries[0], 0)
		if err != nil {
			fmt.Printf("Error decoding enum map entry: %v\n", err)
		} else {
			// Decode the fields within the entry
			decoder := NewFieldDecoder()
			pos := 0
			var key string
			var enumVal int32

			for pos < len(entryData) {
				fieldNum, _, headerSize, err := decoder.DecodeFieldHeader(entryData, pos)
				if err != nil {
					fmt.Printf("Error decoding header: %v\n", err)
					break
				}
				pos += headerSize

				switch fieldNum {
				case 1: // Key
					key, headerSize, err = decoder.DecodeStringField(entryData, pos)
					if err != nil {
						fmt.Printf("Error decoding key: %v\n", err)
					}
					pos += headerSize
				case 2: // Enum value
					val, headerSize, err := decoder.DecodeVarintField(entryData, pos)
					if err != nil {
						fmt.Printf("Error decoding enum: %v\n", err)
					} else {
						enumVal = int32(val)
					}
					pos += headerSize
				}
			}

			fmt.Printf("Decoded enum map entry: '%s' -> %d\n", key, enumVal)
		}
	}

	// Example 10: Complex nested structure test
	fmt.Println("\n--- Complex Nested Structure ---")

	// Create a "Record" with multiple field types
	recordData := []byte{}

	// Field 1: Regular int
	field1 = &Field{Number: 1, Value: NewVarintValue(12345)}
	recordData = append(recordData, fieldCodec.EncodeField(field1)...)

	// Field 2: Enum status
	statusField := &Field{Number: 2, Value: NewEnumValue(1)} // active
	recordData = append(recordData, fieldCodec.EncodeField(statusField)...)

	// Field 3: Repeated tags (packed)
	tags := []int64{10, 20, 30, 40}
	packedTags := repeatedCodec.EncodePackedVarint(tags)
	tagsField := &Field{Number: 3, Value: NewBytesValue(packedTags[len(repeatedCodec.varint.EncodeVarint(int64(len(packedTags)-len(repeatedCodec.varint.EncodeVarint(int64(len(packedTags))))))):])}
	recordData = append(recordData, fieldCodec.EncodeField(tagsField)...)

	// Field 4: String name
	nameField := &Field{Number: 4, Value: NewStringValue("test-record")}
	recordData = append(recordData, fieldCodec.EncodeField(nameField)...)

	fmt.Printf("Complex record: %x\n", recordData)
	fmt.Printf("Record length: %d bytes\n", len(recordData))
}
