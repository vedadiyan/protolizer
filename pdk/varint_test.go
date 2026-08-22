package pdk

import (
	"bytes"
	"testing"
)

func TestVarintEncode(t *testing.T) {
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
			name:  "255",
			value: 255,
			want:  []byte{0xff, 0x01},
		},
		{
			name:  "16384",
			value: 16384,
			want:  []byte{0x80, 0x80, 0x01},
		},
		{
			name:  "max int64",
			value: 1<<63 - 1,
			want: []byte{
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0x7f,
			},
		},
		{
			name:  "min int64",
			value: -1 << 63,
			want: []byte{
				0x80, 0x80, 0x80, 0x80, 0x80,
				0x80, 0x80, 0x80, 0x80, 0x01,
			},
		},
		{
			name:  "negative one",
			value: -1,
			want: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VarintEncode(tt.value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestUvarintEncode(t *testing.T) {
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
			name:  "255",
			value: 255,
			want:  []byte{0xff, 0x01},
		},
		{
			name:  "256",
			value: 256,
			want:  []byte{0x80, 0x02},
		},
		{
			name:  "16383",
			value: 16383,
			want:  []byte{0xff, 0x7f},
		},
		{
			name:  "16384",
			value: 16384,
			want:  []byte{0x80, 0x80, 0x01},
		},
		{
			name:  "max uint64",
			value: ^uint64(0),
			want: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UvarintEncode(tt.value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestUvarintInlineEncode(t *testing.T) {
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
			name:  "16384",
			value: 16384,
			want:  []byte{0x80, 0x80, 0x01},
		},
		{
			name:  "max uint64",
			value: ^uint64(0),
			want: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			UvarintInlineEncode(tt.value, &buf)

			if !bytes.Equal(buf.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buf.Bytes(), tt.want)
			}
		})
	}
}

func TestUvarintInlineEncodeAppends(t *testing.T) {
	var buf bytes.Buffer

	buf.Write([]byte{0xaa, 0xbb})

	UvarintInlineEncode(300, &buf)

	want := []byte{
		0xaa,
		0xbb,
		0xac,
		0x02,
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("got %x, want %x", buf.Bytes(), want)
	}
}

func TestUvarint(t *testing.T) {
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
			name:  "16384",
			value: 16384,
			want:  []byte{0x80, 0x80, 0x01},
		},
		{
			name:  "max uint64",
			value: ^uint64(0),
			want: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			uvarint(tt.value, &buf)

			if !bytes.Equal(buf.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buf.Bytes(), tt.want)
			}
		})
	}
}

func TestUvarintDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  uint64
	}{
		{
			name:  "zero",
			input: []byte{0x00},
			want:  0,
		},
		{
			name:  "one",
			input: []byte{0x01},
			want:  1,
		},
		{
			name:  "127",
			input: []byte{0x7f},
			want:  127,
		},
		{
			name:  "128",
			input: []byte{0x80, 0x01},
			want:  128,
		},
		{
			name:  "300",
			input: []byte{0xac, 0x02},
			want:  300,
		},
		{
			name:  "16384",
			input: []byte{0x80, 0x80, 0x01},
			want:  16384,
		},
		{
			name: "max uint64",
			input: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
			want: ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)

			got, err := UvarintDecode(buf)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if buf.Len() != 0 {
				t.Fatalf("buffer still has %d bytes", buf.Len())
			}
		})
	}
}

func TestUvarintDecodeLeavesFollowingBytes(t *testing.T) {
	buf := bytes.NewBuffer([]byte{
		0xac, 0x02,
		0xaa, 0xbb,
	})

	got, err := UvarintDecode(buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 300 {
		t.Fatalf("got %d, want 300", got)
	}

	want := []byte{0xaa, 0xbb}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("remaining bytes: got %x, want %x", buf.Bytes(), want)
	}
}

func TestUvarintDecodeTruncatedEmpty(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	got, err := UvarintDecode(buf)

	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}

	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "truncated varint" {
		t.Fatalf("got %q, want %q", err.Error(), "truncated varint")
	}
}

func TestUvarintDecodeTruncated(t *testing.T) {
	tests := [][]byte{
		{0x80},
		{0x80, 0x80},
		{0xff, 0xff, 0xff},
		{0x80, 0x80, 0x80, 0x80},
	}

	for _, input := range tests {
		t.Run("", func(t *testing.T) {
			buf := bytes.NewBuffer(input)

			got, err := UvarintDecode(buf)

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}

			if err == nil {
				t.Fatal("expected truncated varint error")
			}

			if err.Error() != "truncated varint" {
				t.Fatalf("got %q, want %q", err.Error(), "truncated varint")
			}
		})
	}
}

func TestUvarintDecodeOverflow(t *testing.T) {
	tests := [][]byte{
		{
			0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0x02,
		},
		{
			0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff,
		},
	}

	for _, input := range tests {
		t.Run("", func(t *testing.T) {
			buf := bytes.NewBuffer(input)

			got, err := UvarintDecode(buf)

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}

			if err == nil {
				t.Fatal("expected overflow error")
			}

			if err.Error() != "varint overflows uint64" {
				t.Fatalf("got %q, want %q", err.Error(), "varint overflows uint64")
			}
		})
	}
}

func TestVarintDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int64
	}{
		{
			name:  "zero",
			input: []byte{0x00},
			want:  0,
		},
		{
			name:  "positive",
			input: []byte{0xac, 0x02},
			want:  300,
		},
		{
			name: "negative one",
			input: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
			want: -1,
		},
		{
			name: "min int64",
			input: []byte{
				0x80, 0x80, 0x80, 0x80, 0x80,
				0x80, 0x80, 0x80, 0x80, 0x01,
			},
			want: -1 << 63,
		},
		{
			name: "max int64",
			input: []byte{
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0x7f,
			},
			want: 1<<63 - 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)

			got, err := VarintDecode(buf)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestVarintDecodeTruncated(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x80})

	got, err := VarintDecode(buf)

	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}

	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "truncated varint" {
		t.Fatalf("got %q, want %q", err.Error(), "truncated varint")
	}
}

func TestVarintDecodeOverflow(t *testing.T) {
	buf := bytes.NewBuffer([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0x02,
	})

	got, err := VarintDecode(buf)

	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}

	if err == nil {
		t.Fatal("expected overflow error")
	}

	if err.Error() != "varint overflows uint64" {
		t.Fatalf("got %q, want %q", err.Error(), "varint overflows uint64")
	}
}

func TestUvarintPeek(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		wantN int
		want  uint64
	}{
		{
			name:  "zero",
			input: []byte{0x00},
			wantN: 1,
			want:  0,
		},
		{
			name:  "one byte",
			input: []byte{0x7f},
			wantN: 1,
			want:  127,
		},
		{
			name:  "two bytes",
			input: []byte{0x80, 0x01},
			wantN: 2,
			want:  128,
		},
		{
			name:  "three bytes",
			input: []byte{0x80, 0x80, 0x01},
			wantN: 3,
			want:  16384,
		},
		{
			name: "max uint64",
			input: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
			wantN: 10,
			want:  ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]byte(nil), tt.input...)
			buf := bytes.NewBuffer(input)

			n, got, err := UvarintPeek(buf)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if n != tt.wantN {
				t.Fatalf("got length %d, want %d", n, tt.wantN)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if !bytes.Equal(buf.Bytes(), input) {
				t.Fatalf("Peek changed buffer")
			}
		})
	}
}

func TestUvarintPeekWithFollowingData(t *testing.T) {
	input := []byte{
		0xac, 0x02,
		0xaa, 0xbb,
	}

	buf := bytes.NewBuffer(input)

	n, got, err := UvarintPeek(buf)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != 2 {
		t.Fatalf("got length %d, want 2", n)
	}

	if got != 300 {
		t.Fatalf("got %d, want 300", got)
	}

	if !bytes.Equal(buf.Bytes(), input) {
		t.Fatal("Peek consumed data")
	}
}

func TestUvarintPeekTruncatedEmpty(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	n, got, err := UvarintPeek(buf)

	if n != 0 {
		t.Fatalf("got length %d, want 0", n)
	}

	if got != 0 {
		t.Fatalf("got value %d, want 0", got)
	}

	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "truncated varint" {
		t.Fatalf("got %q, want %q", err.Error(), "truncated varint")
	}
}

func TestUvarintPeekTruncated(t *testing.T) {
	tests := [][]byte{
		{0x80},
		{0x80, 0x80},
		{0xff, 0xff, 0xff},
	}

	for _, input := range tests {
		t.Run("", func(t *testing.T) {
			buf := bytes.NewBuffer(input)

			n, got, err := UvarintPeek(buf)

			if n != 0 {
				t.Fatalf("got length %d, want 0", n)
			}

			if got != 0 {
				t.Fatalf("got value %d, want 0", got)
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != "truncated varint" {
				t.Fatalf("got %q, want %q", err.Error(), "truncated varint")
			}

			if !bytes.Equal(buf.Bytes(), input) {
				t.Fatal("Peek consumed data")
			}
		})
	}
}

func TestUvarintPeekOverflow(t *testing.T) {
	tests := [][]byte{
		{
			0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0x02,
		},
		{
			0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff,
		},
	}

	for _, input := range tests {
		t.Run("", func(t *testing.T) {
			buf := bytes.NewBuffer(input)

			n, got, err := UvarintPeek(buf)

			if n != 0 {
				t.Fatalf("got length %d, want 0", n)
			}

			if got != 0 {
				t.Fatalf("got value %d, want 0", got)
			}

			if err == nil {
				t.Fatal("expected overflow error")
			}

			if err.Error() != "varint overflows uint64" {
				t.Fatalf("got %q, want %q", err.Error(), "varint overflows uint64")
			}

			if !bytes.Equal(buf.Bytes(), input) {
				t.Fatal("Peek consumed data")
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	values := []uint64{
		0,
		1,
		127,
		128,
		255,
		256,
		16383,
		16384,
		1 << 21,
		1 << 32,
		1 << 63,
		^uint64(0),
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := UvarintEncode(value)

			got, err := UvarintDecode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestVarintEncodeDecodeRoundTrip(t *testing.T) {
	values := []int64{
		0,
		1,
		127,
		128,
		16384,
		-1,
		-2,
		-127,
		-128,
		-16384,
		-1 << 63,
		1<<63 - 1,
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := VarintEncode(value)

			got, err := VarintDecode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}
		})
	}
}

func TestPeekMatchesDecode(t *testing.T) {
	values := []uint64{
		0,
		1,
		127,
		128,
		16384,
		1 << 32,
		^uint64(0),
	}

	for _, value := range values {
		t.Run("", func(t *testing.T) {
			encoded := UvarintEncode(value)

			peekBuf := bytes.NewBuffer(
				append([]byte(nil), encoded.Bytes()...),
			)

			n, peekValue, err := UvarintPeek(peekBuf)

			if err != nil {
				t.Fatalf("peek failed: %v", err)
			}

			if peekValue != value {
				t.Fatalf("peek got %d, want %d", peekValue, value)
			}

			if n != encoded.Len() {
				t.Fatalf("peek length %d, want %d", n, encoded.Len())
			}

			decodeBuf := bytes.NewBuffer(encoded.Bytes())

			decodeValue, err := UvarintDecode(decodeBuf)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if decodeValue != peekValue {
				t.Fatalf(
					"peek %d differs from decode %d",
					peekValue,
					decodeValue,
				)
			}
		})
	}
}
