package protolizer

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	users "github.com/vedadiyan/protolizer/test"
)

func TestSoFar(t *testing.T) {
	u := new(users.User)
	u.FirstName = "test"
	typ := RegisterType(reflect.TypeOf(new(users.User)))
	sort.Slice(typ.Fields, func(i, j int) bool {
		return typ.Fields[i].Tags.Protobuf.Name < typ.Fields[j].Tags.Protobuf.Name
	})
	v := reflect.ValueOf(u).Elem().FieldByName("FirstName")

	bytes, _ := typ.Fields[1].Encode(v)
	v.SetString("")
	typ.Fields[1].Decode(bytes, v, 0)

	fmt.Println(v.String())
}
