package codec

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type (
	WireType     uint8
	Serializable struct {
		Value any
	}
	Tags struct {
		Protobuf *ProtobufInfo
		JsonName string
		MapKey   WireType
		MapValue WireType
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
		Index     []int
		IsPointer bool
		Tags      *Tags
	}
	Type struct {
		Fields []*Field
	}
)

const (
	WireTypeVarint WireType = 0
	WireTypeI64    WireType = 1
	WireTypeLen    WireType = 2
	WireTypeSGroup WireType = 3
	WireTypeEGroup WireType = 4
	WireTypeI32    WireType = 5
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
	sort.Slice(out.Fields, func(i, j int) bool {
		return out.Fields[i].Tags.Protobuf.FieldNum < out.Fields[j].Tags.Protobuf.FieldNum
	})
	_registry[t] = out
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
	out.Index = f.Index
	out.Tags = NewTags(f.Tag)
	return out
}

func NewTags(t reflect.StructTag) *Tags {
	out := new(Tags)
	if tag, ok := t.Lookup("protobuf"); ok {
		out.Protobuf = ParseProtoTag(tag)
	}
	if tag, ok := t.Lookup("protobuf_key"); ok {
		out.MapKey = ParseProtoTag(tag).WireType
	}
	if tag, ok := t.Lookup("protobuf_val"); ok {
		out.MapValue = ParseProtoTag(tag).WireType
	}
	out.JsonName = t.Get("json")
	return out
}

func GetWireType(str string) WireType {
	switch str {
	case "varint":
		{
			return WireTypeVarint
		}
	case "fixed64":
		{
			return WireTypeI64
		}
	case "bytes":
		{
			return WireTypeLen
		}
	case "start_group":
		{
			return WireTypeSGroup
		}
	case "end_group":
		{
			return WireTypeEGroup
		}
	case "fixed32":
		{
			return WireTypeI32
		}
	}
	return 0
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
		out.WireType = GetWireType(segments[0])
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
