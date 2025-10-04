package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vedadiyan/protolizer"
)

func init() {
	// Register all test types
	protolizer.RegisterTypeFor[SimplePerson]()
	protolizer.RegisterTypeFor[ComplexMessage]()
	protolizer.RegisterTypeFor[NestedMessage]()
	protolizer.RegisterTypeFor[AddressInfo]()
	protolizer.RegisterTypeFor[ExtraData]()
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
	}
}

// ----- Benchmarks -----

func BenchmarkPBMarshal_Simple(b *testing.B) {
	p := createSimplePersonPB()
	for i := 0; i < b.N; i++ {
		if _, err := protolizer.FastMarshal(p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Simple(b *testing.B) {
	p := createSimplePersonPB()
	data, err := protolizer.FastMarshal(p)
	if err != nil {
		b.Fatal(err)
	}
	var out SimplePerson
	for i := 0; i < b.N; i++ {
		if err := protolizer.FastUnmarshal(&out, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBMarshal_Complex(b *testing.B) {
	m := createComplexMessagePB()
	for i := 0; i < b.N; i++ {
		if _, err := protolizer.FastMarshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Complex(b *testing.B) {
	m := createComplexMessagePB()
	data, err := protolizer.FastMarshal(m)
	if err != nil {
		b.Fatal(err)
	}
	var out ComplexMessage
	for i := 0; i < b.N; i++ {
		if err := protolizer.FastUnmarshal(&out, data); err != nil {
			b.Fatal(err)
		}
		if out.Email != m.Email {
			b.Fatal(fmt.Errorf("bad unmarshalling"))
		}
	}
}

func BenchmarkPBMarshal_Nested(b *testing.B) {
	m := createNestedMessagePB()
	for i := 0; i < b.N; i++ {
		if _, err := protolizer.FastMarshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Nested(b *testing.B) {
	m := createNestedMessagePB()
	data, err := protolizer.FastMarshal(m)
	if err != nil {
		b.Fatal(err)
	}
	var out NestedMessage
	for i := 0; i < b.N; i++ {
		if err := protolizer.FastUnmarshal(&out, data); err != nil {
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
		data, err := protolizer.FastMarshal(p)
		if err != nil {
			b.Fatal(err)
		}
		var out SimplePerson
		if err := protolizer.FastUnmarshal(&out, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Complex(b *testing.B) {
	m := createComplexMessagePB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.FastMarshal(m)
		if err != nil {
			b.Fatal(err)
		}
		var out ComplexMessage
		if err := protolizer.FastUnmarshal(&out, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Nested(b *testing.B) {
	m := createNestedMessagePB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.FastMarshal(m)
		if err != nil {
			b.Fatal(err)
		}
		var out NestedMessage
		if err := protolizer.FastUnmarshal(&out, data); err != nil {
			b.Fatal(err)
		}
	}
}
