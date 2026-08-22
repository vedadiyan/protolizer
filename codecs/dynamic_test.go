package codecs

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/vedadiyan/protolizer/metadata"
	"github.com/vedadiyan/protolizer/pdk"
)

type dynamicTestSimplePerson struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	Age  int    `protobuf:"varint,2,opt,name=age,proto3" json:"age"`
	Id   uint64 `protobuf:"varint,3,opt,name=id,proto3" json:"id"`
}

type dynamicTestComplexMessage struct {
	Id        uint64            `protobuf:"varint,1,opt,name=id,proto3" json:"id"`
	Name      string            `protobuf:"bytes,2,opt,name=name,proto3" json:"name"`
	Email     string            `protobuf:"bytes,3,opt,name=email,proto3" json:"email"`
	Score     float64           `protobuf:"fixed64,4,opt,name=score,proto3" json:"score"`
	IsActive  bool              `protobuf:"varint,5,opt,name=is_active,json=isActive,proto3" json:"isActive"`
	Tags      []string          `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags"`
	Numbers   []int             `protobuf:"varint,7,rep,packed,name=numbers,proto3" json:"numbers"`
	Metadata  map[string]string `protobuf:"bytes,8,rep,name=metadata,proto3" json:"metadata" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Timestamp int64             `protobuf:"varint,9,opt,name=timestamp,proto3" json:"timestamp"`
}

type dynamicTestAddressInfo struct {
	Street  string `protobuf:"bytes,1,opt,name=street,proto3" json:"street"`
	City    string `protobuf:"bytes,2,opt,name=city,proto3" json:"city"`
	Zipcode string `protobuf:"bytes,3,opt,name=zipcode,proto3" json:"zipcode"`
	Country string `protobuf:"bytes,4,opt,name=country,proto3" json:"country"`
}

type dynamicTestExtraData struct {
	Notes    string    `protobuf:"bytes,1,opt,name=notes,proto3" json:"notes"`
	Priority int       `protobuf:"varint,2,opt,name=priority,proto3" json:"priority"`
	Flags    []bool    `protobuf:"varint,3,rep,packed,name=flags,proto3" json:"flags"`
	Config   []float64 `protobuf:"fixed64,4,rep,packed,name=config,proto3" json:"config"`
}

type dynamicTestNestedMessage struct {
	Person      *dynamicTestSimplePerson  `protobuf:"bytes,1,opt,name=person,proto3" json:"person"`
	Address     *dynamicTestAddressInfo   `protobuf:"bytes,2,opt,name=address,proto3" json:"address"`
	Phones      []string                  `protobuf:"bytes,3,rep,name=phones,proto3" json:"phones"`
	Extra       *dynamicTestExtraData     `protobuf:"bytes,4,opt,name=extra,proto3" json:"extra"`
	PersonArray []dynamicTestSimplePerson `protobuf:"bytes,5,rep,name=personArray,proto3" json:"personArray"`
}

type dynamicTestNilPointer struct {
	Person *dynamicTestSimplePerson `protobuf:"bytes,1,opt,name=person,proto3"`
}

type dynamicTestBytes struct {
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3"`
}

func registerDynamicTestTypes(d *Dynamic) {
	d.Register(reflect.TypeFor[dynamicTestSimplePerson]())
	d.Register(reflect.TypeFor[dynamicTestComplexMessage]())
	d.Register(reflect.TypeFor[dynamicTestNestedMessage]())
	d.Register(reflect.TypeFor[dynamicTestAddressInfo]())
	d.Register(reflect.TypeFor[dynamicTestExtraData]())
	d.Register(reflect.TypeFor[dynamicTestNilPointer]())
	d.Register(reflect.TypeFor[dynamicTestBytes]())
}

func TestDynamicRegister(t *testing.T) {
	d := NewDynamic()
	registerDynamicTestTypes(d)

	types := []reflect.Type{
		reflect.TypeFor[dynamicTestSimplePerson](),
		reflect.TypeFor[dynamicTestComplexMessage](),
		reflect.TypeFor[dynamicTestNestedMessage](),
		reflect.TypeFor[dynamicTestAddressInfo](),
		reflect.TypeFor[dynamicTestExtraData](),
	}

	for _, typ := range types {
		name := metadata.TypeName(typ)

		if _, ok := d._builtEncoders[name]; !ok {
			t.Fatalf("encoder for %s was not registered", name)
		}

		if _, ok := d._builtDecoders[name]; !ok {
			t.Fatalf("decoder for %s was not registered", name)
		}
	}
}

func TestDynamicMarshalUnregistered(t *testing.T) {
	d := NewDynamic()

	_, err := d.Marshal(&dynamicTestSimplePerson{})
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestDynamicUnmarshalUnregistered(t *testing.T) {
	d := NewDynamic()

	var out dynamicTestSimplePerson
	err := d.Unmarshal(nil, &out)
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestDynamicMarshalSimple(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestSimplePerson]())

	in := &dynamicTestSimplePerson{
		Name: "John Doe",
		Age:  30,
		Id:   12345,
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("expected encoded data")
	}
}

func TestDynamicRoundTripSimple(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestSimplePerson]())

	in := &dynamicTestSimplePerson{
		Name: "John Doe",
		Age:  30,
		Id:   12345,
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out dynamicTestSimplePerson
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(out, *in) {
		t.Fatalf("got %#v, want %#v", out, *in)
	}
}

func TestDynamicMarshalComplex(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestComplexMessage]())

	in := &dynamicTestComplexMessage{
		Id:        67890,
		Name:      "Complex Test Message",
		Email:     "test@example.com",
		Score:     95.5,
		IsActive:  true,
		Tags:      []string{"important", "urgent", "customer", "vip"},
		Numbers:   []int{1, 2, 3, 4, 5, 10, 20, 30, 40, 50},
		Metadata:  map[string]string{"source": "api", "version": "1.2.3"},
		Timestamp: 123456789,
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("expected encoded data")
	}
}

func TestDynamicRoundTripComplex(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestComplexMessage]())

	in := &dynamicTestComplexMessage{
		Id:        67890,
		Name:      "Complex Test Message",
		Email:     "test@example.com",
		Score:     95.5,
		IsActive:  true,
		Tags:      []string{"important", "urgent", "customer", "vip"},
		Numbers:   []int{1, 2, 3, 4, 5, 10, 20, 30},
		Metadata:  map[string]string{"source": "api", "version": "1.2.3"},
		Timestamp: 123456789,
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out dynamicTestComplexMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Id != in.Id {
		t.Fatalf("Id = %d, want %d", out.Id, in.Id)
	}
	if out.Name != in.Name {
		t.Fatalf("Name = %q, want %q", out.Name, in.Name)
	}
	if out.Email != in.Email {
		t.Fatalf("Email = %q, want %q", out.Email, in.Email)
	}
	if out.Score != in.Score {
		t.Fatalf("Score = %v, want %v", out.Score, in.Score)
	}
	if out.IsActive != in.IsActive {
		t.Fatalf("IsActive = %v, want %v", out.IsActive, in.IsActive)
	}
	if !reflect.DeepEqual(out.Tags, in.Tags) {
		t.Fatalf("Tags = %#v, want %#v", out.Tags, in.Tags)
	}
	if !reflect.DeepEqual(out.Numbers, in.Numbers) {
		t.Fatalf("Numbers = %#v, want %#v", out.Numbers, in.Numbers)
	}
	if !reflect.DeepEqual(out.Metadata, in.Metadata) {
		t.Fatalf("Metadata = %#v, want %#v", out.Metadata, in.Metadata)
	}
	if out.Timestamp != in.Timestamp {
		t.Fatalf("Timestamp = %d, want %d", out.Timestamp, in.Timestamp)
	}
}

func TestDynamicRoundTripNested(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestNestedMessage]())

	in := &dynamicTestNestedMessage{
		Person: &dynamicTestSimplePerson{
			Name: "Jane Smith",
			Age:  25,
			Id:   54321,
		},
		Address: &dynamicTestAddressInfo{
			Street:  "123 Main Street",
			City:    "New York",
			Zipcode: "10001",
			Country: "USA",
		},
		Phones: []string{"+1-555-1234", "+1-555-5678"},
		Extra: &dynamicTestExtraData{
			Notes:    "test",
			Priority: 1,
			Flags:    []bool{true, false, true},
			Config:   []float64{30.5, 3.0, 100.0},
		},
		PersonArray: []dynamicTestSimplePerson{
			{Name: "Ok", Age: 100, Id: 12345},
		},
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out dynamicTestNestedMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Person == nil {
		t.Fatal("Person is nil")
	}
	if out.Address == nil {
		t.Fatal("Address is nil")
	}
	if out.Extra == nil {
		t.Fatal("Extra is nil")
	}

	if !reflect.DeepEqual(out, *in) {
		t.Fatalf("got %#v, want %#v", out, *in)
	}
}

func TestDynamicBoolField(t *testing.T) {
	type BoolMessage struct {
		Value bool `protobuf:"varint,1,opt,name=value,proto3"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[BoolMessage]())

	in := &BoolMessage{Value: true}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out BoolMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Value != in.Value {
		t.Fatalf("Value = %v, want %v", out.Value, in.Value)
	}
}

func TestDynamicFloat32Field(t *testing.T) {
	type FloatMessage struct {
		Value float32 `protobuf:"fixed32,1,opt,name=value,proto3"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[FloatMessage]())

	in := &FloatMessage{Value: 95.5}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out FloatMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Value != in.Value {
		t.Fatalf("Value = %v, want %v", out.Value, in.Value)
	}
}

func TestDynamicFloat64Field(t *testing.T) {
	type FloatMessage struct {
		Value float64 `protobuf:"fixed64,1,opt,name=value,proto3"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[FloatMessage]())

	in := &FloatMessage{Value: 95.5}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out FloatMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Value != in.Value {
		t.Fatalf("Value = %v, want %v", out.Value, in.Value)
	}
}

func TestDynamicStringField(t *testing.T) {
	type StringMessage struct {
		Value string `protobuf:"bytes,1,opt,name=value,proto3"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[StringMessage]())

	in := &StringMessage{Value: "hello world"}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out StringMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Value != in.Value {
		t.Fatalf("Value = %q, want %q", out.Value, in.Value)
	}
}

func TestDynamicBytesField(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestBytes]())

	in := &dynamicTestBytes{Value: []byte("hello")}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out dynamicTestBytes
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out.Value, in.Value) {
		t.Fatalf("Value = %q, want %q", out.Value, in.Value)
	}
}

func TestDynamicRepeatedStrings(t *testing.T) {
	type RepeatedMessage struct {
		Values []string `protobuf:"bytes,1,rep,name=values,proto3"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[RepeatedMessage]())

	in := &RepeatedMessage{
		Values: []string{"one", "two", "three"},
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out RepeatedMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(out.Values, in.Values) {
		t.Fatalf("Values = %#v, want %#v", out.Values, in.Values)
	}
}

func TestDynamicPackedInts(t *testing.T) {
	type PackedMessage struct {
		Values []int `protobuf:"varint,1,rep,packed,name=values,proto3"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[PackedMessage]())

	in := &PackedMessage{
		Values: []int{1, 2, 3, 10, 20, 30},
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out PackedMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(out.Values, in.Values) {
		t.Fatalf("Values = %#v, want %#v", out.Values, in.Values)
	}
}

func TestDynamicMap(t *testing.T) {
	type MapMessage struct {
		Values map[string]string `protobuf:"bytes,1,rep,name=values,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	}

	d := NewDynamic()
	d.Register(reflect.TypeFor[MapMessage]())

	in := &MapMessage{
		Values: map[string]string{
			"one":   "1",
			"two":   "2",
			"three": "3",
		},
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out MapMessage
	if err := d.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(out.Values, in.Values) {
		t.Fatalf("Values = %#v, want %#v", out.Values, in.Values)
	}
}

func TestDynamicNilPointerField(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestNilPointer]())

	in := &dynamicTestNilPointer{}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 0 {
		t.Fatalf("nil pointer encoded as %x", data)
	}
}

func TestDynamicUnknownField(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestSimplePerson]())

	var data bytes.Buffer

	tag, err := pdk.TagEncode(99, pdk.WireTypeVarint)
	if err != nil {
		t.Fatal(err)
	}

	data.Write(tag.Bytes())

	value := pdk.UvarintEncode(123)
	data.Write(value.Bytes())

	var out dynamicTestSimplePerson
	if err := d.Unmarshal(data.Bytes(), &out); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestDynamicInvalidData(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestSimplePerson]())

	var out dynamicTestSimplePerson

	if err := d.Unmarshal([]byte{0xff, 0xff, 0xff, 0xff, 0xff}, &out); err == nil {
		t.Fatal("expected error for invalid data")
	}
}

func TestDynamicUnmarshalNonPointer(t *testing.T) {
	d := NewDynamic()
	d.Register(reflect.TypeFor[dynamicTestSimplePerson]())

	in := &dynamicTestSimplePerson{
		Name: "John",
		Age:  30,
		Id:   123,
	}

	data, err := d.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	err = d.Unmarshal(data, *in)
	if err == nil {
		t.Fatal("expected error for non-pointer destination")
	}
}
