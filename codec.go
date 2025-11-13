package protolizer

import "github.com/vedadiyan/protolizer/codecs"

var (
	_static   *codecs.Static
	_dynamic  *codecs.Dynamic
	_typeless *codecs.Typeless
)

func init() {
	_static = new(codecs.Static)
	_dynamic = codecs.NewDynamic()
	_typeless = codecs.NewTypeless()
}

func StaticCodec() *codecs.Static {
	return _static
}

func DynamicCodec() *codecs.Dynamic {
	return _dynamic
}

func TypelessCodec() *codecs.Typeless {
	return _typeless
}
