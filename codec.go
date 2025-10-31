package protolizer

import "github.com/vedadiyan/protolizer/codecs"

var (
	_static  *codecs.Static
	_dynamic *codecs.Dynamic
)

func init() {
	_static = new(codecs.Static)
	_dynamic = codecs.NewDynamic()
}

func StaticCodec() *codecs.Static {
	return _static
}

func DynamicCodec() *codecs.Dynamic {
	return _dynamic
}
