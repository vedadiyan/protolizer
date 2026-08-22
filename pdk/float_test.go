package pdk

import (
	"bytes"
	"math"
	"testing"
)

func TestFloat32Encode(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  []byte
	}{
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"negative zero", float32(math.Copysign(0, -1)), []byte{0x00, 0x00, 0x00, 0x80}},
		{"one", 1, []byte{0x00, 0x00, 0x80, 0x3f}},
		{"negative one", -1, []byte{0x00, 0x00, 0x80, 0xbf}},
		{"pi", float32(math.Pi), []byte{0xdb, 0x0f, 0x49, 0x40}},
		{"negative pi", -float32(math.Pi), []byte{0xdb, 0x0f, 0x49, 0xc0}},
		{"positive infinity", float32(math.Inf(1)), []byte{0x00, 0x00, 0x80, 0x7f}},
		{"negative infinity", float32(math.Inf(-1)), []byte{0x00, 0x00, 0x80, 0xff}},
		{"smallest positive subnormal", math.Float32frombits(1), []byte{0x01, 0x00, 0x00, 0x00}},
		{"largest finite", math.MaxFloat32, []byte{0xff, 0xff, 0x7f, 0x7f}},
		{"nan", math.Float32frombits(0x7fc00001), []byte{0x01, 0x00, 0xc0, 0x7f}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Float32Encode(tt.value)

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

func TestFloat32InlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  []byte
	}{
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"negative zero", float32(math.Copysign(0, -1)), []byte{0x00, 0x00, 0x00, 0x80}},
		{"one", 1, []byte{0x00, 0x00, 0x80, 0x3f}},
		{"negative one", -1, []byte{0x00, 0x00, 0x80, 0xbf}},
		{"pi", float32(math.Pi), []byte{0xdb, 0x0f, 0x49, 0x40}},
		{"positive infinity", float32(math.Inf(1)), []byte{0x00, 0x00, 0x80, 0x7f}},
		{"negative infinity", float32(math.Inf(-1)), []byte{0x00, 0x00, 0x80, 0xff}},
		{"smallest positive subnormal", math.Float32frombits(1), []byte{0x01, 0x00, 0x00, 0x00}},
		{"largest finite", math.MaxFloat32, []byte{0xff, 0xff, 0x7f, 0x7f}},
		{"nan", math.Float32frombits(0x7fc00001), []byte{0x01, 0x00, 0xc0, 0x7f}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			Float32InlineEncode(tt.value, &buffer)

			if buffer.Len() != 4 {
				t.Fatalf("got length %d, want 4", buffer.Len())
			}
			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestFloat32InlineEncodeAppends(t *testing.T) {
	var buffer bytes.Buffer
	buffer.Write([]byte{0xaa, 0xbb})

	Float32InlineEncode(1, &buffer)

	want := []byte{0xaa, 0xbb, 0x00, 0x00, 0x80, 0x3f}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestFloat64Encode(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  []byte
	}{
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"negative zero", math.Copysign(0, -1), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}},
		{"one", 1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f}},
		{"negative one", -1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xbf}},
		{"pi", math.Pi, []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}},
		{"negative pi", -math.Pi, []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0xc0}},
		{"positive infinity", math.Inf(1), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f}},
		{"negative infinity", math.Inf(-1), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff}},
		{"smallest positive subnormal", math.Float64frombits(1), []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"largest finite", math.MaxFloat64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f}},
		{"nan", math.Float64frombits(0x7ff8000000000001), []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Float64Encode(tt.value)

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

func TestFloat64InlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  []byte
	}{
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"negative zero", math.Copysign(0, -1), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}},
		{"one", 1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f}},
		{"negative one", -1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xbf}},
		{"pi", math.Pi, []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}},
		{"positive infinity", math.Inf(1), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f}},
		{"negative infinity", math.Inf(-1), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff}},
		{"smallest positive subnormal", math.Float64frombits(1), []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"largest finite", math.MaxFloat64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f}},
		{"nan", math.Float64frombits(0x7ff8000000000001), []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			Float64InlineEncode(tt.value, &buffer)

			if buffer.Len() != 8 {
				t.Fatalf("got length %d, want 8", buffer.Len())
			}
			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestFloat64InlineEncodeAppends(t *testing.T) {
	var buffer bytes.Buffer
	buffer.Write([]byte{0xaa, 0xbb})

	Float64InlineEncode(1, &buffer)

	want := []byte{
		0xaa, 0xbb,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf0, 0x3f,
	}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestFloat32Decode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  float32
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"negative zero", []byte{0x00, 0x00, 0x00, 0x80}, float32(math.Copysign(0, -1))},
		{"one", []byte{0x00, 0x00, 0x80, 0x3f}, 1},
		{"negative one", []byte{0x00, 0x00, 0x80, 0xbf}, -1},
		{"pi", []byte{0xdb, 0x0f, 0x49, 0x40}, float32(math.Pi)},
		{"positive infinity", []byte{0x00, 0x00, 0x80, 0x7f}, float32(math.Inf(1))},
		{"negative infinity", []byte{0x00, 0x00, 0x80, 0xff}, float32(math.Inf(-1))},
		{"subnormal", []byte{0x01, 0x00, 0x00, 0x00}, math.Float32frombits(1)},
		{"max finite", []byte{0xff, 0xff, 0x7f, 0x7f}, math.MaxFloat32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := Float32Decode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Float32bits(got) != math.Float32bits(tt.want) {
				t.Fatalf("got bits %08x, want %08x", math.Float32bits(got), math.Float32bits(tt.want))
			}
			if buffer.Len() != 0 {
				t.Fatalf("buffer has %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestFloat32DecodeNaN(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x01, 0x00, 0xc0, 0x7f})

	got, err := Float32Decode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsNaN(float64(got)) {
		t.Fatal("got non-NaN value")
	}
	if math.Float32bits(got) != 0x7fc00001 {
		t.Fatalf("got bits %08x, want 7fc00001", math.Float32bits(got))
	}
}

func TestFloat32DecodeLeavesFollowingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0x00, 0x00, 0x80, 0x3f,
		0xaa, 0xbb,
	})

	got, err := Float32Decode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
	if !bytes.Equal(buffer.Bytes(), []byte{0xaa, 0xbb}) {
		t.Fatalf("got remaining %x, want aabb", buffer.Bytes())
	}
}

func TestFloat32DecodeInsufficientBytes(t *testing.T) {
	tests := [][]byte{
		nil,
		{0x00},
		{0x00, 0x00},
		{0x00, 0x00, 0x00},
	}

	for _, input := range tests {
		t.Run("", func(t *testing.T) {
			buffer := bytes.NewBuffer(input)

			got, err := Float32Decode(buffer)

			if got != 0 {
				t.Fatalf("got %v, want 0", got)
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != "insufficient bytes for float32" {
				t.Fatalf("got %q", err.Error())
			}
			if !bytes.Equal(buffer.Bytes(), input) {
				t.Fatal("decoder consumed bytes on error")
			}
		})
	}
}

func TestFloat64Decode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  float64
	}{
		{"zero", []byte{0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{"negative zero", []byte{0, 0, 0, 0, 0, 0, 0, 0x80}, math.Copysign(0, -1)},
		{"one", []byte{0, 0, 0, 0, 0, 0, 0xf0, 0x3f}, 1},
		{"negative one", []byte{0, 0, 0, 0, 0, 0, 0xf0, 0xbf}, -1},
		{"pi", []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}, math.Pi},
		{"positive infinity", []byte{0, 0, 0, 0, 0, 0, 0xf0, 0x7f}, math.Inf(1)},
		{"negative infinity", []byte{0, 0, 0, 0, 0, 0, 0xf0, 0xff}, math.Inf(-1)},
		{"subnormal", []byte{1, 0, 0, 0, 0, 0, 0, 0}, math.Float64frombits(1)},
		{"max finite", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f}, math.MaxFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := Float64Decode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Float64bits(got) != math.Float64bits(tt.want) {
				t.Fatalf("got bits %016x, want %016x", math.Float64bits(got), math.Float64bits(tt.want))
			}
			if buffer.Len() != 0 {
				t.Fatalf("buffer has %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestFloat64DecodeNaN(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf8, 0x7f,
	})

	got, err := Float64Decode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsNaN(got) {
		t.Fatal("got non-NaN value")
	}
	if math.Float64bits(got) != 0x7ff8000000000001 {
		t.Fatalf("got bits %016x, want 7ff8000000000001", math.Float64bits(got))
	}
}

func TestFloat64DecodeLeavesFollowingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf0, 0x3f,
		0xaa, 0xbb,
	})

	got, err := Float64Decode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
	if !bytes.Equal(buffer.Bytes(), []byte{0xaa, 0xbb}) {
		t.Fatalf("got remaining %x, want aabb", buffer.Bytes())
	}
}

func TestFloat64DecodeInsufficientBytes(t *testing.T) {
	tests := [][]byte{
		nil,
		{0x00},
		{0x00, 0x00, 0x00, 0x00},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	for _, input := range tests {
		t.Run("", func(t *testing.T) {
			buffer := bytes.NewBuffer(input)

			got, err := Float64Decode(buffer)

			if got != 0 {
				t.Fatalf("got %v, want 0", got)
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != "insufficient bytes for float64" {
				t.Fatalf("got %q", err.Error())
			}
			if !bytes.Equal(buffer.Bytes(), input) {
				t.Fatal("decoder consumed bytes on error")
			}
		})
	}
}

func TestFloat32EncodeDecodeRoundTrip(t *testing.T) {
	values := []float32{
		0,
		float32(math.Copysign(0, -1)),
		1,
		-1,
		0.5,
		-0.5,
		float32(math.Pi),
		math.MaxFloat32,
		math.SmallestNonzeroFloat32,
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
		math.Float32frombits(0x7fc00001),
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := Float32Encode(value)

			got, err := Float32Decode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if math.Float32bits(got) != math.Float32bits(value) {
				t.Fatalf("got %08x, want %08x", math.Float32bits(got), math.Float32bits(value))
			}
		})
	}
}

func TestFloat64EncodeDecodeRoundTrip(t *testing.T) {
	values := []float64{
		0,
		math.Copysign(0, -1),
		1,
		-1,
		0.5,
		-0.5,
		math.Pi,
		math.MaxFloat64,
		math.SmallestNonzeroFloat64,
		math.Inf(1),
		math.Inf(-1),
		math.Float64frombits(0x7ff8000000000001),
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := Float64Encode(value)

			got, err := Float64Decode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if math.Float64bits(got) != math.Float64bits(value) {
				t.Fatalf("got %016x, want %016x", math.Float64bits(got), math.Float64bits(value))
			}
		})
	}
}

func TestFloat32InlineEncodeDecodeRoundTrip(t *testing.T) {
	values := []float32{
		0,
		float32(math.Copysign(0, -1)),
		1,
		-1,
		float32(math.Pi),
		math.MaxFloat32,
		math.SmallestNonzeroFloat32,
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
		math.Float32frombits(0x7fc00001),
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			Float32InlineEncode(value, &buffer)

			got, err := Float32Decode(&buffer)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if math.Float32bits(got) != math.Float32bits(value) {
				t.Fatalf("got %08x, want %08x", math.Float32bits(got), math.Float32bits(value))
			}
		})
	}
}

func TestFloat64InlineEncodeDecodeRoundTrip(t *testing.T) {
	values := []float64{
		0,
		math.Copysign(0, -1),
		1,
		-1,
		math.Pi,
		math.MaxFloat64,
		math.SmallestNonzeroFloat64,
		math.Inf(1),
		math.Inf(-1),
		math.Float64frombits(0x7ff8000000000001),
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			Float64InlineEncode(value, &buffer)

			got, err := Float64Decode(&buffer)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if math.Float64bits(got) != math.Float64bits(value) {
				t.Fatalf("got %016x, want %016x", math.Float64bits(got), math.Float64bits(value))
			}
		})
	}
}
