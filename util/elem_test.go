package util

import (
	"reflect"
	"testing"
)

func TestGetElemenType(t *testing.T) {
	type MyStruct struct {
		Value string
	}
	type MyInt int

	intType := reflect.TypeOf(int(0))
	stringType := reflect.TypeOf("")
	structType := reflect.TypeOf(MyStruct{})
	myIntType := reflect.TypeOf(MyInt(0))

	tests := []struct {
		name string
		typ  reflect.Type
		want reflect.Type
	}{
		{"int", intType, intType},
		{"string", stringType, stringType},
		{"struct", structType, structType},
		{"named type", myIntType, myIntType},
		{"pointer", reflect.TypeOf((*int)(nil)), intType},
		{"double pointer", reflect.TypeOf((**int)(nil)), intType},
		{"triple pointer", reflect.TypeOf((***int)(nil)), intType},
		{"slice", reflect.TypeOf([]int(nil)), intType},
		{"nested slice", reflect.TypeOf([][]int(nil)), intType},
		{"array", reflect.TypeOf([5]int{}), intType},
		{"nested array", reflect.TypeOf([2][3]int{}), intType},
		{"map", reflect.TypeOf(map[string]int(nil)), intType},
		{"nested map", reflect.TypeOf(map[string][]int(nil)), intType},
		{"pointer slice", reflect.TypeOf([]*int(nil)), intType},
		{"slice pointer", reflect.TypeOf((*[]int)(nil)), intType},
		{
			"complex nested",
			reflect.TypeOf(map[string][]*[2]**MyInt{}),
			myIntType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetElemenType(tt.typ)

			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValue(t *testing.T) {
	type MyStruct struct {
		Value int
	}

	tests := []struct {
		name      string
		setup     func() reflect.Value
		wantType  reflect.Type
		wantValue interface{}
	}{
		{
			name: "int",
			setup: func() reflect.Value {
				return reflect.ValueOf(42)
			},
			wantType:  reflect.TypeOf(int(0)),
			wantValue: 42,
		},
		{
			name: "string",
			setup: func() reflect.Value {
				return reflect.ValueOf("hello")
			},
			wantType:  reflect.TypeOf(""),
			wantValue: "hello",
		},
		{
			name: "non nil pointer",
			setup: func() reflect.Value {
				v := 42
				return reflect.ValueOf(&v)
			},
			wantType:  reflect.TypeOf(int(0)),
			wantValue: 42,
		},
		{
			name: "double pointer",
			setup: func() reflect.Value {
				v := 42
				p := &v
				return reflect.ValueOf(&p)
			},
			wantType:  reflect.TypeOf(int(0)),
			wantValue: 42,
		},
		{
			name: "nil pointer",
			setup: func() reflect.Value {
				var v *int
				return reflect.ValueOf(&v)
			},
			wantType:  reflect.TypeOf(int(0)),
			wantValue: 0,
		},
		{
			name: "nil double pointer",
			setup: func() reflect.Value {
				var v **int
				return reflect.ValueOf(&v)
			},
			wantType:  reflect.TypeOf(int(0)),
			wantValue: 0,
		},
		{
			name: "nil triple pointer",
			setup: func() reflect.Value {
				var v ***int
				return reflect.ValueOf(&v)
			},
			wantType:  reflect.TypeOf(int(0)),
			wantValue: 0,
		},
		{
			name: "nil pointer to struct",
			setup: func() reflect.Value {
				var v *MyStruct
				return reflect.ValueOf(&v)
			},
			wantType:  reflect.TypeOf(MyStruct{}),
			wantValue: MyStruct{},
		},
		{
			name: "non nil pointer to struct",
			setup: func() reflect.Value {
				v := MyStruct{Value: 99}
				return reflect.ValueOf(&v)
			},
			wantType:  reflect.TypeOf(MyStruct{}),
			wantValue: MyStruct{Value: 99},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Value(tt.setup())

			if !got.IsValid() {
				t.Fatal("got invalid reflect.Value")
			}

			if got.Type() != tt.wantType {
				t.Fatalf("got type %v, want %v", got.Type(), tt.wantType)
			}

			if !reflect.DeepEqual(got.Interface(), tt.wantValue) {
				t.Fatalf(
					"got %#v, want %#v",
					got.Interface(),
					tt.wantValue,
				)
			}
		})
	}
}

func TestValueInitializesNilPointer(t *testing.T) {
	var value *int
	outer := &value

	got := Value(reflect.ValueOf(outer))

	if value == nil {
		t.Fatal("expected value to be initialized")
	}

	if got.Kind() != reflect.Int {
		t.Fatalf("got kind %v, want int", got.Kind())
	}

	if got.Int() != 0 {
		t.Fatalf("got %d, want 0", got.Int())
	}
}

func TestValueInitializesNestedNilPointers(t *testing.T) {
	var value **int
	outer := &value

	got := Value(reflect.ValueOf(outer))

	if value == nil {
		t.Fatal("expected first pointer to be initialized")
	}

	if *value == nil {
		t.Fatal("expected second pointer to be initialized")
	}

	if got.Kind() != reflect.Int {
		t.Fatalf("got kind %v, want int", got.Kind())
	}

	if got.Int() != 0 {
		t.Fatalf("got %d, want 0", got.Int())
	}
}

func TestValuePreservesExistingValue(t *testing.T) {
	value := 123
	ptr := &value

	got := Value(reflect.ValueOf(ptr))

	if got.Kind() != reflect.Int {
		t.Fatalf("got kind %v, want int", got.Kind())
	}

	if got.Int() != 123 {
		t.Fatalf("got %d, want 123", got.Int())
	}

	if value != 123 {
		t.Fatalf("value changed: got %d, want 123", value)
	}
}

func TestValueInitializesMultipleNilLevels(t *testing.T) {
	var value ***int
	outer := &value

	got := Value(reflect.ValueOf(outer))

	if value == nil {
		t.Fatal("first level was not initialized")
	}

	if *value == nil {
		t.Fatal("second level was not initialized")
	}

	if **value == nil {
		t.Fatal("third level was not initialized")
	}

	if got.Kind() != reflect.Int {
		t.Fatalf("got kind %v, want int", got.Kind())
	}

	if got.Int() != 0 {
		t.Fatalf("got %d, want 0", got.Int())
	}
}

func TestValuePointerToPointerToStruct(t *testing.T) {
	type Item struct {
		Name string
	}

	var item *Item
	value := &item

	got := Value(reflect.ValueOf(value))

	if got.Type() != reflect.TypeOf(Item{}) {
		t.Fatalf(
			"got type %v, want %v",
			got.Type(),
			reflect.TypeOf(Item{}),
		)
	}

	if item == nil {
		t.Fatal("expected item to be initialized")
	}

	if got.FieldByName("Name").String() != "" {
		t.Fatalf("expected zero-value Name")
	}
}
