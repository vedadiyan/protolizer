package pdk

import (
	"bytes"
)

func SignedNumberInlineEncoder(v int64, w WireType, buffer *bytes.Buffer) {
	switch w {
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

func UnsignedNumberDecoder(w WireType, bytes *bytes.Buffer) (uint64, error) {
	switch w {
	case WireTypeI32:
		value, err := Fixed32Decode(bytes)
		if err != nil {
			return 0, err
		}
		return uint64(uint32(value)), nil

	case WireTypeI64:
		value, err := Fixed64Decode(bytes)
		if err != nil {
			return 0, err
		}
		return uint64(value), nil

	default:
		return UvarintDecode(bytes)
	}
}

func SignedNumberDecoder(w WireType, bytes *bytes.Buffer) (int64, error) {
	switch w {
	case WireTypeI32:
		{
			value, err := Fixed32Decode(bytes)
			if err != nil {
				return 0, err
			}
			return int64(value), nil
		}
	case WireTypeI64:
		{
			return Fixed64Decode(bytes)
		}
	default:
		{
			return ZigzagDecode(bytes)
		}
	}
}

func UnsignedNumberInlineEncoder(v uint64, w WireType, buffer *bytes.Buffer) {
	switch w {
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
