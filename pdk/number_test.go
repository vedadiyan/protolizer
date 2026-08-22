package pdk

import (
	"bytes"
	"testing"
)

func TestSignedNumberInlineEncoder(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		wire  WireType
		want  []byte
	}{
		{
			name:  "i32 positive",
			value: 123,
			wire:  WireTypeI32,
			want:  []byte{0x7b, 0x00, 0x00, 0x00},
		},
		{
			name:  "i32 negative",
			value: -123,
			wire:  WireTypeI32,
			want:  []byte{0x85, 0xff, 0xff, 0xff},
		},
		{
			name:  "i32 zero",
			value: 0,
			wire:  WireTypeI32,
			want:  []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "i32 max",
			value: 2147483647,
			wire:  WireTypeI32,
			want:  []byte{0xff, 0xff, 0xff, 0x7f},
		},
		{
			name:  "i32 min",
			value: -2147483648,
			wire:  WireTypeI32,
			want:  []byte{0x00, 0x00, 0x00, 0x80},
		},
		{
			name:  "i64 positive",
			value: 123,
			wire:  WireTypeI64,
			want:  []byte{0x7b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "i64 negative",
			value: -123,
			wire:  WireTypeI64,
			want:  []byte{0x85, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "i64 zero",
			value: 0,
			wire:  WireTypeI64,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "i64 max",
			value: 9223372036854775807,
			wire:  WireTypeI64,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
		},
		{
			name:  "i64 min",
			value: -9223372036854775808,
			wire:  WireTypeI64,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			SignedNumberInlineEncoder(tt.value, tt.wire, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestSignedNumberInlineEncoderDefault(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want:  []byte{0x00},
		},
		{
			name:  "one",
			value: 1,
			want:  []byte{0x02},
		},
		{
			name:  "negative one",
			value: -1,
			want:  []byte{0x01},
		},
		{
			name:  "positive 123",
			value: 123,
			want:  []byte{0xf6, 0x01},
		},
		{
			name:  "negative 123",
			value: -123,
			want:  []byte{0xf5, 0x01},
		},
		{
			name:  "max int64",
			value: 9223372036854775807,
			want:  []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
		},
		{
			name:  "min int64",
			value: -9223372036854775808,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			SignedNumberInlineEncoder(tt.value, WireTypeVarint, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestUnsignedNumberInlineEncoder(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
		wire  WireType
		want  []byte
	}{
		{
			name:  "i32 zero",
			value: 0,
			wire:  WireTypeI32,
			want:  []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "i32 positive",
			value: 123,
			wire:  WireTypeI32,
			want:  []byte{0x7b, 0x00, 0x00, 0x00},
		},
		{
			name:  "i32 high bit",
			value: 0x80000000,
			wire:  WireTypeI32,
			want:  []byte{0x00, 0x00, 0x00, 0x80},
		},
		{
			name:  "i32 max uint32",
			value: 0xffffffff,
			wire:  WireTypeI32,
			want:  []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "i64 zero",
			value: 0,
			wire:  WireTypeI64,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "i64 positive",
			value: 123,
			wire:  WireTypeI64,
			want:  []byte{0x7b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "i64 high bit",
			value: 0x8000000000000000,
			wire:  WireTypeI64,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
		},
		{
			name:  "i64 max",
			value: ^uint64(0),
			wire:  WireTypeI64,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			UnsignedNumberInlineEncoder(tt.value, tt.wire, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestUnsignedNumberInlineEncoderDefault(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want:  []byte{0x00},
		},
		{
			name:  "one",
			value: 1,
			want:  []byte{0x01},
		},
		{
			name:  "127",
			value: 127,
			want:  []byte{0x7f},
		},
		{
			name:  "128",
			value: 128,
			want:  []byte{0x80, 0x01},
		},
		{
			name:  "max uint64",
			value: ^uint64(0),
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			UnsignedNumberInlineEncoder(tt.value, WireTypeVarint, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestUnsignedNumberDecoder(t *testing.T) {
	tests := []struct {
		name  string
		wire  WireType
		input []byte
		want  uint64
	}{
		{
			name:  "i32 zero",
			wire:  WireTypeI32,
			input: []byte{0x00, 0x00, 0x00, 0x00},
			want:  0,
		},
		{
			name:  "i32 positive",
			wire:  WireTypeI32,
			input: []byte{0x7b, 0x00, 0x00, 0x00},
			want:  123,
		},
		{
			name:  "i32 high bit",
			wire:  WireTypeI32,
			input: []byte{0x00, 0x00, 0x00, 0x80},
			want:  uint64(0x80000000),
		},
		{
			name:  "i32 all bits",
			wire:  WireTypeI32,
			input: []byte{0xff, 0xff, 0xff, 0xff},
			want:  uint64(^uint32(0)),
		},
		{
			name:  "i64 zero",
			wire:  WireTypeI64,
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:  0,
		},
		{
			name:  "i64 positive",
			wire:  WireTypeI64,
			input: []byte{0x7b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:  123,
		},
		{
			name:  "i64 high bit",
			wire:  WireTypeI64,
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
			want:  0x8000000000000000,
		},
		{
			name:  "i64 all bits",
			wire:  WireTypeI64,
			input: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			want:  ^uint64(0),
		},
		{
			name:  "varint zero",
			wire:  WireTypeVarint,
			input: []byte{0x00},
			want:  0,
		},
		{
			name:  "varint 128",
			wire:  WireTypeVarint,
			input: []byte{0x80, 0x01},
			want:  128,
		},
		{
			name:  "varint max",
			wire:  WireTypeVarint,
			input: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
			want:  ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := UnsignedNumberDecoder(tt.wire, buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("got %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestUnsignedNumberDecoderErrors(t *testing.T) {
	tests := []struct {
		name string
		wire WireType
		data []byte
	}{
		{
			name: "i32 insufficient",
			wire: WireTypeI32,
			data: []byte{0x01, 0x02, 0x03},
		},
		{
			name: "i64 insufficient",
			wire: WireTypeI64,
			data: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name: "varint truncated",
			wire: WireTypeVarint,
			data: []byte{0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.data)

			got, err := UnsignedNumberDecoder(tt.wire, buffer)
			if err == nil {
				t.Fatal("expected error")
			}

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}
		})
	}
}

func TestSignedNumberDecoder(t *testing.T) {
	tests := []struct {
		name  string
		wire  WireType
		input []byte
		want  int64
	}{
		{
			name:  "i32 zero",
			wire:  WireTypeI32,
			input: []byte{0x00, 0x00, 0x00, 0x00},
			want:  0,
		},
		{
			name:  "i32 positive",
			wire:  WireTypeI32,
			input: []byte{0x7b, 0x00, 0x00, 0x00},
			want:  123,
		},
		{
			name:  "i32 negative",
			wire:  WireTypeI32,
			input: []byte{0x85, 0xff, 0xff, 0xff},
			want:  -123,
		},
		{
			name:  "i32 min",
			wire:  WireTypeI32,
			input: []byte{0x00, 0x00, 0x00, 0x80},
			want:  -2147483648,
		},
		{
			name:  "i64 zero",
			wire:  WireTypeI64,
			input: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			want:  0,
		},
		{
			name:  "i64 positive",
			wire:  WireTypeI64,
			input: []byte{0x7b, 0, 0, 0, 0, 0, 0, 0},
			want:  123,
		},
		{
			name:  "i64 negative",
			wire:  WireTypeI64,
			input: []byte{0x85, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			want:  -123,
		},
		{
			name:  "i64 min",
			wire:  WireTypeI64,
			input: []byte{0, 0, 0, 0, 0, 0, 0, 0x80},
			want:  -9223372036854775808,
		},
		{
			name:  "varint zero",
			wire:  WireTypeVarint,
			input: []byte{0x00},
			want:  0,
		},
		{
			name:  "varint positive",
			wire:  WireTypeVarint,
			input: []byte{0xf6, 0x01},
			want:  123,
		},
		{
			name:  "varint negative",
			wire:  WireTypeVarint,
			input: []byte{0xf5, 0x01},
			want:  -123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := SignedNumberDecoder(tt.wire, buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("got %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestSignedNumberDecoderErrors(t *testing.T) {
	tests := []struct {
		name string
		wire WireType
		data []byte
	}{
		{
			name: "i32 insufficient",
			wire: WireTypeI32,
			data: []byte{0x01, 0x02, 0x03},
		},
		{
			name: "i64 insufficient",
			wire: WireTypeI64,
			data: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name: "varint truncated",
			wire: WireTypeVarint,
			data: []byte{0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.data)

			got, err := SignedNumberDecoder(tt.wire, buffer)
			if err == nil {
				t.Fatal("expected error")
			}

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}
		})
	}
}

func TestSignedNumberRoundTrip(t *testing.T) {
	values := []int64{
		0,
		1,
		-1,
		123,
		-123,
		2147483647,
		-2147483648,
		9223372036854775807,
		-9223372036854775808,
	}

	for _, value := range values {
		t.Run("i32", func(t *testing.T) {
			var buffer bytes.Buffer

			SignedNumberInlineEncoder(value, WireTypeI32, &buffer)

			got, err := SignedNumberDecoder(WireTypeI32, &buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := int64(int32(value))
			if got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		})

		t.Run("i64", func(t *testing.T) {
			var buffer bytes.Buffer

			SignedNumberInlineEncoder(value, WireTypeI64, &buffer)

			got, err := SignedNumberDecoder(WireTypeI64, &buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})

		t.Run("varint", func(t *testing.T) {
			var buffer bytes.Buffer

			SignedNumberInlineEncoder(value, WireTypeVarint, &buffer)

			got, err := SignedNumberDecoder(WireTypeVarint, &buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestUnsignedNumberRoundTrip(t *testing.T) {
	values := []uint64{
		0,
		1,
		127,
		128,
		12345,
		0xffffffff,
		0x80000000,
		^uint64(0),
	}

	for _, value := range values {
		t.Run("i32", func(t *testing.T) {
			var buffer bytes.Buffer

			UnsignedNumberInlineEncoder(value, WireTypeI32, &buffer)

			got, err := UnsignedNumberDecoder(WireTypeI32, &buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := uint64(uint32(value))
			if got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		})

		t.Run("i64", func(t *testing.T) {
			var buffer bytes.Buffer

			UnsignedNumberInlineEncoder(value, WireTypeI64, &buffer)

			got, err := UnsignedNumberDecoder(WireTypeI64, &buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})

		t.Run("varint", func(t *testing.T) {
			var buffer bytes.Buffer

			UnsignedNumberInlineEncoder(value, WireTypeVarint, &buffer)

			got, err := UnsignedNumberDecoder(WireTypeVarint, &buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestSignedNumberInlineEncoderAppends(t *testing.T) {
	var buffer bytes.Buffer
	buffer.Write([]byte{0xaa, 0xbb})

	SignedNumberInlineEncoder(123, WireTypeI32, &buffer)

	want := []byte{0xaa, 0xbb, 0x7b, 0x00, 0x00, 0x00}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestUnsignedNumberInlineEncoderAppends(t *testing.T) {
	var buffer bytes.Buffer
	buffer.Write([]byte{0xaa, 0xbb})

	UnsignedNumberInlineEncoder(123, WireTypeI64, &buffer)

	want := []byte{
		0xaa, 0xbb,
		0x7b, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}
