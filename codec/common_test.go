package codec

import (
	"reflect"
	"sort"
	"testing"
)

type User struct {
	FirstName string `protobuf:"bytes,1,opt,name=first_name,proto3" json:"first_name,omitempty"`
	LastName  string `protobuf:"bytes,2,opt,name=last_name,proto3" json:"last_name,omitempty"`
}

func TestSoFar(t *testing.T) {
	u := new(User)
	u.FirstName = "test"
	typ := RegisterType(reflect.TypeOf(new(User)))
	sort.Slice(typ.Fields, func(i, j int) bool {
		return typ.Fields[i].Tags.Protobuf.Name < typ.Fields[j].Tags.Protobuf.Name
	})

	data, err := typ.Encode(reflect.ValueOf(u))
	if err != nil {
		t.FailNow()
	}
	u2 := new(User)
	pos, err := typ.Decode(reflect.ValueOf(u2), data, 0)
	if err != nil {
		t.FailNow()
	}
	_ = pos
	// v := reflect.ValueOf(u).Elem().FieldByName("FirstName")

	// bytes, _ := typ.Fields[1].Encode(v)
	// v.SetString("")
	// typ.Fields[1].Decode(bytes, v, 0)

	// fmt.Println(v.String())
}
