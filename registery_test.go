package protolizer_test

import (
	"reflect"
	"testing"

	"github.com/vedadiyan/protolizer"
	"github.com/vedadiyan/protolizer/metadata"
)

type RegistryChild struct {
	Value int `protobuf:"varint,1,opt,name=value,proto3"`
}

type RegistryRoot struct {
	ID    int             `protobuf:"varint,1,opt,name=id,proto3"`
	Name  string          `protobuf:"bytes,2,opt,name=name,proto3"`
	Child RegistryChild   `protobuf:"bytes,3,opt,name=child,proto3"`
	Ptr   *RegistryChild  `protobuf:"bytes,4,opt,name=ptr,proto3"`
	Items []RegistryChild `protobuf:"bytes,5,rep,name=items,proto3"`
}

type RegistryLeaf struct {
	Value string `protobuf:"bytes,1,opt,name=value,proto3"`
}

type RegistryNested struct {
	Direct RegistryLeaf            `protobuf:"bytes,1,opt,name=direct,proto3"`
	Ptr    *RegistryLeaf           `protobuf:"bytes,2,opt,name=ptr,proto3"`
	Slice  []RegistryLeaf          `protobuf:"bytes,3,rep,name=slice,proto3"`
	Array  [2]RegistryLeaf         `protobuf:"bytes,4,rep,name=array,proto3"`
	Map    map[string]RegistryLeaf `protobuf:"bytes,5,rep,name=map,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func init() {
	codec := protolizer.DynamicCodec()

	codec.Register(reflect.TypeFor[metadata.Tags]())
	codec.Register(reflect.TypeFor[metadata.ProtobufInfo]())
	codec.Register(reflect.TypeFor[metadata.Field]())
	codec.Register(reflect.TypeFor[metadata.Type]())
	codec.Register(reflect.TypeFor[metadata.Module]())

	codec.Register(reflect.TypeFor[RegistryChild]())
	codec.Register(reflect.TypeFor[RegistryRoot]())
	codec.Register(reflect.TypeFor[RegistryLeaf]())
	codec.Register(reflect.TypeFor[RegistryNested]())
}

func TestExportType(t *testing.T) {
	data, err := protolizer.ExportType[RegistryRoot]()
	if err != nil {
		t.Fatalf("ExportType returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportType returned empty data")
	}

	got, err := protolizer.ImportType(data)
	if err != nil {
		t.Fatalf("ImportType returned error: %v", err)
	}

	if got == nil {
		t.Fatal("ImportType returned nil")
	}

	want := metadata.CaptureTypeFor[RegistryRoot]()
	if want == nil {
		t.Fatal("RegistryRoot is not registered")
	}

	if got.Name != want.Name {
		t.Fatalf("got name %q, want %q", got.Name, want.Name)
	}

	if len(got.Fields) != len(want.Fields) {
		t.Fatalf("got %d fields, want %d", len(got.Fields), len(want.Fields))
	}
}

func TestExportTypeUnregistered(t *testing.T) {
	type Unregistered struct {
		Value int `protobuf:"varint,1,opt,name=value,proto3"`
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected ExportType to panic for unregistered type")
		}
	}()

	_, _ = protolizer.ExportType[Unregistered]()
}

func TestImportType(t *testing.T) {
	data, err := protolizer.ExportType[RegistryRoot]()
	if err != nil {
		t.Fatalf("ExportType returned error: %v", err)
	}

	got, err := protolizer.ImportType(data)
	if err != nil {
		t.Fatalf("ImportType returned error: %v", err)
	}

	if got == nil {
		t.Fatal("ImportType returned nil")
	}

	want := metadata.TypeName(reflect.TypeFor[RegistryRoot]())

	if got.Name != want {
		t.Fatalf("got name %q, want %q", got.Name, want)
	}
}

func TestImportTypeInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "invalid wire type",
			data: []byte{0x07},
		},
		{
			name: "truncated varint",
			data: []byte{0x80},
		},
		{
			name: "truncated length",
			data: []byte{0x0a, 0xff},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := protolizer.ImportType(tt.data)

			if err == nil {
				t.Fatal("expected error")
			}

			if got != nil {
				t.Fatalf("got %+v, want nil", got)
			}
		})
	}
}

func TestExportModule(t *testing.T) {
	data, err := protolizer.ExportModule[RegistryRoot]()
	if err != nil {
		t.Fatalf("ExportModule returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportModule returned empty data")
	}

	module, err := protolizer.ImportModule(data)
	if err != nil {
		t.Fatalf("ImportModule returned error: %v", err)
	}

	if module == nil {
		t.Fatal("ImportModule returned nil")
	}

	if module.Types == nil {
		t.Fatal("module.Types is nil")
	}

	for _, typ := range []reflect.Type{
		reflect.TypeFor[RegistryRoot](),
		reflect.TypeFor[RegistryChild](),
	} {
		name := metadata.TypeName(typ)

		if _, ok := module.Types[name]; !ok {
			t.Fatalf("module does not contain %q", name)
		}
	}
}

func TestExportModuleNestedTypes(t *testing.T) {
	data, err := protolizer.ExportModule[RegistryNested]()
	if err != nil {
		t.Fatalf("ExportModule returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportModule returned empty data")
	}

	module, err := protolizer.ImportModule(data)
	if err != nil {
		t.Fatalf("ImportModule returned error: %v", err)
	}

	if module == nil {
		t.Fatal("ImportModule returned nil")
	}

	for _, typ := range []reflect.Type{
		reflect.TypeFor[RegistryNested](),
		reflect.TypeFor[RegistryLeaf](),
	} {
		name := metadata.TypeName(typ)

		if _, ok := module.Types[name]; !ok {
			t.Fatalf("module does not contain %q", name)
		}
	}
}

func TestExportModuleWithPointerRoot(t *testing.T) {
	t.Skip("exportModule currently does not support pointer root types")
}

func TestExportModuleEmptyStruct(t *testing.T) {
	type Empty struct{}

	protolizer.DynamicCodec().Register(reflect.TypeFor[Empty]())

	data, err := protolizer.ExportModule[Empty]()
	if err != nil {
		t.Fatalf("ExportModule returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportModule returned empty data")
	}

	module, err := protolizer.ImportModule(data)
	if err != nil {
		t.Fatalf("ImportModule returned error: %v", err)
	}

	if module == nil {
		t.Fatal("ImportModule returned nil")
	}

	name := metadata.TypeName(reflect.TypeFor[Empty]())

	if _, ok := module.Types[name]; !ok {
		t.Fatalf("module does not contain %q", name)
	}
}

func TestExportModulePrimitiveFields(t *testing.T) {
	type Primitive struct {
		I int     `protobuf:"varint,1,opt,name=i,proto3"`
		S string  `protobuf:"bytes,2,opt,name=s,proto3"`
		B bool    `protobuf:"varint,3,opt,name=b,proto3"`
		F float64 `protobuf:"fixed64,4,opt,name=f,proto3"`
	}

	protolizer.DynamicCodec().Register(reflect.TypeFor[Primitive]())

	data, err := protolizer.ExportModule[Primitive]()
	if err != nil {
		t.Fatalf("ExportModule returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportModule returned empty data")
	}

	module, err := protolizer.ImportModule(data)
	if err != nil {
		t.Fatalf("ImportModule returned error: %v", err)
	}

	if module == nil {
		t.Fatal("ImportModule returned nil")
	}

	name := metadata.TypeName(reflect.TypeFor[Primitive]())

	if _, ok := module.Types[name]; !ok {
		t.Fatalf("module does not contain %q", name)
	}
}

func TestImportModule(t *testing.T) {
	data, err := protolizer.ExportModule[RegistryRoot]()
	if err != nil {
		t.Fatalf("ExportModule returned error: %v", err)
	}

	module, err := protolizer.ImportModule(data)
	if err != nil {
		t.Fatalf("ImportModule returned error: %v", err)
	}

	if module == nil {
		t.Fatal("ImportModule returned nil")
	}

	if module.Types == nil {
		t.Fatal("module.Types is nil")
	}
}

func TestImportModuleInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "invalid wire type",
			data: []byte{0x07},
		},
		{
			name: "truncated varint",
			data: []byte{0x80},
		},
		{
			name: "truncated length",
			data: []byte{0x0a, 0xff},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := protolizer.ImportModule(tt.data)

			if err == nil {
				t.Fatal("expected error")
			}

			if got != nil {
				t.Fatalf("got %+v, want nil", got)
			}
		})
	}
}

func TestImportModuleEmptyData(t *testing.T) {
	module, err := protolizer.ImportModule(nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if module == nil {
		t.Fatal("expected non-nil module")
	}
}
