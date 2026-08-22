package pdk

import (
	"bytes"
	"testing"
)

func TestBoolEncode(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  []byte
	}{
		{
			name:  "true",
			value: true,
			want:  []byte{0x01},
		},
		{
			name:  "false",
			value: false,
			want:  []byte{0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoolEncode(tt.value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestBoolInlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  []byte
	}{
		{
			name:  "true",
			value: true,
			want:  []byte{0x01},
		},
		{
			name:  "false",
			value: false,
			want:  []byte{0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			BoolInlineEncode(tt.value, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestBoolInlineEncodeAppends(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  []byte
	}{
		{
			name:  "append true",
			value: true,
			want:  []byte{0xaa, 0xbb, 0x01},
		},
		{
			name:  "append false",
			value: false,
			want:  []byte{0xaa, 0xbb, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			buffer.Write([]byte{0xaa, 0xbb})

			BoolInlineEncode(tt.value, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestBoolDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{
			name:  "zero is false",
			input: []byte{0x00},
			want:  false,
		},
		{
			name:  "one is true",
			input: []byte{0x01},
			want:  true,
		},
		{
			name:  "nonzero is true",
			input: []byte{0x02},
			want:  true,
		},
		{
			name:  "127 is true",
			input: []byte{0x7f},
			want:  true,
		},
		{
			name:  "multi byte nonzero is true",
			input: []byte{0x80, 0x01},
			want:  true,
		},
		{
			name: "negative one is true",
			input: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := BoolDecode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("buffer still has %d bytes", buffer.Len())
			}
		})
	}
}

func TestBoolDecodeLeavesFollowingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0x01,
		0xaa,
		0xbb,
	})

	got, err := BoolDecode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got {
		t.Fatal("got false, want true")
	}

	want := []byte{0xaa, 0xbb}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf(
			"remaining data: got %x, want %x",
			buffer.Bytes(),
			want,
		)
	}
}

func TestBoolDecodeTruncated(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: nil,
		},
		{
			name:  "continuation byte",
			input: []byte{0x80},
		},
		{
			name:  "truncated multi byte",
			input: []byte{0xac},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := BoolDecode(buffer)

			if got {
				t.Fatalf("got true, want false")
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != "truncated varint" {
				t.Fatalf(
					"got error %q, want %q",
					err.Error(),
					"truncated varint",
				)
			}
		})
	}
}

func TestBoolDecodeOverflow(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0x02,
	})

	got, err := BoolDecode(buffer)

	if got {
		t.Fatal("got true, want false")
	}

	if err == nil {
		t.Fatal("expected overflow error")
	}

	if err.Error() != "varint overflows uint64" {
		t.Fatalf(
			"got error %q, want %q",
			err.Error(),
			"varint overflows uint64",
		)
	}
}

func TestBoolEncodeDecodeRoundTrip(t *testing.T) {
	tests := []bool{
		false,
		true,
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			encoded := BoolEncode(value)

			got, err := BoolDecode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %v, want %v", got, value)
			}
		})
	}
}

func TestBoolInlineEncodeDecodeRoundTrip(t *testing.T) {
	tests := []bool{
		false,
		true,
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			BoolInlineEncode(value, &buffer)

			got, err := BoolDecode(&buffer)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %v, want %v", got, value)
			}
		})
	}
}

func TestBoolEncodeMatchesBoolInlineEncode(t *testing.T) {
	tests := []bool{
		false,
		true,
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			expected := BoolEncode(value)

			var actual bytes.Buffer
			BoolInlineEncode(value, &actual)

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
