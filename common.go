package protolizer

import (
	"encoding/binary"
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
		GetValue  func(v reflect.Value) *Serializable
		SetValue  func(v reflect.Value, value any)
	}
	Type struct {
		Fields []*Field
	}
)

const (
	WIRETYPE_VARINT           WireType = "varint"
	WIRETYPE_FIXED_64         WireType = "fixed64"
	WIRETYPE_LENGTH_DELIMITED WireType = "bytes"
	WRIETYPE_START_GROUP
	WIRETYPE_END_GROUP
	WIRETYPE_FIXED_32 WireType = "fixed32"
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
	out.Fields = make([]*Field, t.Elem().NumField())
	for i := range t.Elem().NumField() {
		f := NewField(t.Elem().Field(i))
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
	out.Kind, out.IsPointer = GetKind(f.Type)
	out.Tags = NewTags(f.Tag)
	out.GetValue = func(v reflect.Value) *Serializable {
		if out.IsPointer {
			ser := new(Serializable)
			ser.Value = v.FieldByIndex(f.Index).Elem().Interface()
			return ser
		}
		ser := new(Serializable)
		ser.Value = v.FieldByIndex(f.Index).Interface()
		return ser
	}
	out.SetValue = func(v reflect.Value, value any) {
		v.Set(reflect.ValueOf(value))
	}
	return out
}

func ParseProtoTag(tag string) *ProtobufInfo {
	tag = strings.TrimPrefix(tag, "protobuf:\"")
	tag = strings.TrimSuffix(tag, "\"")

	segments := strings.Split(tag, ",")
	l := len(segments)

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
	if len(segments) == 5 {
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

func (ser *Serializable) Serialize(wireType WireType) ([]byte, error) {
	switch wireType {
	case WIRETYPE_VARINT:
		{
			var u uint64
			switch v := ser.Value.(type) {
			case int:
				u = uint64(v)
			case int32:
				u = uint64(v)
			case int64:
				u = uint64(v)
			case uint32:
				u = uint64(v)
			case uint64:
				u = v
			case bool:
				if v {
					u = 1
				} else {
					u = 0
				}
			default:
				return nil, fmt.Errorf("varint expects int/bool, got %T", ser.Value)
			}

			return encodeVarint(u), nil
		}

	case WIRETYPE_LENGTH_DELIMITED:
		{
			v, ok := ser.Value.([]byte)
			if !ok {
				return nil, fmt.Errorf("length-delimited expects []byte, got %T", ser.Value)
			}
			lenBuf := encodeVarint(uint64(len(v)))
			return append(lenBuf, v...), nil
		}

	case WIRETYPE_FIXED_64:
		{
			v, ok := ser.Value.(uint64)
			if !ok {
				return nil, fmt.Errorf("fixed64 expects uint64, got %T", ser.Value)
			}
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, v)
			return buf, nil
		}

	case WIRETYPE_FIXED_32:
		{
			v, ok := ser.Value.(uint32)
			if !ok {
				return nil, fmt.Errorf("fixed32 expects uint32, got %T", ser.Value)
			}
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, v)
			return buf, nil
		}

	default:
		{
			return nil, fmt.Errorf("unexpected wire type %s", wireType)
		}
	}
}

func (ser *Serializable) SerializeRepeated(wireType WireType, packed bool) ([][]byte, error) {
	values, ok := ser.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("repeated expects []any, got %T", ser.Value)
	}

	if packed {
		// --- packed encoding ---
		var packed []byte
		for _, v := range values {
			elem := &Serializable{Value: v}
			b, err := elem.Serialize(wireType)
			if err != nil {
				return nil, err
			}
			packed = append(packed, b...)
		}

		lenBuf := encodeVarint(uint64(len(packed)))
		return [][]byte{append(lenBuf, packed...)}, nil
	}

	// --- unpacked encoding ---
	var out [][]byte
	for _, v := range values {
		elem := &Serializable{Value: v}
		b, err := elem.Serialize(wireType)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func wireTypeNum(wt WireType) int {
	switch wt {
	case WIRETYPE_VARINT:
		return 0
	case WIRETYPE_FIXED_64:
		return 1
	case WIRETYPE_LENGTH_DELIMITED:
		return 2
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

// SerializeStruct takes a struct value (pointer to struct) and encodes it as protobuf
func SerializeStruct(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("SerializeStruct expects pointer to struct, got %T", v)
	}
	rt := rv.Type()

	t := RegisterType(rt) // fetch reflection info
	var out []byte

	for _, f := range t.Fields {
		if f.Tags == nil || !f.Tags.IsProtobuf() {
			continue
		}
		info := f.Tags.Protobuf
		ser := f.GetValue(rv.Elem())
		wireNum := wireTypeNum(info.WireType)
		if wireNum < 0 {
			return nil, fmt.Errorf("invalid wire type %s", info.WireType)
		}

		// Field tag = (field_number << 3) | wire_type
		tag := uint64(info.FieldNum<<3 | wireNum)

		if strings.HasPrefix(info.Label, "repeated") {
			// repeated field
			bufs, err := ser.SerializeRepeated(info.WireType, true) // you can decide packed/unpacked
			if err != nil {
				return nil, err
			}
			for _, b := range bufs {
				out = append(out, encodeVarint(tag)...) // field tag
				out = append(out, b...)                 // payload
			}
		} else {
			// single field
			valBytes, err := ser.Serialize(info.WireType)
			if err != nil {
				return nil, err
			}
			out = append(out, encodeVarint(tag)...)
			out = append(out, valBytes...)
		}
	}
	return out, nil
}
