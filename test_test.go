package protolizer

import (
	"testing"
)

func TestAll(t *testing.T) {

	test := new(Test)
	test.Id = 1
	test.Name = "Hello"
	test.Numbers = []int32{1, 2, 3}
	test.Values = []string{"ok", "then"}
	test.Data = map[int32]string{
		1: "something",
		2: "another",
	}
	modules, err := ExportModule[Type]()
	_ = modules

	xxxxzzzx, err := ImportModule(modules)

	_ = xxxxzzzx
	RegisterTypeFor[Test]()
	xxz := CaptureTypeFor[Test]()

	_ = xxz
	xxx, err := ExportType[Test]()
	_ = xxx

	typ, err := ImportType(xxx)
	_ = typ

	xxzzxxz, err := UnmarshalAnonymous("github.com/vedadiyan/protolizer.Type", xxx)

	_ = xxzzxxz
	bytes, err := Marshal(test)
	if err != nil {
		t.FailNow()
	}

	test2 := new(Test)
	if err := Unmarshal(bytes, test2); err != nil {
		t.FailNow()
	}

	out, err := UnmarshalAnonymous("github.com/vedadiyan/protolizer.Test", bytes)
	if err != nil {
		t.FailNow()
	}
	_ = out
}
