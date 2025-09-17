package protolizer

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type (
	WireType     string
	Serializable struct {
		Value any
	}
	Tags struct {
		Protobuf *ProtobufInfo
		JsonName string
		MapKey   string
		MapValue string
	}
	ProtobufInfo struct {
		WireType WireType
		FieldNum int
		Label    string
		Name     string
		Syntax   string
		OneOf    bool
	}
	Field struct {
		Name      string
		Kind      reflect.Kind
		IsPointer bool
		Tags      *Tags
	}
	Type struct {
		Fields []*Field
	}
)

const (
	WIRETYPE_VARINT           WireType = "varint"
	WIRETYPE_FIXED_64         WireType = "fixed64"
	WIRETYPE_LENGTH_DELIMITED WireType = "bytes"
	WIRETYPE_START_GROUP      WireType = "start_group"
	WIRETYPE_END_GROUP        WireType = "end_group"
	WIRETYPE_FIXED_32         WireType = "fixed32"
)

var (
	_registry map[reflect.Type]*Type
	_mut      sync.Mutex
)

func init() {
	_registry = make(map[reflect.Type]*Type)
}

func RegisterType(t reflect.Type) *Type {
	_mut.Lock()
	defer _mut.Unlock()
	if value, ok := _registry[t]; ok {
		return value
	}
	out := new(Type)

	// Handle pointer types
	elemType := t
	if t.Kind() == reflect.Ptr {
		elemType = t.Elem()
	}

	out.Fields = make([]*Field, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		f := NewField(elemType.Field(i))
		out.Fields[i] = f
	}
	_registry[t] = out
	return out
}

func NewTags(t reflect.StructTag) *Tags {
	out := new(Tags)
	if tag, ok := t.Lookup("protobuf"); ok {
		out.Protobuf = ParseProtoTag(tag)
	}
	out.JsonName = t.Get("json")
	return out
}

func NewField(f reflect.StructField) *Field {
	out := new(Field)
	out.Name = f.Name
	out.Kind = f.Type.Kind()

	// Check for pointer types
	if f.Type.Kind() == reflect.Ptr {
		out.IsPointer = true
		out.Kind = f.Type.Elem().Kind()
	}

	out.Tags = NewTags(f.Tag)
	return out
}

func ParseProtoTag(tag string) *ProtobufInfo {
	// Handle different protobuf tag formats
	tag = strings.Trim(tag, "\"")
	if strings.HasPrefix(tag, "protobuf:") {
		tag = strings.TrimPrefix(tag, "protobuf:")
		tag = strings.Trim(tag, "\"")
	}

	segments := strings.Split(tag, ",")
	l := len(segments)

	if l < 2 {
		return nil
	}

	fieldNum, err := strconv.Atoi(segments[1])
	if err != nil {
		panic(fmt.Errorf("invalid field number: %w", err))
	}

	out := new(ProtobufInfo)
	if l > 0 {
		out.WireType = WireType(segments[0])
	}
	if l > 1 {
		out.FieldNum = fieldNum
	}
	if l > 2 {
		out.Label = segments[2]
	}
	if l > 3 {
		out.Name = strings.TrimPrefix(segments[3], "name=")
	}
	if l > 4 {
		out.Syntax = segments[4]
	}
	if len(segments) == 6 {
		out.OneOf = true
	}

	return out
}

func GetKind(t reflect.Type) (reflect.Kind, bool) {
	if t.Kind() == reflect.Pointer {
		return t.Elem().Kind(), true
	}
	return t.Kind(), false
}

func (t *Tags) IsProtobuf() bool {
	return t.Protobuf != nil
}

func wireTypeNum(wt WireType) int {
	switch wt {
	case WIRETYPE_VARINT:
		return 0
	case WIRETYPE_FIXED_64:
		return 1
	case WIRETYPE_LENGTH_DELIMITED:
		return 2
	case WIRETYPE_START_GROUP:
		return 3
	case WIRETYPE_END_GROUP:
		return 4
	case WIRETYPE_FIXED_32:
		return 5
	default:
		return -1
	}
}

func encodeVarint(u uint64) []byte {
	var buf []byte
	for u >= 0x80 {
		buf = append(buf, byte(u)|0x80)
		u >>= 7
	}
	buf = append(buf, byte(u))
	return buf
}

func decodeVarint(buf []byte, i *int) (uint64, error) {
	var result uint64
	var shift uint
	for {
		if *i >= len(buf) {
			return 0, fmt.Errorf("buffer underflow in varint")
		}
		b := buf[*i]
		*i++
		result |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
		shift += 7
		if shift >= 64 {
			return 0, fmt.Errorf("varint overflow")
		}
	}
	return result, nil
}

func wireTypeFromNum(n int) WireType {
	switch n {
	case 0:
		return WIRETYPE_VARINT
	case 1:
		return WIRETYPE_FIXED_64
	case 2:
		return WIRETYPE_LENGTH_DELIMITED
	case 3:
		return WIRETYPE_START_GROUP
	case 4:
		return WIRETYPE_END_GROUP
	case 5:
		return WIRETYPE_FIXED_32
	default:
		return ""
	}
}

// Encode a tag (field number + wire type)
func encodeTag(fieldNum int, wireType WireType) []byte {
	tag := (fieldNum << 3) | wireTypeNum(wireType)
	return encodeVarint(uint64(tag))
}

// Decode a tag (field number + wire type)
func decodeTag(buf []byte, i *int) (int, WireType, error) {
	tag, err := decodeVarint(buf, i)
	if err != nil {
		return 0, "", err
	}
	fieldNum := int(tag >> 3)
	wireType := wireTypeFromNum(int(tag & 0x7))
	return fieldNum, wireType, nil
}

// Encode fixed32 value
func encodeFixed32(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

// Decode fixed32 value
func decodeFixed32(buf []byte, i *int) (uint32, error) {
	if *i+4 > len(buf) {
		return 0, fmt.Errorf("buffer underflow in fixed32")
	}
	result := binary.LittleEndian.Uint32(buf[*i:])
	*i += 4
	return result, nil
}

// Encode fixed64 value
func encodeFixed64(v uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return buf
}

// Decode fixed64 value
func decodeFixed64(buf []byte, i *int) (uint64, error) {
	if *i+8 > len(buf) {
		return 0, fmt.Errorf("buffer underflow in fixed64")
	}
	result := binary.LittleEndian.Uint64(buf[*i:])
	*i += 8
	return result, nil
}

// Encode length-delimited value (bytes, string, embedded message)
func encodeLengthDelimited(data []byte) []byte {
	result := encodeVarint(uint64(len(data)))
	result = append(result, data...)
	return result
}

// Decode length-delimited value
func decodeLengthDelimited(buf []byte, i *int) ([]byte, error) {
	length, err := decodeVarint(buf, i)
	if err != nil {
		return nil, err
	}
	if *i+int(length) > len(buf) {
		return nil, fmt.Errorf("buffer underflow in length-delimited")
	}
	result := make([]byte, length)
	copy(result, buf[*i:*i+int(length)])
	*i += int(length)
	return result, nil
}

// Marshal serializes a struct to protobuf binary format
func Marshal(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("cannot marshal nil pointer")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("can only marshal struct types")
	}

	typ := RegisterType(reflect.TypeOf(v))
	var result []byte

	for i, field := range typ.Fields {
		if !field.Tags.IsProtobuf() {
			continue
		}

		fieldVal := val.Field(i)
		if field.IsPointer {
			if fieldVal.IsNil() {
				continue // Skip nil pointers
			}
			fieldVal = fieldVal.Elem()
		}

		// Skip zero values for optional fields
		if fieldVal.IsZero() && field.Tags.Protobuf.Label != "required" {
			continue
		}

		encoded, err := encodeField(field, fieldVal)
		if err != nil {
			return nil, fmt.Errorf("error encoding field %s: %w", field.Name, err)
		}

		result = append(result, encoded...)
	}

	return result, nil
}

func encodeField(field *Field, val reflect.Value) ([]byte, error) {
	// Handle repeated fields and maps (but not []byte which is handled as length-delimited)
	if field.Kind == reflect.Map || (field.Kind == reflect.Slice && val.Type().Elem().Kind() != reflect.Uint8) {
		return encodeRepeatedOrMap(field, val)
	}

	tag := encodeTag(field.Tags.Protobuf.FieldNum, field.Tags.Protobuf.WireType)
	var data []byte

	switch field.Tags.Protobuf.WireType {
	case WIRETYPE_VARINT:
		var u uint64
		switch field.Kind {
		case reflect.Bool:
			if val.Bool() {
				u = 1
			}
		case reflect.Int32, reflect.Int64, reflect.Int:
			u = uint64(val.Int())
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			u = val.Uint()
		default:
			return nil, fmt.Errorf("unsupported varint type: %v", field.Kind)
		}
		data = encodeVarint(u)

	case WIRETYPE_FIXED_32:
		switch field.Kind {
		case reflect.Float32:
			u := math.Float32bits(float32(val.Float()))
			data = encodeFixed32(u)
		case reflect.Uint32:
			data = encodeFixed32(uint32(val.Uint()))
		case reflect.Int32:
			data = encodeFixed32(uint32(val.Int()))
		default:
			return nil, fmt.Errorf("unsupported fixed32 type: %v", field.Kind)
		}

	case WIRETYPE_FIXED_64:
		switch field.Kind {
		case reflect.Float64:
			u := math.Float64bits(val.Float())
			data = encodeFixed64(u)
		case reflect.Uint64:
			data = encodeFixed64(val.Uint())
		case reflect.Int64:
			data = encodeFixed64(uint64(val.Int()))
		default:
			return nil, fmt.Errorf("unsupported fixed64 type: %v", field.Kind)
		}

	case WIRETYPE_LENGTH_DELIMITED:
		switch field.Kind {
		case reflect.String:
			data = encodeLengthDelimited([]byte(val.String()))
		case reflect.Slice:
			if val.Type().Elem().Kind() == reflect.Uint8 {
				// []byte
				data = encodeLengthDelimited(val.Bytes())
			} else {
				return nil, fmt.Errorf("unsupported slice type for length-delimited: %v", val.Type())
			}
		case reflect.Struct:
			// Embedded message
			embedded, err := Marshal(val.Interface())
			if err != nil {
				return nil, err
			}
			data = encodeLengthDelimited(embedded)
		default:
			return nil, fmt.Errorf("unsupported length-delimited type: %v", field.Kind)
		}

	default:
		return nil, fmt.Errorf("unsupported wire type: %v", field.Tags.Protobuf.WireType)
	}

	result := make([]byte, len(tag)+len(data))
	copy(result, tag)
	copy(result[len(tag):], data)
	return result, nil
}

func encodeRepeatedOrMap(field *Field, val reflect.Value) ([]byte, error) {
	var result []byte

	if field.Kind == reflect.Map {
		// Handle maps as repeated key-value pairs
		for _, key := range val.MapKeys() {
			mapVal := val.MapIndex(key)

			// Create a map entry message
			var entryData []byte

			// Key (field 1)
			keyTag := encodeTag(1, WIRETYPE_LENGTH_DELIMITED)
			keyData := encodeLengthDelimited([]byte(key.String()))
			entryData = append(entryData, keyTag...)
			entryData = append(entryData, keyData...)

			// Value (field 2)
			valueTag := encodeTag(2, WIRETYPE_LENGTH_DELIMITED)
			valueData := encodeLengthDelimited([]byte(mapVal.String()))
			entryData = append(entryData, valueTag...)
			entryData = append(entryData, valueData...)

			// Wrap the entry in length-delimited encoding
			tag := encodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_LENGTH_DELIMITED)
			data := encodeLengthDelimited(entryData)

			result = append(result, tag...)
			result = append(result, data...)
		}
		return result, nil
	}

	// Handle repeated fields (slices)
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		elemKind := elem.Kind()

		var tag []byte
		var data []byte

		switch elemKind {
		case reflect.String:
			tag = encodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_LENGTH_DELIMITED)
			data = encodeLengthDelimited([]byte(elem.String()))

		case reflect.Bool, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint32, reflect.Uint64, reflect.Uint:
			tag = encodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_VARINT)
			var u uint64
			switch elemKind {
			case reflect.Bool:
				if elem.Bool() {
					u = 1
				}
			case reflect.Int32, reflect.Int64, reflect.Int:
				u = uint64(elem.Int())
			case reflect.Uint32, reflect.Uint64, reflect.Uint:
				u = elem.Uint()
			}
			data = encodeVarint(u)

		case reflect.Float32:
			tag = encodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_FIXED_32)
			u := math.Float32bits(float32(elem.Float()))
			data = encodeFixed32(u)

		case reflect.Float64:
			tag = encodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_FIXED_64)
			u := math.Float64bits(elem.Float())
			data = encodeFixed64(u)

		case reflect.Struct:
			tag = encodeTag(field.Tags.Protobuf.FieldNum, WIRETYPE_LENGTH_DELIMITED)
			embedded, err := Marshal(elem.Interface())
			if err != nil {
				return nil, err
			}
			data = encodeLengthDelimited(embedded)

		default:
			return nil, fmt.Errorf("unsupported repeated element type: %v", elemKind)
		}

		result = append(result, tag...)
		result = append(result, data...)
	}

	return result, nil
}

// Unmarshal deserializes protobuf binary data into a struct
func Unmarshal(data []byte, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshal target must be a non-nil pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("unmarshal target must be a pointer to struct")
	}

	typ := RegisterType(reflect.TypeOf(v))

	// Create field number to field index map
	fieldMap := make(map[int]int)
	for i, field := range typ.Fields {
		if field.Tags.IsProtobuf() {
			fieldMap[field.Tags.Protobuf.FieldNum] = i
		}
	}

	// Initialize slices and maps for repeated fields
	for i, field := range typ.Fields {
		if !field.Tags.IsProtobuf() {
			continue
		}
		fieldVal := val.Field(i)
		if field.Kind == reflect.Slice && fieldVal.IsNil() {
			fieldVal.Set(reflect.MakeSlice(fieldVal.Type(), 0, 0))
		}
		if field.Kind == reflect.Map && fieldVal.IsNil() {
			fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
		}
	}

	i := 0
	for i < len(data) {
		fieldNum, wireType, err := decodeTag(data, &i)
		if err != nil {
			return fmt.Errorf("error decoding tag: %w", err)
		}

		fieldIndex, exists := fieldMap[fieldNum]
		if !exists {
			// Skip unknown field
			if err := skipField(data, &i, wireType); err != nil {
				return fmt.Errorf("error skipping unknown field %d: %w", fieldNum, err)
			}
			continue
		}

		field := typ.Fields[fieldIndex]
		fieldVal := val.Field(fieldIndex)

		if err := decodeField(field, fieldVal, data, &i, wireType); err != nil {
			return fmt.Errorf("error decoding field %s: %w", field.Name, err)
		}
	}

	return nil
}

func decodeField(field *Field, val reflect.Value, buf []byte, i *int, wireType WireType) error {
	// Handle repeated fields (slices) and maps (but not []byte which is handled as length-delimited)
	if field.Kind == reflect.Map || (field.Kind == reflect.Slice && val.Type().Elem().Kind() != reflect.Uint8) {
		return decodeRepeatedOrMap(field, val, buf, i, wireType)
	}

	if field.IsPointer {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}

	switch wireType {
	case WIRETYPE_VARINT:
		u, err := decodeVarint(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.Bool:
			val.SetBool(u != 0)
		case reflect.Int32, reflect.Int64, reflect.Int:
			val.SetInt(int64(u))
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			val.SetUint(u)
		default:
			return fmt.Errorf("unsupported varint type: %v", field.Kind)
		}

	case WIRETYPE_FIXED_32:
		u, err := decodeFixed32(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.Float32:
			val.SetFloat(float64(math.Float32frombits(u)))
		case reflect.Uint32:
			val.SetUint(uint64(u))
		case reflect.Int32:
			val.SetInt(int64(int32(u)))
		default:
			return fmt.Errorf("unsupported fixed32 type: %v", field.Kind)
		}

	case WIRETYPE_FIXED_64:
		u, err := decodeFixed64(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.Float64:
			val.SetFloat(math.Float64frombits(u))
		case reflect.Uint64:
			val.SetUint(u)
		case reflect.Int64:
			val.SetInt(int64(u))
		default:
			return fmt.Errorf("unsupported fixed64 type: %v", field.Kind)
		}

	case WIRETYPE_LENGTH_DELIMITED:
		data, err := decodeLengthDelimited(buf, i)
		if err != nil {
			return err
		}
		switch field.Kind {
		case reflect.String:
			val.SetString(string(data))
		case reflect.Slice:
			if val.Type().Elem().Kind() == reflect.Uint8 {
				val.SetBytes(data)
			} else {
				return fmt.Errorf("unsupported slice type for length-delimited: %v", val.Type())
			}
		case reflect.Struct:
			// Embedded message
			if err := Unmarshal(data, val.Addr().Interface()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported length-delimited type: %v", field.Kind)
		}

	default:
		return fmt.Errorf("unsupported wire type: %v", wireType)
	}

	return nil
}

func decodeRepeatedOrMap(field *Field, val reflect.Value, buf []byte, i *int, wireType WireType) error {
	if field.Kind == reflect.Map {
		// Handle map entry
		if wireType != WIRETYPE_LENGTH_DELIMITED {
			return fmt.Errorf("map field must use length-delimited wire type")
		}

		entryData, err := decodeLengthDelimited(buf, i)
		if err != nil {
			return err
		}

		// Decode key-value pairs from the entry
		var key, value string
		j := 0
		for j < len(entryData) {
			fieldNum, entryWireType, err := decodeTag(entryData, &j)
			if err != nil {
				return err
			}

			if entryWireType != WIRETYPE_LENGTH_DELIMITED {
				return fmt.Errorf("map entry fields must be length-delimited")
			}

			data, err := decodeLengthDelimited(entryData, &j)
			if err != nil {
				return err
			}

			if fieldNum == 1 {
				key = string(data)
			} else if fieldNum == 2 {
				value = string(data)
			}
		}

		val.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return nil
	}

	// Handle repeated fields (slices)
	elemType := val.Type().Elem()

	var newElem reflect.Value
	switch wireType {
	case WIRETYPE_VARINT:
		newElem = reflect.New(elemType).Elem()
		u, err := decodeVarint(buf, i)
		if err != nil {
			return err
		}

		switch elemType.Kind() {
		case reflect.Bool:
			newElem.SetBool(u != 0)
		case reflect.Int32, reflect.Int64, reflect.Int:
			newElem.SetInt(int64(u))
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			newElem.SetUint(u)
		default:
			return fmt.Errorf("unsupported repeated varint type: %v", elemType.Kind())
		}

	case WIRETYPE_FIXED_32:
		newElem = reflect.New(elemType).Elem()
		u, err := decodeFixed32(buf, i)
		if err != nil {
			return err
		}

		switch elemType.Kind() {
		case reflect.Float32:
			newElem.SetFloat(float64(math.Float32frombits(u)))
		case reflect.Uint32:
			newElem.SetUint(uint64(u))
		case reflect.Int32:
			newElem.SetInt(int64(int32(u)))
		default:
			return fmt.Errorf("unsupported repeated fixed32 type: %v", elemType.Kind())
		}

	case WIRETYPE_FIXED_64:
		newElem = reflect.New(elemType).Elem()
		u, err := decodeFixed64(buf, i)
		if err != nil {
			return err
		}

		switch elemType.Kind() {
		case reflect.Float64:
			newElem.SetFloat(math.Float64frombits(u))
		case reflect.Uint64:
			newElem.SetUint(u)
		case reflect.Int64:
			newElem.SetInt(int64(u))
		default:
			return fmt.Errorf("unsupported repeated fixed64 type: %v", elemType.Kind())
		}

	case WIRETYPE_LENGTH_DELIMITED:
		data, err := decodeLengthDelimited(buf, i)
		if err != nil {
			return err
		}

		if elemType.Kind() == reflect.String {
			newElem = reflect.ValueOf(string(data))
		} else if elemType.Kind() == reflect.Struct {
			// Repeated embedded message
			newElem = reflect.New(elemType).Elem()
			if err := Unmarshal(data, newElem.Addr().Interface()); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported repeated length-delimited type: %v", elemType.Kind())
		}

	default:
		return fmt.Errorf("unsupported wire type for repeated field: %v", wireType)
	}

	val.Set(reflect.Append(val, newElem))
	return nil
}

func skipField(buf []byte, i *int, wireType WireType) error {
	switch wireType {
	case WIRETYPE_VARINT:
		_, err := decodeVarint(buf, i)
		return err
	case WIRETYPE_FIXED_32:
		if *i+4 > len(buf) {
			return fmt.Errorf("buffer underflow skipping fixed32")
		}
		*i += 4
		return nil
	case WIRETYPE_FIXED_64:
		if *i+8 > len(buf) {
			return fmt.Errorf("buffer underflow skipping fixed64")
		}
		*i += 8
		return nil
	case WIRETYPE_LENGTH_DELIMITED:
		_, err := decodeLengthDelimited(buf, i)
		return err
	default:
		return fmt.Errorf("cannot skip unsupported wire type: %v", wireType)
	}
}
