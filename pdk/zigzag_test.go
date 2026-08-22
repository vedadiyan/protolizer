package pdk

import (
	"bytes"
	"math"
	"testing"
)

func TestZigzagEncode(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x02}},
		{"minus one", -1, []byte{0x01}},
		{"two", 2, []byte{0x04}},
		{"minus two", -2, []byte{0x03}},
		{"three", 3, []byte{0x06}},
		{"minus three", -3, []byte{0x05}},
		{"127", 127, []byte{0xfe, 0x01}},
		{"minus 127", -127, []byte{0xfd, 0x01}},
		{"128", 128, []byte{0x80, 0x02}},
		{"minus 128", -128, []byte{0xff, 0x01}},
		{"129", 129, []byte{0x82, 0x02}},
		{"minus 129", -129, []byte{0x81, 0x02}},
		{"max int64", math.MaxInt64, []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
		{"min int64", math.MinInt64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ZigzagEncode(tt.value)

			if !bytes.Equal(got.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", got.Bytes(), tt.want)
			}
		})
	}
}

func TestZigzagInlineEncode(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x02}},
		{"minus one", -1, []byte{0x01}},
		{"two", 2, []byte{0x04}},
		{"minus two", -2, []byte{0x03}},
		{"127", 127, []byte{0xfe, 0x01}},
		{"minus 127", -127, []byte{0xfd, 0x01}},
		{"128", 128, []byte{0x80, 0x02}},
		{"minus 128", -128, []byte{0xff, 0x01}},
		{"max int64", math.MaxInt64, []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
		{"min int64", math.MinInt64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer

			ZigzagInlineEncode(tt.value, &buffer)

			if !bytes.Equal(buffer.Bytes(), tt.want) {
				t.Fatalf("got %x, want %x", buffer.Bytes(), tt.want)
			}
		})
	}
}

func TestZigzagInlineEncodeWithExistingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0xaa, 0xbb})

	ZigzagInlineEncode(-1, buffer)

	want := []byte{0xaa, 0xbb, 0x01}

	if !bytes.Equal(buffer.Bytes(), want) {
		t.Fatalf("got %x, want %x", buffer.Bytes(), want)
	}
}

func TestZigzagDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int64
	}{
		{"zero", []byte{0x00}, 0},
		{"one", []byte{0x02}, 1},
		{"minus one", []byte{0x01}, -1},
		{"two", []byte{0x04}, 2},
		{"minus two", []byte{0x03}, -2},
		{"127", []byte{0xfe, 0x01}, 127},
		{"minus 127", []byte{0xfd, 0x01}, -127},
		{"128", []byte{0x80, 0x02}, 128},
		{"minus 128", []byte{0xff, 0x01}, -128},
		{"129", []byte{0x82, 0x02}, 129},
		{"minus 129", []byte{0x81, 0x02}, -129},
		{"max int64", []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, math.MaxInt64},
		{"min int64", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, math.MinInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := ZigzagDecode(buffer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}

			if buffer.Len() != 0 {
				t.Fatalf("got %d remaining bytes, want 0", buffer.Len())
			}
		})
	}
}

func TestZigzagDecodeWithRemainingData(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x01, 0xaa, 0xbb})

	got, err := ZigzagDecode(buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != -1 {
		t.Fatalf("got %d, want -1", got)
	}

	if !bytes.Equal(buffer.Bytes(), []byte{0xaa, 0xbb}) {
		t.Fatalf("got %x, want aabb", buffer.Bytes())
	}
}

func TestZigzagDecodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", nil},
		{"truncated", []byte{0x80}},
		{"truncated multi byte", []byte{0x80, 0x80}},
		{
			"overflow",
			[]byte{
				0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0x02,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(tt.input)

			got, err := ZigzagDecode(buffer)
			if err == nil {
				t.Fatal("expected error")
			}

			if got != 0 {
				t.Fatalf("got %d, want 0", got)
			}
		})
	}
}

func TestZigzagRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{"zero", 0},
		{"one", 1},
		{"minus one", -1},
		{"two", 2},
		{"minus two", -2},
		{"small positive", 12345},
		{"small negative", -12345},
		{"large positive", 1 << 32},
		{"large negative", -(1 << 32)},
		{"max int64", math.MaxInt64},
		{"min int64", math.MinInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := ZigzagEncode(tt.value)

			got, err := ZigzagDecode(encoded)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if got != tt.value {
				t.Fatalf("got %d, want %d", got, tt.value)
			}
		})
	}
}

func TestZigzagInlineRoundTrip(t *testing.T) {
	tests := []int64{
		0,
		1,
		-1,
		2,
		-2,
		127,
		-127,
		128,
		-128,
		1 << 32,
		-(1 << 32),
		math.MaxInt64,
		math.MinInt64,
	}

	for _, value := range tests {
		t.Run("", func(t *testing.T) {
			var buffer bytes.Buffer

			ZigzagInlineEncode(value, &buffer)

			got, err := ZigzagDecode(&buffer)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if got != value {
				t.Fatalf("got %d, want %d", got, value)
			}

			if buffer.Len() != 0 {
				t.Fatalf("got %d remaining bytes, want 0", buffer.Len())
			}
		})
	}
}
