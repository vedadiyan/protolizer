package codecs

var (
	_static  *Static
	_dynamic *Dynamic
)

func init() {
	_static = new(Static)
	_dynamic = NewDynamic()
}

func StaticCodec() *Static {
	return _static
}

func DynamicCodec() *Dynamic {
	return _dynamic
}
