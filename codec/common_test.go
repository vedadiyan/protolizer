package codec

import (
	"reflect"
	"testing"

	users "github.com/vedadiyan/protolizer/test"
)

type User struct {
	FirstName string `protobuf:"bytes,1,opt,name=first_name,proto3" json:"first_name,omitempty"`
	LastName  string `protobuf:"bytes,2,opt,name=last_name,proto3" json:"last_name,omitempty"`
}

func TestSoFar(t *testing.T) {
	u := new(users.User)
	u.FirstName = "test"
	typ := RegisterType(reflect.TypeOf(new(users.User)))

	data, err := typ.Encode(reflect.ValueOf(u))
	if err != nil {
		t.FailNow()
	}
	u2 := new(users.User)
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
