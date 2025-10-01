package protobench

import (
	"bytes"
	"fmt"
	"io"
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
		if _, err := protolizer.Marshal(p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Simple(b *testing.B) {
	p := createSimplePersonPB()
	data, err := protolizer.Marshal(p)
	if err != nil {
		b.Fatal(err)
	}
	var out SimplePerson
	for i := 0; i < b.N; i++ {
		if err := protolizer.Unmarshal(data, &out); err != nil {
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
		if _, err := protolizer.Marshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBUnmarshal_Nested(b *testing.B) {
	m := createNestedMessagePB()
	data, err := protolizer.Marshal(m)
	if err != nil {
		b.Fatal(err)
	}
	var out NestedMessage
	for i := 0; i < b.N; i++ {
		if err := protolizer.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Simple(b *testing.B) {
	p := createSimplePersonPB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.Marshal(p)
		if err != nil {
			b.Fatal(err)
		}
		var out SimplePerson
		if err := protolizer.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Complex(b *testing.B) {
	m := createComplexMessagePB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
		var out ComplexMessage
		if err := protolizer.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPBRoundTrip_Nested(b *testing.B) {
	m := createNestedMessagePB()
	for i := 0; i < b.N; i++ {
		data, err := protolizer.Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
		var out NestedMessage
		if err := protolizer.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func (c *ComplexMessage) New() protolizer.Reflected {
	return new(ComplexMessage)
}

func (c *ComplexMessage) IsZero(f *protolizer.Field) bool {
	switch f.Tags.Protobuf.FieldNum {
	case 1:
		{
			return c.Id == 0
		}
	case 2:
		{
			return len(c.Name) == 0
		}
	case 3:
		{
			return len(c.Email) == 0
		}
	case 4:
		{
			return c.Score == 0
		}
	case 5:
		{
			return c.IsActive == false
		}
	case 6:
		{
			return c.Tags == nil
		}
	case 7:
		{
			return c.Numbers == nil
		}
	case 8:
		{
			return c.Metadata == nil
		}
	case 9:
		{
			return c.Timestamp == 0
		}
	default:
		{
			return true
		}
	}
}

func (c *ComplexMessage) Encode(f *protolizer.Field, buffer *bytes.Buffer) error {
	switch f.Tags.Protobuf.FieldNum {
	case 1:
		{
			data, err := protolizer.UnsignedNumberEncoder(c.Id, f)
			defer protolizer.Dealloc(data)
			if err != nil {
				return err
			}
			data.WriteTo(buffer)
			return nil

		}
	case 2:
		{
			data, err := protolizer.StringEncoder(c.Name, f)
			defer protolizer.Dealloc(data)
			if err != nil {
				return err
			}
			data.WriteTo(buffer)
			return nil
		}
	case 3:
		{
			data, err := protolizer.StringEncoder(c.Email, f)
			defer protolizer.Dealloc(data)
			if err != nil {
				return err
			}
			data.WriteTo(buffer)
			return nil
		}
	case 4:
		{
			data, err := protolizer.DoubleEncoder(c.Score, f)
			defer protolizer.Dealloc(data)
			if err != nil {
				return err
			}
			data.WriteTo(buffer)
			return nil
		}
	case 5:
		{
			data, err := protolizer.BooleanEncoder(c.IsActive, f)
			defer protolizer.Dealloc(data)
			if err != nil {
				return err
			}
			data.WriteTo(buffer)
			return nil
		}
	case 6:
		{
			tag, err := protolizer.TagEncode(int32(f.Tags.Protobuf.FieldNum), protolizer.WireTypeLen)
			defer protolizer.Dealloc(tag)
			if err != nil {
				return err
			}
			for i, x := range c.Tags {
				if i != 0 {
					buffer.Write(tag.Bytes())
				}
				data, err := protolizer.StringEncoder(x, f)
				if err != nil {
					return err
				}
				data.WriteTo(buffer)
				protolizer.Dealloc(data)
			}
			return nil
		}
	case 7:
		{
			innerBuffer := protolizer.Alloc(0)
			defer protolizer.Dealloc(innerBuffer)
			for _, x := range c.Numbers {
				data, err := protolizer.SignedNumberEncoder(int64(x), f)
				if err != nil {
					return err
				}
				data.WriteTo(innerBuffer)
				protolizer.Dealloc(data)
			}
			bytes := protolizer.BytesEncode(innerBuffer.Bytes())
			bytes.WriteTo(buffer)
			protolizer.Dealloc(bytes)
			return nil
		}
	case 8:
		{

			tag, err := protolizer.TagEncode(int32(f.Tags.Protobuf.FieldNum), protolizer.WireTypeLen)
			defer protolizer.Dealloc(tag)
			if err != nil {
				return err
			}
			i := 0
			for key, value := range c.Metadata {
				if i != 0 {
					buffer.Write(tag.Bytes())
				}
				i++
				innerBuffer := protolizer.Alloc(0)
				innerBuffer.Write(f.KeyTag)
				k, err := protolizer.StringEncoder(key, f)
				if err != nil {
					return err
				}

				k.WriteTo(innerBuffer)
				protolizer.Dealloc(k)
				innerBuffer.Write(f.ValueTag)

				v, err := protolizer.StringEncoder(value, f)
				if err != nil {
					return err
				}
				v.WriteTo(innerBuffer)
				protolizer.Dealloc(v)
				bytes := protolizer.BytesEncode(innerBuffer.Bytes())
				bytes.WriteTo(buffer)
				protolizer.Dealloc(innerBuffer)
				protolizer.Dealloc(bytes)
			}
			return nil
		}
	case 9:
		{
			data, err := protolizer.SignedNumberEncoder(c.Timestamp, f)
			defer protolizer.Dealloc(data)
			if err != nil {
				return err
			}
			data.WriteTo(buffer)
			return nil
		}
	default:
		{
			return fmt.Errorf("field not found")
		}
	}
}

func (c *ComplexMessage) Decode(f *protolizer.Field, buffer *bytes.Buffer) error {
	switch f.Tags.Protobuf.FieldNum {
	case 1:
		{
			value, err := protolizer.SignedNumberDecoder(f, buffer)
			if err != nil {
				return err
			}
			c.Id = uint64(value)
			return nil
		}
	case 2:
		{
			value, err := protolizer.StringDecoder(f, buffer)
			if err != nil {
				return err
			}
			c.Name = value
			return nil
		}
	case 3:
		{
			value, err := protolizer.StringDecoder(f, buffer)
			if err != nil {
				return err
			}
			c.Email = value
			return nil
		}
	case 4:
		{
			value, err := protolizer.DoubleDecoder(f, buffer)
			if err != nil {
				return err
			}
			c.Score = value
			return nil
		}
	case 5:
		{
			value, err := protolizer.BooleanDecoder(f, buffer)
			if err != nil {
				return err
			}
			c.IsActive = value
			return nil
		}
	case 6:
		{
			i := 0
			for {
				if i != 0 {
					i, _, read, err := protolizer.TagPeek(buffer)
					if err != nil {
						if err == io.EOF {
							return nil
						}
						return err
					}
					if i != int32(f.Tags.Protobuf.FieldNum) {
						break
					}
					read()
				}
				i++
				value, err := protolizer.StringDecode(buffer)
				if err != nil {
					return nil
				}
				c.Tags = append(c.Tags, value)
			}
			return nil
		}
	case 7:
		{
			bytes, err := protolizer.BytesDecode(buffer)
			if err != nil {
				return err
			}
			innerBuffer := protolizer.Alloc(0)
			innerBuffer.Write(bytes)
			defer protolizer.Dealloc(innerBuffer)
			for innerBuffer.Len() != 0 {
				value, err := protolizer.SignedNumberDecoder(f, innerBuffer)
				if err != nil {
					return err
				}
				c.Numbers = append(c.Numbers, int32(value))
			}
			return nil
		}
	case 8:
		{
			c.Metadata = make(map[string]string)
			i := 0
			for {
				if i != 0 {
					i, _, read, err := protolizer.TagPeek(buffer)
					if err != nil {
						if err == io.EOF {
							return nil
						}
						return err
					}
					if i != int32(f.Tags.Protobuf.FieldNum) {
						break
					}
					read()
				}
				i++
				bytes, err := protolizer.BytesDecode(buffer)
				if err != nil {
					return nil
				}
				innerBuffer := protolizer.Alloc(0)
				innerBuffer.Write(bytes)
				x, _, err := protolizer.TagDecode(innerBuffer)
				if err != nil {
					protolizer.Dealloc(innerBuffer)
					return err
				}
				_ = x
				key, err := protolizer.StringDecoder(f, innerBuffer)
				if err != nil {
					protolizer.Dealloc(innerBuffer)
					return err
				}
				y, _, err := protolizer.TagDecode(innerBuffer)
				if err != nil {
					protolizer.Dealloc(innerBuffer)
					return err
				}
				_ = y
				value, err := protolizer.StringDecoder(f, innerBuffer)
				if err != nil {
					protolizer.Dealloc(innerBuffer)
					return err
				}
				c.Metadata[key] = value
				protolizer.Dealloc(innerBuffer)

			}
			return nil
		}
	case 9:
		{
			value, err := protolizer.SignedNumberDecoder(f, buffer)
			if err != nil {
				return err
			}
			c.Timestamp = value
			return nil
		}
	default:
		{
			return fmt.Errorf("field not found")
		}
	}
}

func (c *ComplexMessage) Type() protolizer.Type {
	return *protolizer.CaptureTypeByName("protobench.ComplexMessage")
}
