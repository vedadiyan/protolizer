# Protolizer

A zero-dependency, reflection-based Protocol Buffers library for Go that enables dynamic serialization and deserialization without requiring generated code or proto.reflect.

## ✨ Features

- **Zero Dependencies**: No external dependencies beyond Go's standard library
- **Dynamic Serialization**: Serialize/deserialize protobuf messages without generated Go code
- **Reflection-Based**: Uses Go's reflection to introspect struct tags and types
- **Map Conversion**: Convert protobuf messages to `map[string]any` for inspection and manipulation
- **Type Registry**: Built-in type registration and schema export/import
- **Wire Format Compliant**: Full support for all protobuf wire types and encoding rules
- **No proto.reflect**: Independent implementation that doesn't rely on Google's protobuf-go

## 🚀 Installation

```bash
go get github.com/vedadiyan/protolizer
```

## 📖 Quick Start

### Basic Usage with Structs

```go
package main

import (
    "fmt"
    "github.com/vedadiyan/protolizer"
)

type Person struct {
    Name  string `protobuf:"bytes,1,opt,name=name,proto3"`
    Age   int32  `protobuf:"varint,2,opt,name=age,proto3"`
    Email string `protobuf:"bytes,3,opt,name=email,proto3"`
}

func main() {
    // Register the type
    protolizer.RegisterTypeFor[Person]()
    
    // Create a person
    person := Person{
        Name:  "John Doe",
        Age:   30,
        Email: "john@example.com",
    }
    
    // Marshal to protobuf bytes
    data, err := protolizer.Marshal(&person)
    if err != nil {
        panic(err)
    }
    
    // Unmarshal back to struct
    var decoded Person
    err = protolizer.Unmarshal(data, &decoded)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Original: %+v\n", person)
    fmt.Printf("Decoded:  %+v\n", decoded)
}
```

### Dynamic Map-Based Usage

```go
// Convert protobuf bytes to map for inspection
personMap, err := protolizer.Read("main.Person", data)
if err != nil {
    panic(err)
}
fmt.Printf("As map: %+v\n", personMap)

// Modify the map
personMap["Age"] = float64(31)
personMap["Email"] = "john.doe@example.com"

// Convert map back to protobuf bytes
newData, err := protolizer.Write("main.Person", personMap)
if err != nil {
    panic(err)
}

// Unmarshal the modified data
var modifiedPerson Person
err = protolizer.Unmarshal(newData, &modifiedPerson)
if err != nil {
    panic(err)
}
fmt.Printf("Modified: %+v\n", modifiedPerson)
```

## 🏗️ Advanced Usage

### Complex Types

```go
type Address struct {
    Street  string `protobuf:"bytes,1,opt,name=street,proto3"`
    City    string `protobuf:"bytes,2,opt,name=city,proto3"`
    Country string `protobuf:"bytes,3,opt,name=country,proto3"`
}

type Contact struct {
    Person    Person             `protobuf:"bytes,1,opt,name=person,proto3"`
    Address   *Address           `protobuf:"bytes,2,opt,name=address,proto3"`
    Phones    []string           `protobuf:"bytes,3,rep,name=phones,proto3"`
    Metadata  map[string]string  `protobuf:"bytes,4,rep,name=metadata,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

// Register all types
protolizer.RegisterTypeFor[Address]()
protolizer.RegisterTypeFor[Contact]()
```

### Supported Field Types

#### Primitive Types
```go
type Primitives struct {
    // Integer types
    Int32Field  int32   `protobuf:"varint,1,opt,name=int32_field,proto3"`
    Int64Field  int64   `protobuf:"varint,2,opt,name=int64_field,proto3"`
    Uint32Field uint32  `protobuf:"varint,3,opt,name=uint32_field,proto3"`
    Uint64Field uint64  `protobuf:"varint,4,opt,name=uint64_field,proto3"`
    
    // Fixed-width types
    Fixed32     uint32  `protobuf:"fixed32,5,opt,name=fixed32,proto3"`
    Fixed64     uint64  `protobuf:"fixed64,6,opt,name=fixed64,proto3"`
    Sfixed32    int32   `protobuf:"fixed32,7,opt,name=sfixed32,proto3"`
    Sfixed64    int64   `protobuf:"fixed64,8,opt,name=sfixed64,proto3"`
    
    // Float types
    FloatField  float32 `protobuf:"fixed32,9,opt,name=float_field,proto3"`
    DoubleField float64 `protobuf:"fixed64,10,opt,name=double_field,proto3"`
    
    // String and bytes
    StringField string  `protobuf:"bytes,11,opt,name=string_field,proto3"`
    BytesField  []byte  `protobuf:"bytes,12,opt,name=bytes_field,proto3"`
    
    // Boolean
    BoolField   bool    `protobuf:"varint,13,opt,name=bool_field,proto3"`
}
```

#### Collections
```go
type Collections struct {
    // Repeated fields
    Numbers    []int32           `protobuf:"varint,1,rep,packed,name=numbers,proto3"`
    Names      []string          `protobuf:"bytes,2,rep,name=names,proto3"`
    
    // Maps
    StringMap  map[string]string `protobuf:"bytes,3,rep,name=string_map,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
    IntMap     map[int32]string  `protobuf:"bytes,4,rep,name=int_map,proto3" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}
```

### Schema Export/Import

```go
// Export type schema
schemaBytes, err := protolizer.ExportType[Person]()
if err != nil {
    panic(err)
}

// Import type schema
importedType, err := protolizer.ImportType(schemaBytes)
if err != nil {
    panic(err)
}

// Export entire module (all related types)
moduleBytes, err := protolizer.ExportModule[Contact]()
if err != nil {
    panic(err)
}

// Import module
module, err := protolizer.ImportModule(moduleBytes)
if err != nil {
    panic(err)
}
```

## 🏷️ Protobuf Tag Format

Protolizer uses standard protobuf struct tags with the following format:

```go
`protobuf:"<wire_type>,<field_number>,<label>,name=<field_name>,<syntax>"`
```

### Wire Types
- `varint` - Variable-length integers (int32, int64, uint32, uint64, bool)
- `fixed64` - Fixed 64-bit values (double, fixed64, sfixed64)
- `bytes` - Length-delimited (string, bytes, messages, packed repeated)
- `fixed32` - Fixed 32-bit values (float, fixed32, sfixed32)

### Labels
- `opt` - Optional field
- `req` - Required field (proto2)
- `rep` - Repeated field

### Map Fields
For map fields, specify key and value wire types:
```go
MapField map[string]int32 `protobuf:"bytes,1,rep,name=map_field,proto3" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
```

## 🎯 API Reference

### Core Functions

#### `RegisterTypeFor[T any]()`
Registers a type in the global type registry for dynamic serialization.

#### `Marshal(v any) ([]byte, error)`
Serializes a Go struct to protobuf wire format.

#### `Unmarshal(bytes []byte, v any) error`
Deserializes protobuf bytes into a Go struct.

#### `Read(typeName string, bytes []byte) (map[string]any, error)`
Converts protobuf bytes to a map for dynamic inspection/manipulation.

#### `Write(typeName string, v map[string]any) ([]byte, error)`
Converts a map back to protobuf bytes.

### Type Introspection

#### `CaptureTypeFor[T any]() *Type`
Returns type information for a registered type.

#### `CaptureType(t reflect.Type) *Type`
Returns type information for a reflect.Type.

#### `CaptureTypeByName(typeName string) *Type`
Returns type information by type name.

### Schema Export/Import

#### `ExportType[T any]() ([]byte, error)`
Exports a single type's schema as protobuf bytes.

#### `ImportType(bytes []byte) (*Type, error)`
Imports a type schema from protobuf bytes.

#### `ExportModule[T any]() ([]byte, error)`
Exports all related types as a module.

#### `ImportModule(bytes []byte) (*Module, error)`
Imports a complete module with all types.

## 🔧 Wire Format Details

Protolizer implements the complete Protocol Buffers wire format specification:

### Encoding Rules
- **Varints**: Variable-length encoding for integers
- **Fixed32/64**: Little-endian fixed-width encoding
- **Length-Delimited**: Length-prefixed encoding for strings, bytes, and messages
- **Packed Repeated**: Efficient encoding for repeated numeric fields

### Tag Format
Each field is prefixed with a tag containing:
- Field number (bits 3+)
- Wire type (bits 0-2)

## ⚡ Performance Considerations

- **Reflection Overhead**: Uses reflection for type introspection, which has some performance cost
- **Memory Allocation**: Creates temporary objects during marshaling/unmarshaling
- **Type Registration**: Types should be registered once at startup, not per operation
- **Large Messages**: For very large messages, consider streaming approaches

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repo
git clone https://github.com/vedadiyan/protolizer.git
cd protolizer

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

## 📋 Limitations

- **No Proto Files**: Does not parse .proto files directly (struct tags define schema)
- **No Code Generation**: Requires manual struct tag annotation
- **Reflection Required**: Cannot eliminate reflection for type safety
- **Go-Specific**: Designed specifically for Go, not cross-language compatible without schema export


## 📄 License

This project is licensed under the Apache 2 License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Protocol Buffers specification by Google
- Go reflection and type system
- The Go community for inspiration and best practices

---

**Made with ❤️ for dynamic protobuf handling in Go**