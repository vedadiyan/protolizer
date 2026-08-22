package pdk

import (
	"bytes"
	"testing"
)

func TestFixed32Encode(t *testing.T) {
	tests := []struct {
		name  string
		value int32
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want:  []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "one",
			value: 1,
			want:  []byte{0x01, 0x00, 0x00, 0x00},
		},
		{
			name:  "negative one",
			value: -1,
			want:  []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "255",
			value: 255,
			want:  []byte{0xff, 0x00, 0x00, 0x00},
		},
		{
			name:  "256",
			value: 256,
			want:  []byte{0x00, 0x01, 0x00, 0x00},
		},
		{
			name:  "max int32",
			value: 1<<31 - 1,
			want:  []byte{0xff, 0xff, 0xff, 0x7f},
		},
		{
			name:  "min int32",
			value: -1 << 31,
			want:  []byte{0x00, 0x00, 0x00, 0x80},
		},
		{
			name:  "high bit",
			value: 1 << 30,
			want:  []byte{0x00, 0x00, 0x00, 0x40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fixed32Encode(tt.value)

			if got == nil {
				t.Fatal("got nil buffer")
			}

			if got.Len() != 4 {
				t.Fatalf("got length %d, want 4", got.Len())
			}

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestFixed32InlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value int32
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want:  []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "one",
			value: 1,
			want:  []byte{0x01, 0x00, 0x00, 0x00},
		},
		{
			name:  "negative one",
			value: -1,
			want:  []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "255",
			value: 255,
			want:  []byte{0xff, 0x00, 0x00, 0x00},
		},
		{
			name:  "256",
			value: 256,
			want:  []byte{0x00, 0x01, 0x00, 0x00},
		},
		{
			name:  "max int32",
			value: 1<<31 - 1,
			want:  []byte{0xff, 0xff, 0xff, 0x7f},
		},
		{
			name:  "min int32",
			value: -1 << 31,
			want:  []byte{0x00, 0x00, 0x00, 0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			Fixed32InlineEncode(tt.value, &buffer)

			if buffer.Len() != 4 {
				t.Fatalf("got length %d, want 4", buffer.Len())
			}

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestFixed32InlineEncodeAppends(t *testing.T) {
	var buffer bytes.Buffer

	buffer.Write([]byte{0xaa, 0xbb})

	Fixed32InlineEncode(0x12345678, &buffer)

	want := []byte{
		0xaa,
		0xbb,
		0x78,
		0x56,
		0x34,
		0x12,
	}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestFixed64Encode(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "one",
			value: 1,
			want:  []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "negative one",
			value: -1,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "255",
			value: 255,
			want:  []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "256",
			value: 256,
			want:  []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "max int64",
			value: 1<<63 - 1,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
		},
		{
			name:  "min int64",
			value: -1 << 63,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
		},
		{
			name:  "high bit",
			value: 1 << 62,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fixed64Encode(tt.value)

			if got == nil {
				t.Fatal("got nil buffer")
			}

			if got.Len() != 8 {
				t.Fatalf("got length %d, want 8", got.Len())
			}

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestFixed64InlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "one",
			value: 1,
			want:  []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "negative one",
			value: -1,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:  "255",
			value: 255,
			want:  []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "256",
			value: 256,
			want:  []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "max int64",
			value: 1<<63 - 1,
			want:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
		},
		{
			name:  "min int64",
			value: -1 << 63,
			want:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			Fixed64InlineEncode(tt.value, &buffer)

			if buffer.Len() != 8 {
				t.Fatalf("got length %d, want 8", buffer.Len())
			}

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestFixed64InlineEncodeAppends(t *testing.T) {
	var buffer bytes.Buffer

	buffer.Write([]byte{0xaa, 0xbb})

	Fixed64InlineEncode(0x123456789abcdef0, &buffer)

	want := []byte{
		0xaa,
		0xbb,
		0xf0,
		0xde,
		0xbc,
		0x9a,
		0x78,
		0x56,
		0x34,
		0x12,
	}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestFixed32Decode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int32
	}{
		{
			name:  "zero",
			input: []byte{0x00, 0x00, 0x00, 0x00},
			want:  0,
		},
		{
			name:  "one",
			input: []byte{0x01, 0x00, 0x00, 0x00},
			want:  1,
		},
		{
			name:  "negative one",
			input: []byte{0xff, 0xff, 0xff, 0xff},
			want:  -1,
		},
		{
			name:  "positive",
			input: []byte{0x78, 0x56, 0x34, 0x12},
			want:  0x12345678,
		},
		{
			name:  "max int32",
			input: []byte{0xff, 0xff, 0xff, 0x7f},
			want:  1<<31 - 1,
		},
		{
			name:  "min int32",
			input: []byte{0x00, 0x00, 0x00, 0x80},
			want:  -1 << 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := Fixed32Decode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("buffer has %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestFixed32DecodeLeavesFollowingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0x78,
		0x56,
		0x34,
		0x12,
		0xaa,
		0xbb,
	})

	got, err := Fixed32Decode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 0x12345678 {
		t.Fatalf("got %x, want %x", got, int32(0x12345678))
	}

	want := []byte{0xaa, 0xbb}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got remaining %x, want %x", buffer.Bytes(), want)
	}
}

func TestFixed32DecodeInsufficientBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: nil,
		},
		{
			name:  "one byte",
			input: []byte{0x01},
		},
		{
			name:  "two bytes",
			input: []byte{0x01, 0x02},
		},
		{
			name:  "three bytes",
			input: []byte{0x01, 0x02, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := Fixed32Decode(buffer)

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != "insufficient bytes for fixed32" {
				t.Fatalf(
					"got %q, want %q",
					err.Error(),
					"insufficient bytes for fixed32",
				)
			}

			if !bytes.Equal(buffer.Bytes(), tt.input) {
				t.Fatalf("decoder consumed bytes on error")
			}
		})
	}
}

func TestFixed64Decode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int64
	}{
		{
			name:  "zero",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:  0,
		},
		{
			name:  "one",
			input: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:  1,
		},
		{
			name:  "negative one",
			input: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			want:  -1,
		},
		{
			name: "positive",
			input: []byte{
				0xf0,
				0xde,
				0xbc,
				0x9a,
				0x78,
				0x56,
				0x34,
				0x12,
			},
			want: 0x123456789abcdef0,
		},
		{
			name: "max int64",
			input: []byte{
				0xff,
				0xff,
				0xff,
				0xff,
				0xff,
				0xff,
				0xff,
				0x7f,
			},
			want: 1<<63 - 1,
		},
		{
			name: "min int64",
			input: []byte{
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x00,
				0x80,
			},
			want: -1 << 63,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := Fixed64Decode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("buffer has %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestFixed64DecodeLeavesFollowingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0xf0,
		0xde,
		0xbc,
		0x9a,
		0x78,
		0x56,
		0x34,
		0x12,
		0xaa,
		0xbb,
	})

	got, err := Fixed64Decode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 0x123456789abcdef0 {
		t.Fatalf("got %x, want %x", got, int64(0x123456789abcdef0))
	}

	want := []byte{0xaa, 0xbb}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got remaining %x, want %x", buffer.Bytes(), want)
	}
}

func TestFixed64DecodeInsufficientBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: nil,
		},
		{
			name:  "one byte",
			input: []byte{0x01},
		},
		{
			name:  "four bytes",
			input: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:  "seven bytes",
			input: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := Fixed64Decode(buffer)

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != "insufficient bytes for fixed64" {
				t.Fatalf(
					"got %q, want %q",
					err.Error(),
					"insufficient bytes for fixed64",
				)
			}

			if !bytes.Equal(buffer.Bytes(), tt.input) {
				t.Fatalf("decoder consumed bytes on error")
			}
		})
	}
}

func TestFixed32EncodeDecodeRoundTrip(t *testing.T) {
	values := []int32{
		0,
		1,
		-1,
		127,
		128,
		255,
		256,
		-128,
		-255,
		1<<30 - 1,
		-1 << 30,
		1<<31 - 1,
		-1 << 31,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := Fixed32Encode(value)

			got, err := Fixed32Decode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestFixed64EncodeDecodeRoundTrip(t *testing.T) {
	values := []int64{
		0,
		1,
		-1,
		127,
		128,
		255,
		256,
		-128,
		-255,
		1 << 30,
		-1 << 30,
		1 << 32,
		-1 << 32,
		1<<62 - 1,
		-1 << 62,
		1<<63 - 1,
		-1 << 63,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := Fixed64Encode(value)

			got, err := Fixed64Decode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestFixed32InlineEncodeDecodeRoundTrip(t *testing.T) {
	values := []int32{
		0,
		1,
		-1,
		127,
		128,
		1 << 30,
		1<<31 - 1,
		-1 << 31,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			Fixed32InlineEncode(value, &buffer)

			got, err := Fixed32Decode(&buffer)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestFixed64InlineEncodeDecodeRoundTrip(t *testing.T) {
	values := []int64{
		0,
		1,
		-1,
		127,
		128,
		1 << 30,
		1 << 32,
		1<<62 - 1,
		1<<63 - 1,
		-1 << 63,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			Fixed64InlineEncode(value, &buffer)

			got, err := Fixed64Decode(&buffer)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestFixed32EncodeMatchesInlineEncode(t *testing.T) {
	values := []int32{
		0,
		1,
		-1,
		1<<31 - 1,
		-1 << 31,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			expected := Fixed32Encode(value)

			var actual bytes.Buffer
			Fixed32InlineEncode(value, &actual)

			if !bytes.Equal(expected.Bytes(), actual.Bytes()) {
				t.Fatalf(
					"got %x, want %x",
					actual.Bytes(),
					expected.Bytes(),
				)
			}
		})
	}
}

func TestFixed64EncodeMatchesInlineEncode(t *testing.T) {
	values := []int64{
		0,
		1,
		-1,
		1<<63 - 1,
		-1 << 63,
		0x123456789abcdef0,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			expected := Fixed64Encode(value)

			var actual bytes.Buffer
			Fixed64InlineEncode(value, &actual)

			if !bytes.Equal(expected.Bytes(), actual.Bytes()) {
				t.Fatalf(
					"got %x, want %x",
					actual.Bytes(),
					expected.Bytes(),
				)
			}
		})
	}
}
