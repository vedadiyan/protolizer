package codecs

import (
	"bytes"

	"github.com/vedadiyan/protolizer"
)

func SignedNumberInlineEncoder(v int64, field *protolizer.Field, buffer *bytes.Buffer) {
	switch field.Tags.Protobuf.WireType {
	case protolizer.WireTypeI32:
		{
			protolizer.Fixed32InlineEncode(int32(v), buffer)
		}
	case protolizer.WireTypeI64:
		{
			protolizer.Fixed64InlineEncode(v, buffer)
		}
	default:
		{
			protolizer.ZigzagInlineEncode(v, buffer)
		}
	}
}

func UnsignedNumberInlineEncoder(v uint64, field *protolizer.Field, buffer *bytes.Buffer) {
	switch field.Tags.Protobuf.WireType {
	case protolizer.WireTypeI32:
		{
			protolizer.Fixed32InlineEncode(int32(v), buffer)
		}
	case protolizer.WireTypeI64:
		{
			protolizer.Fixed64InlineEncode(int64(v), buffer)
		}
	default:
		{
			protolizer.UvarintInlineEncode(v, buffer)
		}
	}
}
