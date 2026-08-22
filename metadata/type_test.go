package metadata

import (
	"reflect"
	"strings"
	"testing"

	"github.com/vedadiyan/protolizer/pdk"
)

func TestTypeName(t *testing.T) {
	type Named struct{}

	tests := []struct {
		name string
		typ  reflect.Type
		want string
	}{
		{"struct", reflect.TypeOf(Named{}), reflect.TypeOf(Named{}).String()},
		{"pointer", reflect.TypeOf((*Named)(nil)), reflect.TypeOf(Named{}).String()},
		{"double pointer", reflect.TypeOf((**Named)(nil)), reflect.TypeOf(Named{}).String()},
		{"int", reflect.TypeOf(int(0)), "int"},
		{"string", reflect.TypeOf(""), "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeName(tt.typ); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetWireType(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want pdk.WireType
	}{
		{"varint", "varint", WireTypeVarint},
		{"fixed64", "fixed64", WireTypeI64},
		{"bytes", "bytes", WireTypeLen},
		{"start group", "start_group", WireTypeSGroup},
		{"end group", "end_group", WireTypeEGroup},
		{"fixed32", "fixed32", WireTypeI32},
		{"unknown", "unknown", WireTypeVarint},
		{"empty", "", WireTypeVarint},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getWireType(tt.in); got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseProtoTag(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *ProtobufInfo
	}{
		{
			"full",
			"bytes,1,opt,name=foo,proto3",
			&ProtobufInfo{
				WireType: WireTypeLen,
				FieldNum: 1,
				Label:    "opt",
				Name:     "foo",
				Syntax:   "proto3",
			},
		},
		{
			"quoted",
			`"varint,2,opt,name=id,proto3"`,
			&ProtobufInfo{
				WireType: WireTypeVarint,
				FieldNum: 2,
				Label:    "opt",
				Name:     "id",
				Syntax:   "proto3",
			},
		},
		{
			"protobuf prefix",
			`protobuf:bytes,3,opt,name=data,proto3`,
			&ProtobufInfo{
				WireType: WireTypeLen,
				FieldNum: 3,
				Label:    "opt",
				Name:     "data",
				Syntax:   "proto3",
			},
		},
		{
			"oneof",
			"bytes,5,opt,name=value,proto3,oneof",
			&ProtobufInfo{
				WireType: WireTypeLen,
				FieldNum: 5,
				Label:    "opt",
				Name:     "value",
				Syntax:   "proto3",
				OneOf:    true,
			},
		},
		{
			"two segments",
			"varint,10",
			&ProtobufInfo{
				WireType: WireTypeVarint,
				FieldNum: 10,
			},
		},
		{
			"three segments",
			"fixed32,11,rep",
			&ProtobufInfo{
				WireType: WireTypeI32,
				FieldNum: 11,
				Label:    "rep",
			},
		},
		{
			"four segments",
			"fixed64,12,opt,name=created",
			&ProtobufInfo{
				WireType: WireTypeI64,
				FieldNum: 12,
				Label:    "opt",
				Name:     "created",
			},
		},
		{
			"unknown wire type",
			"unknown,15,opt,name=x,proto3",
			&ProtobufInfo{
				WireType: WireTypeVarint,
				FieldNum: 15,
				Label:    "opt",
				Name:     "x",
				Syntax:   "proto3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseProtoTag(tt.in)

			if got == nil {
				t.Fatal("got nil")
			}

			if *got != *tt.want {
				t.Fatalf("got %+v, want %+v", *got, *tt.want)
			}
		})
	}
}

func TestParseProtoTagTooShort(t *testing.T) {
	if got := parseProtoTag("varint"); got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}

func TestParseProtoTagInvalidFieldNumber(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	parseProtoTag("varint,invalid")
}

func TestTagsIsProtobuf(t *testing.T) {
	tests := []struct {
		name string
		tags *Tags
		want bool
	}{
		{"nil protobuf", &Tags{}, false},
		{"protobuf", &Tags{Protobuf: &ProtobufInfo{}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tags.isProtobuf(); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTags(t *testing.T) {
	type TestStruct struct {
		Value string         `protobuf:"bytes,1,opt,name=value,proto3" json:"value"`
		Map   map[string]int `protobuf:"bytes,2,rep,name=map,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value" json:"map"`
		Plain int            `json:"plain"`
	}

	typ := reflect.TypeOf(TestStruct{})

	t.Run("value", func(t *testing.T) {
		tags := newTags(typ.Field(0).Tag)

		if tags.Protobuf == nil {
			t.Fatal("expected protobuf metadata")
		}

		if tags.Protobuf.WireType != WireTypeLen {
			t.Fatalf("got wire type %d, want %d", tags.Protobuf.WireType, WireTypeLen)
		}

		if tags.Protobuf.FieldNum != 1 {
			t.Fatalf("got field number %d, want 1", tags.Protobuf.FieldNum)
		}

		if tags.JsonName != "value" {
			t.Fatalf("got json name %q, want value", tags.JsonName)
		}
	})

	t.Run("map", func(t *testing.T) {
		tags := newTags(typ.Field(1).Tag)

		if tags.MapKey != WireTypeLen {
			t.Fatalf("got map key wire type %d, want %d", tags.MapKey, WireTypeLen)
		}

		if tags.MapValue != WireTypeVarint {
			t.Fatalf("got map value wire type %d, want %d", tags.MapValue, WireTypeVarint)
		}

		if tags.JsonName != "map" {
			t.Fatalf("got json name %q, want map", tags.JsonName)
		}
	})

	t.Run("plain", func(t *testing.T) {
		tags := newTags(typ.Field(2).Tag)

		if tags.Protobuf != nil {
			t.Fatal("expected nil protobuf metadata")
		}

		if tags.JsonName != "plain" {
			t.Fatalf("got json name %q, want plain", tags.JsonName)
		}
	})
}

func TestNewTagsWithoutOptionalTags(t *testing.T) {
	typ := reflect.TypeOf(struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}{})

	tags := newTags(typ.Field(0).Tag)

	if tags.Protobuf == nil {
		t.Fatal("expected protobuf metadata")
	}

	if tags.JsonName != "" {
		t.Fatalf("got json name %q, want empty", tags.JsonName)
	}

	if tags.MapKey != WireTypeVarint {
		t.Fatalf("got map key %d, want 0", tags.MapKey)
	}

	if tags.MapValue != WireTypeVarint {
		t.Fatalf("got map value %d, want 0", tags.MapValue)
	}
}

func TestNewFieldScalar(t *testing.T) {
	type Example struct {
		Value int `protobuf:"varint,7,opt,name=value,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Name != "Value" {
		t.Fatalf("got name %q, want Value", field.Name)
	}

	if field.Kind != reflect.Int {
		t.Fatalf("got kind %v, want int", field.Kind)
	}

	if field.IsPointer {
		t.Fatal("expected IsPointer false")
	}

	if field.TypeName != "int" {
		t.Fatalf("got type name %q, want int", field.TypeName)
	}

	if len(field.FieldIndex) != 1 || field.FieldIndex[0] != 0 {
		t.Fatalf("got field index %v, want [0]", field.FieldIndex)
	}

	if field.Tags == nil || field.Tags.Protobuf == nil {
		t.Fatal("expected protobuf tags")
	}

	if len(field.Tag) == 0 {
		t.Fatal("expected encoded tag")
	}

	if len(field.KeyTag) == 0 {
		t.Fatal("expected key tag")
	}

	if len(field.ValueTag) == 0 {
		t.Fatal("expected value tag")
	}
}

func TestNewFieldPointer(t *testing.T) {
	type Example struct {
		Value *int `protobuf:"varint,8,opt,name=value,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if !field.IsPointer {
		t.Fatal("expected pointer")
	}

	if field.Kind != reflect.Int {
		t.Fatalf("got kind %v, want int", field.Kind)
	}

	if field.TypeName != "int" {
		t.Fatalf("got type name %q, want int", field.TypeName)
	}

	if len(field.Tag) == 0 {
		t.Fatal("expected encoded tag")
	}
}

func TestNewFieldSlice(t *testing.T) {
	type Example struct {
		Values []int `protobuf:"varint,9,rep,name=values,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Kind != reflect.Slice {
		t.Fatalf("got kind %v, want slice", field.Kind)
	}

	if field.Index != reflect.Int {
		t.Fatalf("got index %v, want int", field.Index)
	}

	if field.IndexType != "int" {
		t.Fatalf("got index type %q, want int", field.IndexType)
	}

	if field.TypeName != "[]int" {
		t.Fatalf("got type name %q, want []int", field.TypeName)
	}

	if len(field.Tag) == 0 {
		t.Fatal("expected encoded tag")
	}

	expected, err := pdk.TagEncode(9, WireTypeLen)
	if err != nil {
		t.Fatal(err)
	}

	if string(field.Tag) != expected.String() {
		t.Fatalf("got tag %x, want %x", field.Tag, expected.Bytes())
	}
}

func TestNewFieldSliceOfPointers(t *testing.T) {
	type Item struct{}

	type Example struct {
		Values []*Item `protobuf:"bytes,10,rep,name=values,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Kind != reflect.Slice {
		t.Fatalf("got kind %v, want slice", field.Kind)
	}

	if field.Index != reflect.Struct {
		t.Fatalf("got index %v, want struct", field.Index)
	}

	if field.IndexType != TypeName(reflect.TypeOf(Item{})) {
		t.Fatalf("got index type %q, want %q", field.IndexType, TypeName(reflect.TypeOf(Item{})))
	}
}

func TestNewFieldArray(t *testing.T) {
	type Example struct {
		Values [3]string `protobuf:"bytes,11,opt,name=values,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Kind != reflect.Array {
		t.Fatalf("got kind %v, want array", field.Kind)
	}

	if field.Index != reflect.String {
		t.Fatalf("got index %v, want string", field.Index)
	}

	if field.IndexType != "string" {
		t.Fatalf("got index type %q, want string", field.IndexType)
	}

	if field.TypeName != "[3]string" {
		t.Fatalf("got type name %q, want [3]string", field.TypeName)
	}
}

func TestNewFieldMap(t *testing.T) {
	type Example struct {
		Values map[string]int `protobuf:"bytes,12,rep,name=values,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Kind != reflect.Map {
		t.Fatalf("got kind %v, want map", field.Kind)
	}

	if field.Key != reflect.String {
		t.Fatalf("got key %v, want string", field.Key)
	}

	if field.KeyType != "string" {
		t.Fatalf("got key type %q, want string", field.KeyType)
	}

	if field.Index != reflect.Int {
		t.Fatalf("got index %v, want int", field.Index)
	}

	if field.IndexType != "int" {
		t.Fatalf("got index type %q, want int", field.IndexType)
	}

	if field.TypeName != "map[string]int" {
		t.Fatalf("got type name %q, want map[string]int", field.TypeName)
	}

	if len(field.KeyTag) == 0 {
		t.Fatal("expected key tag")
	}

	if len(field.ValueTag) == 0 {
		t.Fatal("expected value tag")
	}
}

func TestNewFieldMapOfPointers(t *testing.T) {
	type Item struct{}

	type Example struct {
		Values map[int]*Item `protobuf:"bytes,13,rep,name=values,proto3" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Key != reflect.Int {
		t.Fatalf("got key %v, want int", field.Key)
	}

	if field.KeyType != "int" {
		t.Fatalf("got key type %q, want int", field.KeyType)
	}

	if field.Index != reflect.Struct {
		t.Fatalf("got index %v, want struct", field.Index)
	}

	if field.IndexType != TypeName(reflect.TypeOf(Item{})) {
		t.Fatalf("got index type %q, want %q", field.IndexType, TypeName(reflect.TypeOf(Item{})))
	}
}

func TestNewFieldWithoutProtobufTag(t *testing.T) {
	type Example struct {
		Value int `json:"value"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if field.Tags == nil {
		t.Fatal("expected Tags")
	}

	if field.Tags.Protobuf != nil {
		t.Fatal("expected nil protobuf metadata")
	}

	if len(field.Tag) != 0 {
		t.Fatalf("got tag %x, want empty", field.Tag)
	}
}

func TestNewFieldNestedPointer(t *testing.T) {
	type Example struct {
		Value **int `protobuf:"varint,14,opt,name=value,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	if !field.IsPointer {
		t.Fatal("expected pointer")
	}

	if field.Kind != reflect.Pointer {
		t.Fatalf("got kind %v, want pointer", field.Kind)
	}

	if !strings.Contains(field.TypeName, "int") {
		t.Fatalf("got type name %q", field.TypeName)
	}
}

func TestRegisterTypeForAndCapture(t *testing.T) {
	type Registered struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	RegisterTypeFor[Registered]()

	got := CaptureTypeFor[Registered]()
	if got == nil {
		t.Fatal("expected registered type")
	}

	if got.Name != TypeName(reflect.TypeFor[Registered]()) {
		t.Fatalf("unexpected name %q", got.Name)
	}

	if len(got.Fields) != 1 {
		t.Fatalf("got %d fields, want 1", len(got.Fields))
	}

	if got.FieldsIndexer[1] == nil {
		t.Fatal("expected field indexer entry")
	}
}

func TestRegisterTypeAndCapture(t *testing.T) {
	type Registered struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	RegisterType(reflect.TypeOf(Registered{}))

	got := CaptureType(reflect.TypeOf(Registered{}))
	if got == nil {
		t.Fatal("expected registered type")
	}
}

func TestRegisterTypeAs(t *testing.T) {
	type Registered struct {
		Value string `protobuf:"bytes,1,opt,name=value,proto3"`
	}

	const name = "custom.Registered"

	RegisterTypeAs[Registered](name)

	got := CaptureTypeByName(name)
	if got == nil {
		t.Fatal("expected registered type")
	}

	if got.Name != name {
		t.Fatalf("got name %q, want %q", got.Name, name)
	}
}

func TestCaptureTypeByNameMissing(t *testing.T) {
	if got := CaptureTypeByName("does.not.exist"); got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}

func TestRegisterTypeIgnoresNonStruct(t *testing.T) {
	const name = "test-non-struct-int"

	RegisterTypeAs[int](name)

	if got := CaptureTypeByName(name); got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}

func TestRegisterTypeDuplicate(t *testing.T) {
	type Duplicate struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	RegisterType(reflect.TypeOf(Duplicate{}))

	name := TypeName(reflect.TypeOf(Duplicate{}))
	first := CaptureTypeByName(name)

	RegisterType(reflect.TypeOf(Duplicate{}))
	second := CaptureTypeByName(name)

	if first == nil || second == nil {
		t.Fatal("expected registered type")
	}

	if first != second {
		t.Fatal("duplicate registration replaced existing type")
	}
}

func TestRegisterTypeFieldSorting(t *testing.T) {
	type Sorted struct {
		Third   int    `protobuf:"varint,30,opt,name=third,proto3"`
		First   int    `protobuf:"varint,10,opt,name=first,proto3"`
		Ignored string `json:"ignored"`
		Second  int    `protobuf:"varint,20,opt,name=second,proto3"`
	}

	RegisterTypeFor[Sorted]()

	got := CaptureTypeFor[Sorted]()

	if len(got.Fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(got.Fields))
	}

	want := []int{10, 20, 30}

	for i, field := range got.Fields {
		if field.Tags.Protobuf.FieldNum != want[i] {
			t.Fatalf("field %d: got %d, want %d", i, field.Tags.Protobuf.FieldNum, want[i])
		}
	}

	for _, n := range want {
		if got.FieldsIndexer[n] == nil {
			t.Fatalf("missing field index %d", n)
		}
	}
}

func TestRegisterTypeRegistersNestedFields(t *testing.T) {
	type Child struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	type Parent struct {
		Child Child `protobuf:"bytes,1,opt,name=child,proto3"`
	}

	RegisterTypeFor[Parent]()

	if CaptureTypeFor[Parent]() == nil {
		t.Fatal("expected parent")
	}

	if CaptureTypeFor[Child]() == nil {
		t.Fatal("expected child")
	}
}

func TestRegisterTypeRegistersPointerFieldType(t *testing.T) {
	type Child struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	type Parent struct {
		Child *Child `protobuf:"bytes,1,opt,name=child,proto3"`
	}

	RegisterTypeFor[Parent]()

	if CaptureTypeFor[Child]() == nil {
		t.Fatal("expected child")
	}
}

func TestRegisterTypeRegistersSliceFieldType(t *testing.T) {
	type Child struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	type Parent struct {
		Children []Child `protobuf:"bytes,1,rep,name=children,proto3"`
	}

	RegisterTypeFor[Parent]()

	if CaptureTypeFor[Child]() == nil {
		t.Fatal("expected child")
	}
}

func TestRegisterTypeRegistersMapFieldType(t *testing.T) {
	type Child struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	type Parent struct {
		Children map[string]Child `protobuf:"bytes,1,rep,name=children,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	}

	RegisterTypeFor[Parent]()

	if CaptureTypeFor[Child]() == nil {
		t.Fatal("expected child")
	}
}

func TestRegisterTypeKeepsOnlyProtobufFields(t *testing.T) {
	type Example struct {
		Included int    `protobuf:"varint,2,opt,name=included,proto3"`
		Ignored  string `json:"ignored"`
	}

	RegisterTypeFor[Example]()

	got := CaptureTypeFor[Example]()

	if len(got.Fields) != 1 {
		t.Fatalf("got %d fields, want 1", len(got.Fields))
	}

	if got.Fields[0].Name != "Included" {
		t.Fatalf("got %q, want Included", got.Fields[0].Name)
	}
}

func TestNewFieldTagEncoding(t *testing.T) {
	type Example struct {
		Value int `protobuf:"varint,300,opt,name=value,proto3" protobuf_key:"fixed64,1,opt,name=key" protobuf_val:"fixed32,2,opt,name=value"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	tag, err := pdk.TagEncode(300, WireTypeVarint)
	if err != nil {
		t.Fatal(err)
	}

	if string(field.Tag) != tag.String() {
		t.Fatalf("got tag %x, want %x", field.Tag, tag.Bytes())
	}

	keyTag, err := pdk.TagEncode(1, WireTypeI64)
	if err != nil {
		t.Fatal(err)
	}

	if string(field.KeyTag) != keyTag.String() {
		t.Fatalf("got key tag %x, want %x", field.KeyTag, keyTag.Bytes())
	}

	valueTag, err := pdk.TagEncode(2, WireTypeI32)
	if err != nil {
		t.Fatal(err)
	}

	if string(field.ValueTag) != valueTag.String() {
		t.Fatalf("got value tag %x, want %x", field.ValueTag, valueTag.Bytes())
	}
}

func TestNewFieldSliceOverridesWireType(t *testing.T) {
	type Example struct {
		Value []int `protobuf:"varint,20,rep,name=value,proto3"`
	}

	field := newField(reflect.TypeOf(Example{}).Field(0))

	want, err := pdk.TagEncode(20, WireTypeLen)
	if err != nil {
		t.Fatal(err)
	}

	if string(field.Tag) != want.String() {
		t.Fatalf("got tag %x, want %x", field.Tag, want.Bytes())
	}
}

func TestBuiltInTypesRegistered(t *testing.T) {
	tests := []reflect.Type{
		reflect.TypeOf(Tags{}),
		reflect.TypeOf(ProtobufInfo{}),
		reflect.TypeOf(Field{}),
		reflect.TypeOf(Type{}),
		reflect.TypeOf(Module{}),
	}

	for _, typ := range tests {
		t.Run(TypeName(typ), func(t *testing.T) {
			if CaptureType(typ) == nil {
				t.Fatalf("type %q is not registered", TypeName(typ))
			}
		})
	}
}
