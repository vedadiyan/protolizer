package protolizer

import (
	"encoding/hex"
	"reflect"
	"testing"
)

type Address struct {
	Street string `protobuf:"bytes,1,opt,name=street,proto3" json:"street,omitempty"`
	City   string `protobuf:"bytes,2,opt,name=city,proto3" json:"city,omitempty"`
	Zip    int32  `protobuf:"varint,3,opt,name=zip,proto3" json:"zip,omitempty"`
}

type Person struct {
	Id        int64             `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string            `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Active    bool              `protobuf:"varint,3,opt,name=active,proto3" json:"active,omitempty"`
	Rating    float32           `protobuf:"fixed32,4,opt,name=rating,proto3" json:"rating,omitempty"`
	Balance   uint64            `protobuf:"fixed64,5,opt,name=balance,proto3" json:"balance,omitempty"`
	Data      []byte            `protobuf:"bytes,6,opt,name=data,proto3" json:"data,omitempty"`
	Tags      []string          `protobuf:"bytes,7,repeated,name=tags,proto3" json:"tags,omitempty"`
	Scores    []int64           `protobuf:"varint,8,repeated,name=scores,proto3" json:"scores,omitempty"`
	Flags     []bool            `protobuf:"varint,9,repeated,name=flags,proto3" json:"flags,omitempty"`
	Labels    map[string]string `protobuf:"bytes,10,map,name=labels,proto3" json:"labels,omitempty"`
	Addresses []Address         `protobuf:"bytes,11,repeated,name=addresses,proto3" json:"addresses,omitempty"`
	MainAddr  *Address          `protobuf:"bytes,12,opt,name=main_addr,proto3,oneof" json:"main_addr,omitempty"`
}

func TestSerializeDeserializePerson(t *testing.T) {
	p := &Person{
		Id:      42,
		Name:    "Alice",
		Active:  true,
		Rating:  4.5,
		Balance: 100000,
		Data:    []byte{0xde, 0xad, 0xbe, 0xef},
		Tags:    []string{"golang", "protobuf", "test"},
		Scores:  []int64{100, 200, 300},
		Flags:   []bool{true, false, true},
		Labels: map[string]string{
			"env":  "dev",
			"role": "tester",
		},
		Addresses: []Address{
			{Street: "123 Main St", City: "Springfield", Zip: 12345},
			{Street: "456 Side St", City: "Shelbyville", Zip: 54321},
		},
		MainAddr: &Address{
			Street: "789 Central Ave",
			City:   "Capital City",
			Zip:    99999,
		},
	}

	RegisterType(reflect.TypeOf(new(Person)))
	RegisterType(reflect.TypeOf(new(Address)))

	b, err := Marshal(p)
	if err != nil {
		t.Fatalf("SerializeStruct failed: %v", err)
	}
	t.Logf("Encoded Person protobuf: %s", hex.EncodeToString(b))

	var p2 Person
	err = Unmarshal(b, &p2)
	if err != nil {
		t.Fatalf("DeserializeStruct failed: %v", err)
	}

	// basic scalar checks
	if p.Id != p2.Id || p.Name != p2.Name || p.Active != p2.Active || p.Rating != p2.Rating || p.Balance != p2.Balance {
		t.Errorf("scalar mismatch\noriginal: %+v\ndecoded: %+v", p, p2)
	}

	// bytes
	if len(p.Data) != len(p2.Data) {
		t.Errorf("data mismatch %v vs %v", p.Data, p2.Data)
	}

	// repeated string
	if len(p.Tags) != len(p2.Tags) {
		t.Errorf("tags length mismatch: %v vs %v", p.Tags, p2.Tags)
	}
	for i := range p.Tags {
		if p.Tags[i] != p2.Tags[i] {
			t.Errorf("tag mismatch at %d: %s vs %s", i, p.Tags[i], p2.Tags[i])
		}
	}

	// repeated int64
	if len(p.Scores) != len(p2.Scores) {
		t.Errorf("scores mismatch: %v vs %v", p.Scores, p2.Scores)
	}

	// repeated bool
	if len(p.Flags) != len(p2.Flags) {
		t.Errorf("flags mismatch: %v vs %v", p.Flags, p2.Flags)
	}

	// map
	for k, v := range p.Labels {
		if p2.Labels[k] != v {
			t.Errorf("map mismatch at %s: %s vs %s", k, v, p2.Labels[k])
		}
	}

	// repeated struct
	if len(p.Addresses) != len(p2.Addresses) {
		t.Errorf("addresses length mismatch: %v vs %v", p.Addresses, p2.Addresses)
	}

	// pointer struct
	if p.MainAddr == nil || p2.MainAddr == nil || p.MainAddr.Street != p2.MainAddr.Street {
		t.Errorf("main addr mismatch: %+v vs %+v", p.MainAddr, p2.MainAddr)
	}
}
