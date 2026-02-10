package utils

import (
	"bytes"
	"sync"
)

var BytesPool *bytesPool

type bytesPool struct {
	*sync.Pool
}

func newConditionPool() *bytesPool {
	return &bytesPool{
		Pool: &sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func init() {
	BytesPool = newConditionPool()
}

func (p *bytesPool) Get() *bytes.Buffer {
	return p.Pool.Get().(*bytes.Buffer)
}

func (p *bytesPool) Free(b *bytes.Buffer) {
	b.Reset()
	p.Pool.Put(b)
}
