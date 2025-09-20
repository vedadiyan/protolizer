package protolizer

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/vedadiyan/protolizer/codec"
)

type (
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
		WireType codec.WireType
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
		Encode    func(reflect.Value) ([]byte, error)
		Decode    func([]byte, reflect.Value, int) (int, error)
	}
	Type struct {
		Fields []*Field
	}
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

	elemType := t
	if t.Kind() == reflect.Ptr {
		elemType = t.Elem()
	}

	out.Fields = make([]*Field, 0)
	for i := range elemType.NumField() {
		f := NewField(elemType.Field(i))
		if !f.Tags.IsProtobuf() {
			continue
		}
		out.Fields = append(out.Fields, f)
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

	if f.Type.Kind() == reflect.Ptr {
		out.IsPointer = true
		out.Kind = f.Type.Elem().Kind()
	}
	out.Tags = NewTags(f.Tag)
	out.Encode = func(v reflect.Value) ([]byte, error) {
		if out.IsPointer {
			v = v.Elem()
		}
		tag, err := codec.EncodeTag(int32(out.Tags.Protobuf.FieldNum), out.Tags.Protobuf.WireType)
		if err != nil {
			return nil, err
		}
		switch out.Kind {
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
			{
				if out.Tags.Protobuf.WireType == codec.WireTypeI32 {
					return append(tag, codec.EncodeFixed32(int32(v.Int()))...), nil
				}
				if out.Tags.Protobuf.WireType == codec.WireTypeI64 {
					return append(tag, codec.EncodeFixed64(v.Int())...), nil
				}
				return append(tag, codec.EncodeVarint(v.Int())...), nil
			}
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
			{
				return append(tag, codec.EncodeUvarint(v.Uint())...), nil
			}
		case reflect.Float32:
			{
				return append(tag, codec.EncodeFloat32(float32(v.Float()))...), nil
			}
		case reflect.Float64:
			{
				return append(tag, codec.EncodeFloat64(v.Float())...), nil
			}
		case reflect.Bool:
			{
				return append(tag, codec.EncodeBool(v.Bool())...), nil
			}
		case reflect.String:
			{
				return append(tag, codec.EncodeString(v.String())...), nil
			}
		}
		return nil, fmt.Errorf("")
	}
	out.Decode = func(b []byte, v reflect.Value, pos int) (int, error) {
		if out.IsPointer {
			v = v.Elem()
		}
		_, _, n, err := codec.DecodeTag(b, pos)
		if err != nil {
			return pos, err
		}
		pos += n
		switch out.Kind {
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
			{
				if out.Tags.Protobuf.WireType == codec.WireTypeI32 {
					value, consumed, err := codec.DecodeFixed32(b, pos)
					if err != nil {
						return pos, err
					}
					v.SetInt(int64(value))
					return pos + consumed, nil
				}
				if out.Tags.Protobuf.WireType == codec.WireTypeI64 {
					value, consumed, err := codec.DecodeFixed64(b, pos)
					if err != nil {
						return pos, err
					}
					v.SetInt(value)
					return pos + consumed, nil
				}
				value, consumed, err := codec.DecodeVarint(b, pos)
				if err != nil {
					return pos, err
				}
				v.SetInt(value)
				return pos + consumed, nil
			}
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
			{
				value, consumed, err := codec.DecodeUvarint(b, pos)
				if err != nil {
					return pos, err
				}
				v.SetUint(value)
				return pos + consumed, nil
			}
		case reflect.Float32:
			{
				value, consumed, err := codec.DecodeFloat32(b, pos)
				if err != nil {
					return pos, err
				}
				v.SetFloat(float64(value))
				return pos + consumed, nil
			}
		case reflect.Float64:
			{
				value, consumed, err := codec.DecodeFloat64(b, pos)
				if err != nil {
					return pos, err
				}
				v.SetFloat(value)
				return pos + consumed, nil
			}
		case reflect.Bool:
			{
				value, consumed, err := codec.DecodeBool(b, pos)
				if err != nil {
					return pos, err
				}
				v.SetBool(value)
				return pos + consumed, nil
			}
		case reflect.String:
			{
				value, consumed, err := codec.DecodeString(b, pos)
				if err != nil {
					return pos, err
				}
				v.SetString(value)
				return pos + consumed, nil
			}
		}
		return pos, fmt.Errorf("")
	}
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
		out.WireType = codec.GetWireType(segments[0])
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
