package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vedadiyan/protolizer"
	"github.com/vedadiyan/protolizer/metadata"
)

func init() {
	// Register all test types
	metadata.RegisterTypeFor[SimplePerson]()
	metadata.RegisterTypeFor[ComplexMessage]()
	metadata.RegisterTypeFor[NestedMessage]()
	metadata.RegisterTypeFor[AddressInfo]()
	metadata.RegisterTypeFor[ExtraData]()
}

// ----- Test data generators -----

func createSimplePersonPB() *SimplePerson {
	return &SimplePerson{
		Name: "John Doe",
		Age:  30,
		Id:   12345,
	}
}

func createComplexMessagePB() *ComplexMessage {
	return &ComplexMessage{
		Id:       67890,
		Name:     "Complex Test Message",
		Email:    "test@example.com",
		Score:    95.5,
		IsActive: true,
		Tags:     []string{"important", "urgent", "customer", "vip"},
		Numbers:  []int{1, 2, 3, 4, 5, 10, 20, 30, 40, 50},
		Metadata: map[string]string{
			"source":      "api",
			"version":     "1.2.3",
			"environment": "production",
			"region":      "us-west-2",
		},
		Timestamp: time.Now().Unix(),
	}
}

func createNestedMessagePB() *NestedMessage {
	return &NestedMessage{
		Person: &SimplePerson{Name: "Jane Smith", Age: 25, Id: 54321},
		Address: &AddressInfo{
			Street:  "123 Main Street",
			City:    "New York",
			Zipcode: "10001",
			Country: "USA",
		},
		Phones: []string{"+1-555-1234", "+1-555-5678", "+1-555-9999"},
		Extra: &ExtraData{
			Notes:    "This is a test message",
			Priority: 1,
			Flags:    []bool{true, false, true},
			Config:   []float64{30.5, 3.0, 100.0},
		},
		PersonArray: []SimplePerson{
			SimplePerson{Name: "Ok", Age: 100, Id: 12345},
		},
	}
}

// ----- Benchmarks -----

func BenchmarkPBMarshal_Simple(b *testing.B) {
	p := createSimplePersonPB()
	for i := 0; i < b.N; i++ {
		if _, err := protolizer.StaticCodec().Marshal(p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Simple(b *testing.B) {
	p := createSimplePersonPB()
	data, err := protolizer.StaticCodec().Marshal(p)
	if err != nil {
		b.Fatal(err)
	}
	var out SimplePerson
	for i := 0; i < b.N; i++ {
		if err := protolizer.StaticCodec().Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBMarshal_Complex(b *testing.B) {
	m := createComplexMessagePB()
	// protolizer.StaticCodec().RegisterTypeFor[ComplexMessage]()
	// fn := protolizer.StaticCodec().BuildEncoder(reflect.TypeOf(m).Elem())
	// fn2 := protolizer.StaticCodec().BuildDecoder(reflect.TypeOf(m).Elem())
	// _ = fn2
	for i := 0; i < b.N; i++ {
		_, err := protolizer.StaticCodec().Marshal(m)
		if err != nil {
			b.Fatal(err)
		}

	}
}

func BenchmarkPBUnmarshal_Complex(b *testing.B) {
	m := createComplexMessagePB()
	data, err := protolizer.StaticCodec().Marshal(m)
	if err != nil {
		b.Fatal(err)
	}
	metadata.RegisterTypeFor[ComplexMessage]()
	// fn2 := protolizer.StaticCodec().BuildDecoder(reflect.TypeOf(m).Elem())
	// t := metadata.CaptureType(reflect.TypeOf(m).Elem())
	// protolizer.TypelessCodec().Register(t)
	var out ComplexMessage
	for i := 0; i < b.N; i++ {
		err := protolizer.StaticCodec().Unmarshal(data, &out)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}

func BenchmarkPBMarshal_Nested(b *testing.B) {
	m := createNestedMessagePB()
	for i := 0; i < b.N; i++ {
		if _, err := protolizer.StaticCodec().Marshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Nested(b *testing.B) {
	m := createNestedMessagePB()
	data, err := protolizer.StaticCodec().Marshal(m)
	if err != nil {
		b.Fatal(err)
	}
	var out NestedMessage
	for i := 0; i < b.N; i++ {
		if err := protolizer.StaticCodec().Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
		if out.Person.Name != m.Person.Name {
			b.Fatal(fmt.Errorf("bad unmarshalling"))
		}
	}
}

func BenchmarkPBRoundTrip_Simple(b *testing.B) {
	p := createSimplePersonPB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.StaticCodec().Marshal(p)
		if err != nil {
			b.Fatal(err)
		}
		var out SimplePerson
		if err := protolizer.StaticCodec().Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Complex(b *testing.B) {
	m := createComplexMessagePB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.StaticCodec().Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
		var out ComplexMessage
		if err := protolizer.StaticCodec().Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Nested(b *testing.B) {
	m := createNestedMessagePB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.StaticCodec().Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
		var out NestedMessage
		if err := protolizer.StaticCodec().Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}
