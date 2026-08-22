package pdk

import (
	"bytes"
	"io"
	"testing"
)

func TestTagEncode(t *testing.T) {
	tests := []struct {
		name        string
		fieldNumber int32
		wireType    WireType
		want        []byte
	}{
		{"varint", 1, WireTypeVarint, []byte{0x08}},
		{"i64", 1, WireTypeI64, []byte{0x09}},
		{"len", 1, WireTypeLen, []byte{0x0a}},
		{"start group", 1, WireTypeSGroup, []byte{0x0b}},
		{"end group", 1, WireTypeEGroup, []byte{0x0c}},
		{"i32", 1, WireTypeI32, []byte{0x0d}},
		{"field 15", 15, WireTypeVarint, []byte{0x78}},
		{"field 16", 16, WireTypeVarint, []byte{0x80, 0x01}},
		{"field 127", 127, WireTypeLen, []byte{0xfa, 0x07}},
		{"field 128", 128, WireTypeI32, []byte{0x85, 0x08}},
		{"field 1000", 1000, WireTypeVarint, []byte{0xc0, 0x3e}},
		{
			"max field number",
			2147483647,
			WireTypeI32,
			[]byte{0xfd, 0xff, 0xff, 0xff, 0x3f},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TagEncode(tt.fieldNumber, tt.wireType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestTagEncodeErrors(t *testing.T) {
	tests := []struct {
		name        string
		fieldNumber int32
		wireType    WireType
	}{
		{"zero field number", 0, WireTypeVarint},
		{"negative field number", -1, WireTypeVarint},
		{"wire type 6", 1, WireType(6)},
		{"wire type 7", 1, WireType(7)},
		{"wire type 255", 1, WireType(255)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TagEncode(tt.fieldNumber, tt.wireType)

			if err == nil {
				t.Fatal("expected error")
			}

			if got != nil {
				t.Fatalf("expected nil buffer, got %x", got.Bytes())
			}
		})
	}
}

func TestTagInlineEncode(t *testing.T) {
	tests := []struct {
		name        string
		prefix      []byte
		fieldNumber int32
		wireType    WireType
		want        []byte
	}{
		{
			"varint",
			nil,
			1,
			WireTypeVarint,
			[]byte{0x08},
		},
		{
			"i64",
			[]byte{0xaa, 0xbb},
			1,
			WireTypeI64,
			[]byte{0xaa, 0xbb, 0x09},
		},
		{
			"len",
			nil,
			1,
			WireTypeLen,
			[]byte{0x0a},
		},
		{
			"start group",
			nil,
			1,
			WireTypeSGroup,
			[]byte{0x0b},
		},
		{
			"end group",
			nil,
			1,
			WireTypeEGroup,
			[]byte{0x0c},
		},
		{
			"i32",
			nil,
			1,
			WireTypeI32,
			[]byte{0x0d},
		},
		{
			"multi byte tag",
			[]byte{0x01},
			1000,
			WireTypeVarint,
			[]byte{0x01, 0xc0, 0x3e},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(append([]byte(nil), tt.prefix...))

			err := TagInlineEncode(
				tt.fieldNumber,
				tt.wireType,
				buffer,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestTagInlineEncodeErrors(t *testing.T) {
	tests := []struct {
		name        string
		fieldNumber int32
		wireType    WireType
	}{
		{"zero field", 0, WireTypeVarint},
		{"negative field", -1, WireTypeVarint},
		{"wire type 6", 1, WireType(6)},
		{"wire type 7", 1, WireType(7)},
		{"wire type 255", 1, WireType(255)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{0xaa, 0xbb})

			err := TagInlineEncode(
				tt.fieldNumber,
				tt.wireType,
				buffer,
			)

			if err == nil {
				t.Fatal("expected error")
			}

			if !bytes.Equal(buffer.Bytes(), []byte{0xaa, 0xbb}) {
				t.Fatalf("buffer changed: %x", buffer.Bytes())
			}
		})
	}
}

func TestTagDecode(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		wantField  int32
		wantWire   WireType
		wantRemain int
	}{
		{"varint", []byte{0x08}, 1, WireTypeVarint, 0},
		{"i64", []byte{0x09}, 1, WireTypeI64, 0},
		{"len", []byte{0x0a}, 1, WireTypeLen, 0},
		{"start group", []byte{0x0b}, 1, WireTypeSGroup, 0},
		{"end group", []byte{0x0c}, 1, WireTypeEGroup, 0},
		{"i32", []byte{0x0d}, 1, WireTypeI32, 0},
		{"field 16", []byte{0x80, 0x01}, 16, WireTypeVarint, 0},
		{"field 1000", []byte{0xc0, 0x3e}, 1000, WireTypeVarint, 0},
		{
			"remaining data",
			[]byte{0x08, 0xaa, 0xbb},
			1,
			WireTypeVarint,
			2,
		},
		{
			"max field number",
			[]byte{0xfd, 0xff, 0xff, 0xff, 0x3f},
			2147483647,
			WireTypeI32,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			field, wire, err := TagDecode(buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if field != tt.wantField {
				t.Fatalf("got field %d, want %d", field, tt.wantField)
			}

			if wire != tt.wantWire {
				t.Fatalf("got wire %d, want %d", wire, tt.wantWire)
			}

			if buffer.Len() != tt.wantRemain {
				t.Fatalf(
					"got %d remaining bytes, want %d",
					buffer.Len(),
					tt.wantRemain,
				)
			}
		})
	}
}

func TestTagDecodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", nil},
		{"truncated", []byte{0x80}},
		{"zero tag", []byte{0x00}},
		{"zero field", []byte{0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			field, wire, err := TagDecode(buffer)
			if err == nil {
				t.Fatal("expected error")
			}

			if field != 0 {
				t.Fatalf("got field %d, want 0", field)
			}

			if wire != 0 {
				t.Fatalf("got wire %d, want 0", wire)
			}
		})
	}
}

func TestTagPeek(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		wantField int32
		wantWire  WireType
		wantLen   int
	}{
		{
			"single byte",
			[]byte{0x08, 0xaa},
			1,
			WireTypeVarint,
			1,
		},
		{
			"multi byte",
			[]byte{0x80, 0x01, 0xaa},
			16,
			WireTypeVarint,
			2,
		},
		{
			"large field",
			[]byte{0xc0, 0x3e, 0xaa},
			1000,
			WireTypeVarint,
			2,
		},
		{
			"i64",
			[]byte{0x09, 0xaa},
			1,
			WireTypeI64,
			1,
		},
		{
			"len",
			[]byte{0x0a, 0xaa},
			1,
			WireTypeLen,
			1,
		},
		{
			"start group",
			[]byte{0x0b, 0xaa},
			1,
			WireTypeSGroup,
			1,
		},
		{
			"end group",
			[]byte{0x0c, 0xaa},
			1,
			WireTypeEGroup,
			1,
		},
		{
			"i32",
			[]byte{0x0d, 0xaa},
			1,
			WireTypeI32,
			1,
		},
		{
			"max field number",
			[]byte{0xfd, 0xff, 0xff, 0xff, 0x3f, 0xaa},
			2147483647,
			WireTypeI32,
			5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)
			original := append([]byte(nil), tt.input...)

			field, wire, consume, err := TagPeek(buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if field != tt.wantField {
				t.Fatalf("got field %d, want %d", field, tt.wantField)
			}

			if wire != tt.wantWire {
				t.Fatalf("got wire %d, want %d", wire, tt.wantWire)
			}

			if consume == nil {
				t.Fatal("expected consume function")
			}

			if !bytes.Equal(buffer.Bytes(), original) {
				t.Fatalf(
					"peek consumed data: got %x, want %x",
					buffer.Bytes(),
					original,
				)
			}

			consume()

			if !bytes.Equal(
				buffer.Bytes(),
				original[tt.wantLen:],
			) {
				t.Fatalf(
					"after consume got %x, want %x",
					buffer.Bytes(),
					original[tt.wantLen:],
				)
			}
		})
	}
}

func TestTagPeekEmpty(t *testing.T) {
	buffer := bytes.NewBuffer(nil)

	field, wire, consume, err := TagPeek(buffer)

	if err != io.EOF {
		t.Fatalf("got %v, want io.EOF", err)
	}

	if field != 0 {
		t.Fatalf("got field %d, want 0", field)
	}

	if wire != 0 {
		t.Fatalf("got wire %d, want 0", wire)
	}

	if consume != nil {
		t.Fatal("expected nil consume function")
	}

	if buffer.Len() != 0 {
		t.Fatalf("got %d bytes remaining, want 0", buffer.Len())
	}
}

func TestTagPeekErrors(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"truncated", []byte{0x80}},
		{"zero tag", []byte{0x00}},
		{"zero field", []byte{0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]byte(nil), tt.input...)
			buffer := bytes.NewBuffer(input)

			field, wire, consume, err := TagPeek(buffer)
			if err == nil {
				t.Fatal("expected error")
			}

			if field != 0 {
				t.Fatalf("got field %d, want 0", field)
			}

			if wire != 0 {
				t.Fatalf("got wire %d, want 0", wire)
			}

			if consume != nil {
				t.Fatal("expected nil consume function")
			}

			if !bytes.Equal(buffer.Bytes(), input) {
				t.Fatalf(
					"buffer changed: got %x, want %x",
					buffer.Bytes(),
					input,
				)
			}
		})
	}
}

func TestTagRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		fieldNumber int32
		wireType    WireType
	}{
		{"varint", 1, WireTypeVarint},
		{"i64", 1, WireTypeI64},
		{"len", 1, WireTypeLen},
		{"start group", 1, WireTypeSGroup},
		{"end group", 1, WireTypeEGroup},
		{"i32", 1, WireTypeI32},
		{"field 16", 16, WireTypeVarint},
		{"field 1000", 1000, WireTypeI64},
		{"max field", 2147483647, WireTypeI32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer, err := TagEncode(
				tt.fieldNumber,
				tt.wireType,
			)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			field, wire, err := TagDecode(buffer)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if field != tt.fieldNumber {
				t.Fatalf(
					"got field %d, want %d",
					field,
					tt.fieldNumber,
				)
			}

			if wire != tt.wireType {
				t.Fatalf(
					"got wire %d, want %d",
					wire,
					tt.wireType,
				)
			}

			if buffer.Len() != 0 {
				t.Fatalf("got %d remaining bytes", buffer.Len())
			}
		})
	}
}

func TestTagInlineRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		fieldNumber int32
		wireType    WireType
	}{
		{"varint", 1, WireTypeVarint},
		{"i64", 16, WireTypeI64},
		{"len", 1000, WireTypeLen},
		{"start group", 1, WireTypeSGroup},
		{"end group", 1, WireTypeEGroup},
		{"i32", 2147483647, WireTypeI32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			err := TagInlineEncode(
				tt.fieldNumber,
				tt.wireType,
				&buffer,
			)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			field, wire, err := TagDecode(&buffer)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if field != tt.fieldNumber {
				t.Fatalf(
					"got field %d, want %d",
					field,
					tt.fieldNumber,
				)
			}

			if wire != tt.wireType {
				t.Fatalf(
					"got wire %d, want %d",
					wire,
					tt.wireType,
				)
			}
		})
	}
}

func TestTagPeekRoundTrip(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{
		0xc0, 0x3e,
		0xaa, 0xbb,
	})

	field, wire, consume, err := TagPeek(buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field != 1000 {
		t.Fatalf("got field %d, want 1000", field)
	}

	if wire != WireTypeVarint {
		t.Fatalf(
			"got wire %d, want %d",
			wire,
			WireTypeVarint,
		)
	}

	if consume == nil {
		t.Fatal("expected consume function")
	}

	consume()

	if !bytes.Equal(buffer.Bytes(), []byte{0xaa, 0xbb}) {
		t.Fatalf(
			"got %x, want %x",
			buffer.Bytes(),
			[]byte{0xaa, 0xbb},
		)
	}
}
