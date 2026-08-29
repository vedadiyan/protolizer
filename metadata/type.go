package metadata

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	aloc "github.com/vedadiyan/protolizer/memory"
	"github.com/vedadiyan/protolizer/pdk"
	"github.com/vedadiyan/protolizer/util"
)

type (
	Tags struct {
		Protobuf *ProtobufInfo `protobuf:"bytes,1,opt,name=protobuf,proto3"`
		JsonName string        `protobuf:"bytes,2,opt,name=json_name,proto3"`
		MapKey   pdk.WireType  `protobuf:"varint,3,opt,name=map_key,proto3,enum"`
		MapValue pdk.WireType  `protobuf:"varint,4,opt,name=map_value,proto3,enum"`
	}

	ProtobufInfo struct {
		WireType pdk.WireType `protobuf:"varint,1,opt,name=wire_type,proto3,enum"`
		FieldNum int          `protobuf:"varint,2,opt,name=field_num,proto3"`
		Label    string       `protobuf:"bytes,3,opt,name=label,proto3"`
		Name     string       `protobuf:"bytes,4,opt,name=name,proto3"`
		Syntax   string       `protobuf:"bytes,5,opt,name=syntax,proto3"`
		OneOf    bool         `protobuf:"varint,6,opt,name=one_of,proto3"`
	}

	Field struct {
		Name             string       `protobuf:"bytes,1,opt,name=name,proto3"`
		Kind             reflect.Kind `protobuf:"varint,2,opt,name=kind,proto3,enum"`
		Key              reflect.Kind `protobuf:"varint,3,opt,name=key,proto3,enum"`
		Index            reflect.Kind `protobuf:"varint,4,opt,name=index,proto3"`
		KeyType          string       `protobuf:"bytes,5,opt,name=key_type,proto3,enum"`
		IndexType        string       `protobuf:"bytes,6,opt,name=index_type,proto3,enum"`
		FieldIndex       []int        `protobuf:"varint,7,rep,packed,name=field_index,proto3"`
		IsPointer        bool         `protobuf:"varint,8,opt,name=is_pointer,proto3"`
		TypeName         string       `protobuf:"bytes,9,opt,name=type_name,proto3"`
		ConcreteTypeName string       `protobuf:"bytes,10,opt,name=type_name,proto3"`
		Tags             *Tags        `protobuf:"bytes,11,opt,name=tags,proto3"`
		Tag              []byte       `protobuf:"bytes,12,opt,name=tag,proto3"`
		KeyTag           []byte       `protobuf:"bytes,13,opt,name=key_tags,proto3"`
		ValueTag         []byte       `protobuf:"bytes,14,opt,name=value_tags,proto3"`
	}

	Type struct {
		Name          string         `protobuf:"bytes,1,opt,name=name,proto3"`
		Fields        []*Field       `protobuf:"bytes,2,rep,name=fields,proto3"`
		FieldsIndexer map[int]*Field `protobuf:"bytes,3,rep,name=fields_indexer,proto3" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	}

	Module struct {
		Types map[string]*Type `protobuf:"bytes,1,rep,name=types,proto3" protobuf_key:"string,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	}
)

const (
	WireTypeVarint pdk.WireType = 0
	WireTypeI64    pdk.WireType = 1
	WireTypeLen    pdk.WireType = 2
	WireTypeSGroup pdk.WireType = 3
	WireTypeEGroup pdk.WireType = 4
	WireTypeI32    pdk.WireType = 5
)

var (
	_registry map[string]*Type
	_mut      sync.Mutex
)

func init() {
	_registry = make(map[string]*Type)
	RegisterTypeFor[Tags]()
	RegisterTypeFor[ProtobufInfo]()
	RegisterTypeFor[Field]()
	RegisterTypeFor[Type]()
	RegisterTypeFor[Module]()
}

func RegisterTypeFor[T any]() {
	t := reflect.TypeFor[T]()
	registerType(t, TypeName(t))
}

func RegisterTypeAs[T any](name string) {
	registerType(reflect.TypeFor[T](), name)
}

func RegisterType(t reflect.Type) {
	registerType(t, TypeName(t))
}

func registerType(t reflect.Type, name string) {
	_mut.Lock()
	if _, ok := _registry[name]; ok {
		_mut.Unlock()
		return
	}
	_mut.Unlock()

	elemType := util.GetElemenType(t)
	if elemType.Kind() != reflect.Struct {
		return
	}

	out := new(Type)

	out.Name = name
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

	for i := range elemType.NumField() {
		f := elemType.Field(i)
		RegisterType(f.Type)
	}

	_registry[name] = out
}

func TypeName(t reflect.Type) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.String()
}

func ConcreteTypeName(t reflect.Type) string {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return t.String()
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
			elem := f.Type.Elem()
			if elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}
			out.Index = elem.Kind()
			out.IndexType = TypeName(elem)
		}
	case reflect.Map:
		{
			elem := f.Type.Elem()
			if elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}
			out.Key = f.Type.Key().Kind()
			out.KeyType = TypeName(f.Type.Key())
			out.Index = elem.Kind()
			out.IndexType = TypeName(elem)
		}
	}
	if out.IsPointer {
		out.TypeName = TypeName(f.Type.Elem())
		out.ConcreteTypeName = ConcreteTypeName(f.Type.Elem())
	} else {
		out.TypeName = TypeName(f.Type)
		out.ConcreteTypeName = ConcreteTypeName(f.Type)
	}

	if out.Tags.Protobuf == nil {
		return out
	}
	w := out.Tags.Protobuf.WireType
	if out.Kind == reflect.Slice {
		w = WireTypeLen
	}
	tag, err := pdk.TagEncode(int32(out.Tags.Protobuf.FieldNum), w)
	if err != nil {
		panic(err)
	}
	out.Tag = bytes.Clone(tag.Bytes())
	aloc.Dealloc(tag)

	keyTag, err := pdk.TagEncode(int32(1), out.Tags.MapKey)
	if err != nil {
		panic(err)
	}
	out.KeyTag = bytes.Clone(keyTag.Bytes())
	aloc.Dealloc(keyTag)

	valueTag, err := pdk.TagEncode(int32(2), out.Tags.MapValue)
	if err != nil {
		panic(err)
	}
	out.ValueTag = bytes.Clone(valueTag.Bytes())
	aloc.Dealloc(valueTag)
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

func getWireType(str string) pdk.WireType {
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

// func ExportType[T any]() ([]byte, error) {
// 	t := CaptureTypeFor[T]()
// 	return Marshal(t)
// }

// func ImportType(bytes []byte) (*Type, error) {
// 	t := new(Type)
// 	if err := Unmarshal(bytes, t); err != nil {
// 		return nil, err
// 	}
// 	return t, nil
// }

// func exportModule(t reflect.Type) (*Module, error) {
// 	module := new(Module)
// 	module.Types = make(map[string]*Type)
// 	module.Types[TypeName(t)] = CaptureType(t)
// 	for i := range t.NumField() {
// 		fieldType := t.Field(i).Type
// 		if fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map {
// 			fieldType = fieldType.Elem()
// 		}
// 		if fieldType.Kind() == reflect.Pointer {
// 			fieldType = fieldType.Elem()
// 		}
// 		if fieldType.Kind() == reflect.Struct {
// 			modules, err := exportModule(fieldType)
// 			if err != nil {
// 				return nil, err
// 			}
// 			for key, value := range modules.Types {
// 				module.Types[key] = value
// 			}
// 			continue
// 		}
// 	}
// 	return module, nil
// }

// func ExportModule[T any]() ([]byte, error) {
// 	modules, err := exportModule(reflect.TypeFor[T]())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return Marshal(modules)
// }

// func ImportModule(bytes []byte) (*Module, error) {
// 	module := new(Module)
// 	err := Unmarshal(bytes, module)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return module, nil
// }
