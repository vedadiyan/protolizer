package protolizer

import (
	"reflect"
	"testing"
	"time"
)

// Test data structures
type SimplePerson struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3"`
	Age  int32  `protobuf:"varint,2,opt,name=age,proto3"`
	ID   uint64 `protobuf:"varint,3,opt,name=id,proto3"`
}

type ComplexMessage struct {
	ID        uint64            `protobuf:"varint,1,opt,name=id,proto3"`
	Name      string            `protobuf:"bytes,2,opt,name=name,proto3"`
	Email     string            `protobuf:"bytes,3,opt,name=email,proto3"`
	Score     float64           `protobuf:"fixed64,4,opt,name=score,proto3"`
	IsActive  bool              `protobuf:"varint,5,opt,name=is_active,proto3"`
	Tags      []string          `protobuf:"bytes,6,rep,name=tags,proto3"`
	Numbers   []int32           `protobuf:"varint,7,rep,packed,name=numbers,proto3"`
	Metadata  map[string]string `protobuf:"bytes,8,rep,name=metadata,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Timestamp int64             `protobuf:"varint,9,opt,name=timestamp,proto3"`
}

type NestedMessage struct {
	Person  SimplePerson `protobuf:"bytes,1,opt,name=person,proto3"`
	Address *AddressInfo `protobuf:"bytes,2,opt,name=address,proto3"`
	Phones  []string     `protobuf:"bytes,3,rep,name=phones,proto3"`
	Extra   *ExtraData   `protobuf:"bytes,4,opt,name=extra,proto3"`
}

type AddressInfo struct {
	Street  string `protobuf:"bytes,1,opt,name=street,proto3"`
	City    string `protobuf:"bytes,2,opt,name=city,proto3"`
	Zipcode string `protobuf:"bytes,3,opt,name=zipcode,proto3"`
	Country string `protobuf:"bytes,4,opt,name=country,proto3"`
}

type ExtraData struct {
	Notes    string    `protobuf:"bytes,1,opt,name=notes,proto3"`
	Priority int32     `protobuf:"varint,2,opt,name=priority,proto3"`
	Flags    []bool    `protobuf:"varint,3,rep,packed,name=flags,proto3"`
	Config   []float64 `protobuf:"fixed64,4,rep,packed,name=config,proto3"`
}

// Simpler message without complex maps to avoid map[any]any issues
type SimpleMessage struct {
	ID       uint64   `protobuf:"varint,1,opt,name=id,proto3"`
	Name     string   `protobuf:"bytes,2,opt,name=name,proto3"`
	Email    string   `protobuf:"bytes,3,opt,name=email,proto3"`
	IsActive bool     `protobuf:"varint,4,opt,name=is_active,proto3"`
	Tags     []string `protobuf:"bytes,5,rep,name=tags,proto3"`
	Numbers  []int32  `protobuf:"varint,6,rep,packed,name=numbers,proto3"`
}

// Test data generators
func createSimplePerson() SimplePerson {
	return SimplePerson{
		Name: "John Doe",
		Age:  30,
		ID:   12345,
	}
}

func createComplexMessage() ComplexMessage {
	return ComplexMessage{
		ID:       67890,
		Name:     "Complex Test Message",
		Email:    "test@example.com",
		Score:    95.5,
		IsActive: true,
		Tags:     []string{"important", "urgent", "customer", "vip"},
		Numbers:  []int32{1, 2, 3, 4, 5, 10, 20, 30, 40, 50},
		Metadata: map[string]string{
			"source":      "api",
			"version":     "1.2.3",
			"environment": "production",
			"region":      "us-west-2",
		},
		Timestamp: time.Now().Unix(),
	}
}

func createNestedMessage() NestedMessage {
	return NestedMessage{
		Person: SimplePerson{
			Name: "Jane Smith",
			Age:  25,
			ID:   54321,
		},
		Address: &AddressInfo{
			Street:  "123 Main Street",
			City:    "New York",
			Zipcode: "10001",
			Country: "USA",
		},
		Phones: []string{"+1-555-1234", "+1-555-5678", "+1-555-9999"},
		Extra: &ExtraData{
			Notes:    "This is a test message with nested structures",
			Priority: 1,
			Flags:    []bool{true, false, true, true, false},
			Config:   []float64{30.5, 3.0, 100.0},
		},
	}
}

func createSimpleMessage() SimpleMessage {
	return SimpleMessage{
		ID:       67890,
		Name:     "Simple Test Message",
		Email:    "test@example.com",
		IsActive: true,
		Tags:     []string{"test", "benchmark"},
		Numbers:  []int32{1, 2, 3, 4, 5},
	}
}

func init() {
	// Register all test types
	RegisterTypeFor[SimplePerson]()
	RegisterTypeFor[ComplexMessage]()
	RegisterTypeFor[NestedMessage]()
	RegisterTypeFor[AddressInfo]()
	RegisterTypeFor[ExtraData]()
	RegisterTypeFor[SimpleMessage]()
}

// Benchmark Marshal operations
func BenchmarkMarshal_Simple(b *testing.B) {
	person := createSimplePerson()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Marshal(person)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Complex(b *testing.B) {
	msg := createComplexMessage()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Nested(b *testing.B) {
	msg := createNestedMessage()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark Unmarshal operations
func BenchmarkUnmarshal_Simple(b *testing.B) {
	person := createSimplePerson()
	data, err := Marshal(person)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decoded SimplePerson
		err := Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Complex(b *testing.B) {
	msg := createComplexMessage()
	data, err := Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decoded ComplexMessage
		err := Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Nested(b *testing.B) {
	msg := createNestedMessage()
	data, err := Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decoded NestedMessage
		err := Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark Round-trip operations
func BenchmarkRoundTrip_Simple(b *testing.B) {
	person := createSimplePerson()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := Marshal(person)
		if err != nil {
			b.Fatal(err)
		}
		var decoded SimplePerson
		err = Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTrip_Complex(b *testing.B) {
	msg := createComplexMessage()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		var decoded ComplexMessage
		err = Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTrip_Nested(b *testing.B) {
	msg := createNestedMessage()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		var decoded NestedMessage
		err = Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark Map-based operations using SimpleMessage to avoid map issues
func BenchmarkRead_Simple(b *testing.B) {
	person := createSimplePerson()
	data, err := Marshal(person)
	if err != nil {
		b.Fatal(err)
	}
	typeName := getTypeName(person)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Read(typeName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRead_SimpleMessage(b *testing.B) {
	msg := createSimpleMessage()
	data, err := Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	typeName := getTypeName(msg)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Read(typeName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWrite_Simple(b *testing.B) {
	person := createSimplePerson()
	data, err := Marshal(person)
	if err != nil {
		b.Fatal(err)
	}
	typeName := getTypeName(person)
	mapData, err := Read(typeName, data)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Write(typeName, mapData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWrite_SimpleMessage(b *testing.B) {
	msg := createSimpleMessage()
	data, err := Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	typeName := getTypeName(msg)
	mapData, err := Read(typeName, data)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Write(typeName, mapData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark Map round-trip operations
func BenchmarkMapRoundTrip_Simple(b *testing.B) {
	person := createSimplePerson()
	data, err := Marshal(person)
	if err != nil {
		b.Fatal(err)
	}
	typeName := getTypeName(person)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapData, err := Read(typeName, data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Write(typeName, mapData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapRoundTrip_SimpleMessage(b *testing.B) {
	msg := createSimpleMessage()
	data, err := Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	typeName := getTypeName(msg)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapData, err := Read(typeName, data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Write(typeName, mapData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark specific encoding operations
func BenchmarkEncodeVarint(b *testing.B) {
	values := []int64{1, 127, 128, 16383, 16384, 2097151, 2097152, 268435455, 268435456}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, val := range values {
			encodeVarint(val)
		}
	}
}

func BenchmarkDecodeVarint(b *testing.B) {
	// Pre-encode test values
	testData := make([][]byte, 0)
	values := []int64{1, 127, 128, 16383, 16384, 2097151, 2097152, 268435455, 268435456}
	for _, val := range values {
		testData = append(testData, encodeVarint(val))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, data := range testData {
			_, _, _ = decodeVarint(data, 0)
		}
	}
}

func BenchmarkEncodeString(b *testing.B) {
	strings := []string{
		"",
		"hello",
		"Hello, World!",
		"This is a longer string for testing purposes",
		"🌍 Unicode string with emojis 🚀 and special characters éñ",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, str := range strings {
			encodeString(str)
		}
	}
}

func BenchmarkDecodeString(b *testing.B) {
	// Pre-encode test strings
	testData := make([][]byte, 0)
	strings := []string{
		"",
		"hello",
		"Hello, World!",
		"This is a longer string for testing purposes",
		"🌍 Unicode string with emojis 🚀 and special characters éñ",
	}
	for _, str := range strings {
		testData = append(testData, encodeString(str))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, data := range testData {
			_, _, _ = decodeString(data, 0)
		}
	}
}

// Benchmark type operations
func BenchmarkTypeRegistration(b *testing.B) {
	type TempStruct struct {
		Field1 string `protobuf:"bytes,1,opt,name=field1,proto3"`
		Field2 int32  `protobuf:"varint,2,opt,name=field2,proto3"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterTypeFor[TempStruct]()
	}
}

func BenchmarkExportType(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExportType[SimplePerson]()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImportType(b *testing.B) {
	exported, err := ExportType[SimplePerson]()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ImportType(exported)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory allocation benchmarks
func BenchmarkMarshal_Allocs(b *testing.B) {
	person := createSimplePerson()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Marshal(person)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Allocs(b *testing.B) {
	person := createSimplePerson()
	data, err := Marshal(person)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decoded SimplePerson
		err := Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Utility function to get type name for benchmarks
func getTypeName(v interface{}) string {
	return TypeName(reflect.TypeOf(v))
}

// Size comparison benchmarks
func BenchmarkDataSize(b *testing.B) {
	person := createSimplePerson()
	complex := createComplexMessage()
	nested := createNestedMessage()
	simple := createSimpleMessage()

	personData, _ := Marshal(person)
	complexData, _ := Marshal(complex)
	nestedData, _ := Marshal(nested)
	simpleData, _ := Marshal(simple)

	b.Logf("Simple person data size: %d bytes", len(personData))
	b.Logf("Complex message data size: %d bytes", len(complexData))
	b.Logf("Nested message data size: %d bytes", len(nestedData))
	b.Logf("Simple message data size: %d bytes", len(simpleData))
}
