package protolizer

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type (
	WireType uint8
	Tags     struct {
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
		Name       string
		Kind       reflect.Kind
		Key        reflect.Kind
		Index      reflect.Kind
		FieldIndex []int
		IsPointer  bool
		TypeName   string
		Tags       *Tags
	}
	Type struct {
		Fields        []*Field
		FieldsIndexer map[int]*Field
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
	_registry map[string]*Type
)

func init() {
	_registry = make(map[string]*Type)
}

func RegisterTypeFor[T any]() {
	out := new(Type)

	t := reflect.TypeFor[T]()
	elemType := t
	if t.Kind() == reflect.Ptr {
		elemType = t.Elem()
	}

	out.Fields = make([]*Field, 0)
	for i := range elemType.NumField() {
		f := newField(elemType.Field(i))
		if !f.Tags.isProtobuf() {
			continue
		}
		out.Fields = append(out.Fields, f)
	}
	sort.Slice(out.Fields, func(i, j int) bool {
		return out.Fields[i].Tags.Protobuf.FieldNum < out.Fields[j].Tags.Protobuf.FieldNum
	})

	out.FieldsIndexer = make(map[int]*Field)
	for _, i := range out.Fields {
		out.FieldsIndexer[i.Tags.Protobuf.FieldNum] = i
	}

	_registry[TypeName(t)] = out
}

func TypeName(t reflect.Type) string {
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

func CaptureTypeFor[T any]() *Type {
	return _registry[TypeName(reflect.TypeFor[T]())]
}

func CaptureType(t reflect.Type) *Type {
	return _registry[TypeName(t)]
}

func CaptureTypeByName(typeName string) *Type {
	return _registry[typeName]
}

func newField(f reflect.StructField) *Field {
	out := new(Field)
	out.Name = f.Name
	out.Kind = f.Type.Kind()

	if f.Type.Kind() == reflect.Ptr {
		out.IsPointer = true
		out.Kind = f.Type.Elem().Kind()
	}
	out.FieldIndex = f.Index
	out.Tags = newTags(f.Tag)
	switch out.Kind {
	case reflect.Array, reflect.Slice:
		{
			out.Index = f.Type.Elem().Kind()
		}
	case reflect.Map:
		{
			out.Key = f.Type.Key().Kind()
			out.Index = f.Type.Elem().Kind()
		}
	}
	out.TypeName = TypeName(f.Type)
	return out
}

func newTags(t reflect.StructTag) *Tags {
	out := new(Tags)
	if tag, ok := t.Lookup("protobuf"); ok {
		out.Protobuf = parseProtoTag(tag)
	}
	if tag, ok := t.Lookup("protobuf_key"); ok {
		out.MapKey = parseProtoTag(tag).WireType
	}
	if tag, ok := t.Lookup("protobuf_val"); ok {
		out.MapValue = parseProtoTag(tag).WireType
	}
	out.JsonName = t.Get("json")
	return out
}

func getWireType(str string) WireType {
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

func parseProtoTag(tag string) *ProtobufInfo {
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
		out.WireType = getWireType(segments[0])
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

func (t *Tags) isProtobuf() bool {
	return t.Protobuf != nil
}
