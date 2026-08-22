package codecs

import (
	"reflect"
	"testing"

	"github.com/vedadiyan/protolizer/metadata"
)

type typelessTestMessage struct {
	Name      string            `protobuf:"bytes,1,opt,name=name,proto3"`
	Age       int               `protobuf:"varint,2,opt,name=age,proto3"`
	Active    bool              `protobuf:"varint,3,opt,name=active,proto3"`
	Score     float64           `protobuf:"fixed64,4,opt,name=score,proto3"`
	Tags      []string          `protobuf:"bytes,5,rep,name=tags,proto3"`
	Numbers   []int             `protobuf:"varint,6,rep,packed,name=numbers,proto3"`
	Metadata  map[string]string `protobuf:"bytes,7,rep,name=metadata,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Data      []byte            `protobuf:"bytes,8,opt,name=data,proto3"`
	Timestamp int64             `protobuf:"varint,9,opt,name=timestamp,proto3"`
}

type typelessUnsignedNumbersMessage struct {
	Uint   uint   `protobuf:"varint,1,opt,name=uint,proto3"`
	Uint8  uint8  `protobuf:"varint,2,opt,name=uint8,proto3"`
	Uint16 uint16 `protobuf:"varint,3,opt,name=uint16,proto3"`
	Uint32 uint32 `protobuf:"varint,4,opt,name=uint32,proto3"`
	Uint64 uint64 `protobuf:"varint,5,opt,name=uint64,proto3"`
}

type typelessFloat32Message struct {
	Value float32 `protobuf:"fixed32,1,opt,name=value,proto3"`
}

type typelessMissingFieldsMessage struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3"`
	Age  int    `protobuf:"varint,2,opt,name=age,proto3"`
}

type typelessNestedMessage struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3"`
}

type typelessRootMessage struct {
	Nested typelessNestedMessage `protobuf:"bytes,1,opt,name=nested,proto3"`
}

type typelessEmptyMessage struct{}

func newTypelessTestCodec() (*Typeless, *metadata.Type) {
	metadata.RegisterTypeFor[typelessTestMessage]()

	t := metadata.CaptureTypeFor[typelessTestMessage]()

	codec := NewTypeless()
	codec.Register(t)

	return codec, t
}

func TestTypelessMarshalUnmarshal(t *testing.T) {
	codec, typ := newTypelessTestCodec()

	input := map[string]any{
		"Name":      "John",
		"Age":       30,
		"Active":    true,
		"Score":     95.5,
		"Tags":      []string{"one", "two", "three"},
		"Numbers":   []int{1, 2, 3, 4},
		"Metadata":  map[string]string{"source": "api", "version": "1.0"},
		"Data":      []byte("hello"),
		"Timestamp": int64(123456),
	}

	data, err := codec.Marshal(input, typ)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	got, err := codec.Unmarshal(data, typ)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	out, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}

	if out["Name"] != "John" {
		t.Fatalf("Name = %#v, want John", out["Name"])
	}

	if out["Age"] != int64(30) {
		t.Fatalf("Age = %#v, want 30", out["Age"])
	}

	if out["Active"] != true {
		t.Fatalf("Active = %#v, want true", out["Active"])
	}

	if out["Score"] != float64(95.5) {
		t.Fatalf("Score = %#v, want 95.5", out["Score"])
	}

	if out["Timestamp"] != int64(123456) {
		t.Fatalf("Timestamp = %#v, want 123456", out["Timestamp"])
	}

	if !reflect.DeepEqual(out["Data"], []byte("hello")) {
		t.Fatalf("Data = %#v, want []byte(%q)", out["Data"], "hello")
	}

	tags, ok := out["Tags"].([]any)
	if !ok {
		t.Fatalf("Tags has type %T, want []any", out["Tags"])
	}

	if !reflect.DeepEqual(tags, []any{"one", "two", "three"}) {
		t.Fatalf("Tags = %#v", tags)
	}

	numbers, ok := out["Numbers"].([]any)
	if !ok {
		t.Fatalf("Numbers has type %T, want []any", out["Numbers"])
	}

	if !reflect.DeepEqual(numbers, []any{
		int64(1),
		int64(2),
		int64(3),
		int64(4),
	}) {
		t.Fatalf("Numbers = %#v", numbers)
	}

	meta, ok := out["Metadata"].(map[any]any)
	if !ok {
		t.Fatalf("Metadata has type %T, want map[any]any", out["Metadata"])
	}

	if meta["source"] != "api" {
		t.Fatalf("Metadata[source] = %#v", meta["source"])
	}

	if meta["version"] != "1.0" {
		t.Fatalf("Metadata[version] = %#v", meta["version"])
	}
}

func TestTypelessUnsignedNumbers(t *testing.T) {
	metadata.RegisterTypeFor[typelessUnsignedNumbersMessage]()
	typ := metadata.CaptureTypeFor[typelessUnsignedNumbersMessage]()

	codec := NewTypeless()
	codec.Register(typ)

	input := map[string]any{
		"Uint":   uint(1),
		"Uint8":  uint8(2),
		"Uint16": uint16(3),
		"Uint32": uint32(4),
		"Uint64": uint64(5),
	}

	data, err := codec.Marshal(input, typ)
	if err != nil {
		t.Fatal(err)
	}

	got, err := codec.Unmarshal(data, typ)
	if err != nil {
		t.Fatal(err)
	}

	out := got.(map[string]any)

	want := map[string]any{
		"Uint":   uint64(1),
		"Uint8":  uint64(2),
		"Uint16": uint64(3),
		"Uint32": uint64(4),
		"Uint64": uint64(5),
	}

	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %#v, want %#v", out, want)
	}
}

func TestTypelessFloat32(t *testing.T) {
	metadata.RegisterTypeFor[typelessFloat32Message]()
	typ := metadata.CaptureTypeFor[typelessFloat32Message]()

	codec := NewTypeless()
	codec.Register(typ)

	input := map[string]any{
		"Value": float32(123.456),
	}

	data, err := codec.Marshal(input, typ)
	if err != nil {
		t.Fatal(err)
	}

	got, err := codec.Unmarshal(data, typ)
	if err != nil {
		t.Fatal(err)
	}

	out := got.(map[string]any)

	if out["Value"] != float32(123.456) {
		t.Fatalf("got %v (%T), want %v", out["Value"], out["Value"], float32(123.456))
	}
}

func TestTypelessMissingFields(t *testing.T) {
	metadata.RegisterTypeFor[typelessMissingFieldsMessage]()
	typ := metadata.CaptureTypeFor[typelessMissingFieldsMessage]()

	codec := NewTypeless()
	codec.Register(typ)

	data, err := codec.Marshal(map[string]any{
		"Name": "only-name",
	}, typ)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	got, err := codec.Unmarshal(data, typ)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	out := got.(map[string]any)

	if out["Name"] != "only-name" {
		t.Fatalf("Name = %#v", out["Name"])
	}

	if _, ok := out["Age"]; ok {
		t.Fatal("Age should not be present")
	}
}

func TestTypelessUnknownType(t *testing.T) {
	codec := NewTypeless()

	typ := &metadata.Type{Name: "not_registered"}

	if _, err := codec.Marshal(map[string]any{}, typ); err == nil {
		t.Fatal("expected Marshal error for unregistered type")
	}

	if _, err := codec.Unmarshal(nil, typ); err == nil {
		t.Fatal("expected Unmarshal error for unregistered type")
	}
}

func TestTypelessStructUnsupported(t *testing.T) {
	metadata.RegisterTypeFor[typelessRootMessage]()

	typ := metadata.CaptureTypeFor[typelessRootMessage]()

	codec := NewTypeless()
	codec.Register(typ)

	_, err := codec.Marshal(map[string]any{
		"Nested": typelessNestedMessage{Name: "test"},
	}, typ)

	if err == nil {
		t.Fatal("expected struct unsupported error")
	}

	if err.Error() != "typeless coded does not support structs" {
		t.Fatalf("error = %q", err)
	}
}

func TestTypelessEmptyMessage(t *testing.T) {
	metadata.RegisterTypeFor[typelessEmptyMessage]()

	typ := metadata.CaptureTypeFor[typelessEmptyMessage]()

	codec := NewTypeless()
	codec.Register(typ)

	data, err := codec.Marshal(map[string]any{}, typ)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	if len(data) != 0 {
		t.Fatalf("expected empty data, got %v", data)
	}

	got, err := codec.Unmarshal(nil, typ)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	out, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got %T, want map[string]any", got)
	}

	if len(out) != 0 {
		t.Fatalf("expected empty map, got %#v", out)
	}
}
