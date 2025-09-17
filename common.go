package protolizer

import (
	"fmt"
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
	for i := range elemType.NumField() {
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

func EncodeVarint(u uint64) []byte {
	var buf []byte
	for u >= 0x80 {
		buf = append(buf, byte(u)|0x80)
		u >>= 7
	}
	buf = append(buf, byte(u))
	return buf
}

func DecodeVarint(buf []byte, i *int) (uint64, error) {
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

func WireTypeFromNum(n int) WireType {
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

		encoded, err := EncodeField(field, fieldVal)
		if err != nil {
			return nil, fmt.Errorf("error encoding field %s: %w", field.Name, err)
		}

		result = append(result, encoded...)
	}

	return result, nil
}

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
		fieldNum, wireType, err := DecodeTag(data, &i)
		if err != nil {
			return fmt.Errorf("error decoding tag: %w", err)
		}

		fieldIndex, exists := fieldMap[fieldNum]
		if !exists {
			// Skip unknown field
			if err := SkipField(data, &i, wireType); err != nil {
				return fmt.Errorf("error skipping unknown field %d: %w", fieldNum, err)
			}
			continue
		}

		field := typ.Fields[fieldIndex]
		fieldVal := val.Field(fieldIndex)

		if err := DecodeField(field, fieldVal, data, &i, wireType); err != nil {
			return fmt.Errorf("error decoding field %s: %w", field.Name, err)
		}
	}

	return nil
}
