package bufferPool

import "github.com/valyala/bytebufferpool"

func GetBuffer() *bytebufferpool.ByteBuffer {
	return bytebufferpool.Get()
}

func PutBuffer(p *bytebufferpool.ByteBuffer) {
	bytebufferpool.Put(p)
}
