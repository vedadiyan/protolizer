package protolizer

import (
	"bytes"
	"sync"
)

var (
	_pool sync.Pool
)

func init() {
	_pool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer([]byte{})
		},
	}
}

func Alloc(size int) *bytes.Buffer {
	buffer := _pool.Get().(*bytes.Buffer)
	if size != 0 && buffer.Cap() < size {
		buffer.Grow(size)
	}
	return buffer
}

func Dealloc(buffer *bytes.Buffer) {
	buffer.Reset()
	_pool.Put(buffer)
}
