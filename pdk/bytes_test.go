package pdk

import (
	"bytes"
	"testing"
)

func TestBytesEncode(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  []byte
	}{
		{
			name:  "nil",
			value: nil,
			want:  []byte{0x00},
		},
		{
			name:  "empty",
			value: []byte{},
			want:  []byte{0x00},
		},
		{
			name:  "single byte",
			value: []byte{0xaa},
			want:  []byte{0x01, 0xaa},
		},
		{
			name:  "multiple bytes",
			value: []byte{0xaa, 0xbb, 0xcc},
			want:  []byte{0x03, 0xaa, 0xbb, 0xcc},
		},
		{
			name:  "128 bytes",
			value: bytes.Repeat([]byte{0xaa}, 128),
			want: append(
				[]byte{0x80, 0x01},
				bytes.Repeat([]byte{0xaa}, 128)...,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesEncode(tt.value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestBufferEncode(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  []byte
	}{
		{
			name:  "empty",
			value: nil,
			want:  []byte{0x00},
		},
		{
			name:  "single byte",
			value: []byte{0xaa},
			want:  []byte{0x01, 0xaa},
		},
		{
			name:  "multiple bytes",
			value: []byte{0xaa, 0xbb, 0xcc},
			want:  []byte{0x03, 0xaa, 0xbb, 0xcc},
		},
		{
			name:  "128 bytes",
			value: bytes.Repeat([]byte{0xaa}, 128),
			want: append(
				[]byte{0x80, 0x01},
				bytes.Repeat([]byte{0xaa}, 128)...,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := bytes.NewBuffer(append([]byte(nil), tt.value...))

			got := BufferEncode(value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}

			if value.Len() != 0 {
				t.Fatalf("source buffer has %d bytes remaining", value.Len())
			}
		})
	}
}

func TestBufferInlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  []byte
	}{
		{
			name:  "empty",
			value: nil,
			want:  []byte{0x00},
		},
		{
			name:  "single byte",
			value: []byte{0xaa},
			want:  []byte{0x01, 0xaa},
		},
		{
			name:  "multiple bytes",
			value: []byte{0xaa, 0xbb, 0xcc},
			want:  []byte{0x03, 0xaa, 0xbb, 0xcc},
		},
		{
			name:  "128 bytes",
			value: bytes.Repeat([]byte{0xaa}, 128),
			want: append(
				[]byte{0x80, 0x01},
				bytes.Repeat([]byte{0xaa}, 128)...,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := bytes.NewBuffer(
				append([]byte(nil), tt.value...),
			)

			var dest bytes.Buffer

			BufferInlineEncode(source, &dest)

			if !bytes.Equal(dest.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", dest.Bytes(), tt.want)
			}

			if source.Len() != 0 {
				t.Fatalf("source buffer has %d bytes remaining", source.Len())
			}
		})
	}
}

func TestBufferInlineEncodeAppends(t *testing.T) {
	source := bytes.NewBuffer([]byte{0xaa, 0xbb})

	var dest bytes.Buffer
	dest.Write([]byte{0x11, 0x22})

	BufferInlineEncode(source, &dest)

	want := []byte{
		0x11,
		0x22,
		0x02,
		0xaa,
		0xbb,
	}

	if !bytes.Equal(dest.Bytes(), want) {
		t.Fatalf("got %x, want %x", dest.Bytes(), want)
	}

	if source.Len() != 0 {
		t.Fatalf("source buffer has %d bytes remaining", source.Len())
	}
}

func TestStringEncode(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []byte
	}{
		{
			name:  "empty",
			value: "",
			want:  []byte{0x00},
		},
		{
			name:  "ascii",
			value: "hello",
			want:  []byte{0x05, 'h', 'e', 'l', 'l', 'o'},
		},
		{
			name:  "unicode",
			value: "hello 世界",
			want: append(
				[]byte{0x0c},
				[]byte("hello 世界")...,
			),
		},
		{
			name:  "unicode only",
			value: "世界",
			want: append(
				[]byte{0x06},
				[]byte("世界")...,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringEncode(tt.value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestStringInlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []byte
	}{
		{
			name:  "empty",
			value: "",
			want:  []byte{0x00},
		},
		{
			name:  "ascii",
			value: "hello",
			want:  []byte{0x05, 'h', 'e', 'l', 'l', 'o'},
		},
		{
			name:  "unicode",
			value: "hello 世界",
			want: append(
				[]byte{0x0c},
				[]byte("hello 世界")...,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			StringInlineEncode(tt.value, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestStringInlineEncodeAppends(t *testing.T) {
	var buffer bytes.Buffer

	buffer.Write([]byte{0xaa, 0xbb})

	StringInlineEncode("hello", &buffer)

	want := []byte{
		0xaa,
		0xbb,
		0x05,
		'h',
		'e',
		'l',
		'l',
		'o',
	}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestBytesDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{
			name:  "zero length",
			input: []byte{0x00},
			want:  []byte{},
		},
		{
			name:  "single byte",
			input: []byte{0x01, 0xaa},
			want:  []byte{0xaa},
		},
		{
			name:  "multiple bytes",
			input: []byte{0x03, 0xaa, 0xbb, 0xcc},
			want:  []byte{0xaa, 0xbb, 0xcc},
		},
		{
			name: "128 bytes",
			input: append(
				[]byte{0x80, 0x01},
				bytes.Repeat([]byte{0xaa}, 128)...,
			),
			want: bytes.Repeat([]byte{0xaa}, 128),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := BytesDecode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(got, tt.want) {
				t.Fatalf("got %x, want %x", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("buffer has %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestBytesDecodeLeavesFollowingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0x03,
		0xaa,
		0xbb,
		0xcc,
		0x11,
		0x22,
	})

	got, err := BytesDecode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(got, []byte{0xaa, 0xbb, 0xcc}) {
		t.Fatalf("got %x", got)
	}

	wantRemaining := []byte{0x11, 0x22}

	if !bytes.Equal(buffer.Bytes(), wantRemaining) {
		t.Fatalf(
			"remaining bytes: got %x, want %x",
			buffer.Bytes(),
			wantRemaining,
		)
	}
}

func TestBytesDecodeTruncatedLength(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: nil,
		},
		{
			name:  "truncated varint",
			input: []byte{0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := BytesDecode(buffer)

			if got != nil {
				t.Fatalf("got %x, want nil", got)
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != "truncated varint" {
				t.Fatalf(
					"got %q, want %q",
					err.Error(),
					"truncated varint",
				)
			}
		})
	}
}

func TestBytesDecodeNegativeLength(t *testing.T) {
	// Varint encoding of -1.
	input := []byte{
		0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0x01,
	}

	buffer := bytes.NewBuffer(input)

	got, err := BytesDecode(buffer)

	if got != nil {
		t.Fatalf("got %x, want nil", got)
	}

	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "negative length" {
		t.Fatalf(
			"got %q, want %q",
			err.Error(),
			"negative length",
		)
	}
}

func TestBytesDecodeInsufficientBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "length greater than available",
			input: []byte{0x05, 0xaa, 0xbb},
		},
		{
			name:  "length one but no data",
			input: []byte{0x01},
		},
		{
			name:  "length two but one byte",
			input: []byte{0x02, 0xaa},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := BytesDecode(buffer)

			if got != nil {
				t.Fatalf("got %x, want nil", got)
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != "insufficient bytes for length-prefixed data" {
				t.Fatalf(
					"got %q, want %q",
					err.Error(),
					"insufficient bytes for length-prefixed data",
				)
			}
		})
	}
}

func TestStringDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty",
			input: []byte{0x00},
			want:  "",
		},
		{
			name:  "ascii",
			input: []byte{0x05, 'h', 'e', 'l', 'l', 'o'},
			want:  "hello",
		},
		{
			name: "unicode",
			input: append(
				[]byte{0x0c},
				[]byte("hello 世界")...,
			),
			want: "hello 世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := StringDecode(buffer)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("buffer has %d bytes remaining", buffer.Len())
			}
		})
	}
}

func TestStringDecodeLeavesFollowingData(t *testing.T) {
	input := append(
		[]byte{0x05},
		[]byte("hello")...,
	)
	input = append(input, 0xaa, 0xbb)

	buffer := bytes.NewBuffer(input)

	got, err := StringDecode(buffer)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}

	wantRemaining := []byte{0xaa, 0xbb}

	if !bytes.Equal(buffer.Bytes(), wantRemaining) {
		t.Fatalf(
			"remaining bytes: got %x, want %x",
			buffer.Bytes(),
			wantRemaining,
		)
	}
}

func TestStringDecodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr string
	}{
		{
			name:    "truncated varint",
			input:   []byte{0x80},
			wantErr: "truncated varint",
		},
		{
			name: "negative length",
			input: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x01,
			},
			wantErr: "negative length",
		},
		{
			name:    "insufficient data",
			input:   []byte{0x03, 'a'},
			wantErr: "insufficient bytes for length-prefixed data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := StringDecode(buffer)

			if got != "" {
				t.Fatalf("got %q, want empty string", got)
			}

			if err == nil {
				t.Fatal("expected error")
			}

			if err.Error() != tt.wantErr {
				t.Fatalf(
					"got %q, want %q",
					err.Error(),
					tt.wantErr,
				)
			}
		})
	}
}

func TestBytesEncodeDecodeRoundTrip(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		{0x00},
		{0x01},
		{0xaa, 0xbb, 0xcc},
		bytes.Repeat([]byte{0xaa}, 127),
		bytes.Repeat([]byte{0xbb}, 128),
		bytes.Repeat([]byte{0xcc}, 255),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			encoded := BytesEncode(value)

			got, err := BytesDecode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if !bytes.Equal(got, value) {
				t.Fatalf("got %x, want %x", got, value)
			}
		})
	}
}

func TestBufferEncodeDecodeRoundTrip(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		{0x00},
		{0xaa, 0xbb, 0xcc},
		bytes.Repeat([]byte{0xaa}, 128),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			source := bytes.NewBuffer(
				append([]byte(nil), value...),
			)

			encoded := BufferEncode(source)

			got, err := BytesDecode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if !bytes.Equal(got, value) {
				t.Fatalf("got %x, want %x", got, value)
			}

			if source.Len() != 0 {
				t.Fatalf("source buffer has %d bytes remaining", source.Len())
			}
		})
	}
}

func TestBufferInlineEncodeDecodeRoundTrip(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		{0x00},
		{0xaa, 0xbb, 0xcc},
		bytes.Repeat([]byte{0xaa}, 128),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			source := bytes.NewBuffer(
				append([]byte(nil), value...),
			)

			var dest bytes.Buffer

			BufferInlineEncode(source, &dest)

			got, err := BytesDecode(&dest)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if !bytes.Equal(got, value) {
				t.Fatalf("got %x, want %x", got, value)
			}

			if source.Len() != 0 {
				t.Fatalf("source buffer has %d bytes remaining", source.Len())
			}
		})
	}
}

func TestStringEncodeDecodeRoundTrip(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"hello world",
		"世界",
		"hello 世界",
		"😀",
		string(bytes.Repeat([]byte("a"), 128)),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			encoded := StringEncode(value)

			got, err := StringDecode(encoded)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %q, want %q", got, value)
			}
		})
	}
}

func TestStringInlineEncodeDecodeRoundTrip(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"世界",
		"hello 世界",
		"😀",
		string(bytes.Repeat([]byte("a"), 128)),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			StringInlineEncode(value, &buffer)

			got, err := StringDecode(&buffer)

			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if got != value {
				t.Fatalf("got %q, want %q", got, value)
			}
		})
	}
}

func TestBytesEncodeMatchesBufferEncode(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		{0xaa},
		{0xaa, 0xbb, 0xcc},
		bytes.Repeat([]byte{0xaa}, 128),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			bytesEncoded := BytesEncode(value)

			buffer := bytes.NewBuffer(
				append([]byte(nil), value...),
			)
			bufferEncoded := BufferEncode(buffer)

			if !bytes.Equal(
				bytesEncoded.Bytes(),
				bufferEncoded.Bytes(),
			) {
				t.Fatalf(
					"BytesEncode: %x, BufferEncode: %x",
					bytesEncoded.Bytes(),
					bufferEncoded.Bytes(),
				)
			}
		})
	}
}

func TestStringEncodeMatchesBytesEncode(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"世界",
		"hello 世界",
		"😀",
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			stringEncoded := StringEncode(value)
			bytesEncoded := BytesEncode([]byte(value))

			if !bytes.Equal(
				stringEncoded.Bytes(),
				bytesEncoded.Bytes(),
			) {
				t.Fatalf(
					"StringEncode: %x, BytesEncode: %x",
					stringEncoded.Bytes(),
					bytesEncoded.Bytes(),
				)
			}
		})
	}
}

func TestStringInlineEncodeMatchesStringEncode(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"世界",
		"hello 世界",
		"😀",
		string(bytes.Repeat([]byte("a"), 128)),
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			expected := StringEncode(value)

			var actual bytes.Buffer
			StringInlineEncode(value, &actual)

			if !bytes.Equal(actual.Bytes(), expected.Bytes()) {
				t.Fatalf(
					"got %x, want %x",
					actual.Bytes(),
					expected.Bytes(),
				)
			}
		})
	}
}
