package protolizer

import "bytes"

func SignedNumberInlineEncoder(v int64, field *Field, buffer *bytes.Buffer) {
	switch field.Tags.Protobuf.WireType {
	case WireTypeI32:
		{
			Fixed32InlineEncode(int32(v), buffer)
		}
	case WireTypeI64:
		{
			Fixed64InlineEncode(v, buffer)
		}
	default:
		{
			ZigzagInlineEncode(v, buffer)
		}
	}
}

func UnsignedNumberInlineEncoder(v uint64, field *Field, buffer *bytes.Buffer) {
	switch field.Tags.Protobuf.WireType {
	case WireTypeI32:
		{
			Fixed32InlineEncode(int32(v), buffer)
		}
	case WireTypeI64:
		{
			Fixed64InlineEncode(int64(v), buffer)
		}
	default:
		{
			UvarintInlineEncode(v, buffer)
		}
	}
}
