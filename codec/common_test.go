package codec

import (
	"fmt"
	"reflect"
	"testing"

	users "github.com/vedadiyan/protolizer/test"
	"google.golang.org/protobuf/proto"
)

type User struct {
	FirstName string `protobuf:"bytes,1,opt,name=first_name,proto3" json:"first_name,omitempty"`
	LastName  string `protobuf:"bytes,2,opt,name=last_name,proto3" json:"last_name,omitempty"`
}

func TestSoFar(t *testing.T) {
	u := new(users.General)
	u.Boolean = true
	u.Bytes = []byte{1, 2, 3, 4, 5}
	u.Dbl = 100
	u.Enum = users.TestEnum_TEST
	u.Flt = 200
	// u.Fx32 = 1000
	// u.Fx64 = 10000
	u.I32 = 2
	u.I64 = 1000000
	u.Str = "ok"
	u.Ui32 = 3900
	u.Ui64 = 5678
	u.Sfixed32 = 3
	u.Sfixed64 = 5
	u.Map = make(map[int32]string)
	u.Map[1] = "Hello World"
	typ := RegisterType(reflect.TypeOf(new(users.General)))

	data, err := typ.Encode(reflect.ValueOf(u))
	if err != nil {
		t.FailNow()
	}
	data2, err := proto.Marshal(u)
	_ = data

	fmt.Println(data)
	fmt.Println(data2)

	u2 := new(users.General)
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

func TestSoFar2(t *testing.T) {
	u := new(users.Fixed)
	u.Fx32 = 1000
	u.Fx64 = 10000

	typ := RegisterType(reflect.TypeOf(new(users.Fixed)))

	data, err := typ.Encode(reflect.ValueOf(u))
	if err != nil {
		t.FailNow()
	}
	data2, err := proto.Marshal(u)
	_ = data

	fmt.Println(data)
	fmt.Println(data2)

	u2 := new(users.Fixed)
	pos, err := typ.Decode(reflect.ValueOf(u2), data2, 0)
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
