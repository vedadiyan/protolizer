package codec

type WireType uint8

const (
	WireTypeVarint WireType = 0
	WireTypeI64    WireType = 1
	WireTypeLen    WireType = 2
	WireTypeSGroup WireType = 3
	WireTypeEGroup WireType = 4
	WireTypeI32    WireType = 5
)

func GetWireType(str string) WireType {
	switch str {
	case "varint":
		{
			return WireTypeVarint
		}
	case "fixed64":
		{
			return WireTypeI64
		}
	case "bytes":
		{
			return WireTypeLen
		}
	case "start_group":
		{
			return WireTypeSGroup
		}
	case "end_group":
		{
			return WireTypeEGroup
		}
	case "fixed32":
		{
			return WireTypeI32
		}
	}
	return 0
}
