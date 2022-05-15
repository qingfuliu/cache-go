package net

import (
	"cache-go/net/pool/bufferPool"
	"cache-go/net/pool/slicePool"
	"github.com/valyala/bytebufferpool"
)

var (
	DefaultRingBufferSize  = 1024
	DefaultExtendThreshold = 1048576
)

type ringBuffer struct {
	read    int
	write   int
	buf     []byte
	isEmpty bool
}

func NewDefaultRingBuffer() *ringBuffer {
	return &ringBuffer{
		isEmpty: true,
		buf:     slicePool.Get(DefaultRingBufferSize),
	}
}

func NewRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		isEmpty: true,
		buf:     slicePool.Get(FloorPower2(size)),
	}
}

func (r *ringBuffer) Size() int {
	return len(r.buf)
}

func (r *ringBuffer) Len() int {
	if r.read == r.write {
		if r.isEmpty {
			return 0
		}
		return len(r.buf)
	}

	if r.read > r.write {
		return len(r.buf) + r.write - r.read
	}
	return r.write - r.read
}

func (r *ringBuffer) IsEmpty() bool {
	return r.read == r.write && r.isEmpty
}

func (r *ringBuffer) Free() int {
	return len(r.buf) - r.Len()
}

func (r *ringBuffer) IsFull() bool {
	return r.read == r.write && !r.isEmpty
}

func (r *ringBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 || r.isEmpty {
		return
	}
	n = len(p)
	if n > r.Len() {
		n = r.Len()
	}

	if r.read >= r.write {
		len1 := copy(p, r.buf[r.read:])
		if len1 == n {
			return
		}
		len2 := n - len1
		copy(p[len1:], r.buf[:len2])
		r.read = len2
		if r.read == r.write {
			r.Reset()
		}
		return
	}
	copy(p, r.buf[r.read:r.read+n])
	r.read += n
	if r.read == len(r.buf) {
		r.read = 0
	}
	if r.read == r.write {
		r.Reset()
	}
	return

}

func (r *ringBuffer) Write(p []byte) (n int, err error) {
	if len(p) <= 0 {
		return
	}
	n = len(p)
	r.extend(len(p) - r.Free())
	if r.read > r.write {
		copy(r.buf[r.write:], p)
		r.isEmpty = false
		r.write += n
		return
	}

	len1 := copy(r.buf[r.write:], p)
	r.isEmpty = false
	r.write += len1
	if r.write == len(r.buf) {
		r.write = 0
	}
	if len1 == n {
		return
	}

	len2 := n - len1
	copy(r.buf, p[len1:])
	r.write = len2
	return
}

func (r *ringBuffer) WriteV(p [][]byte) (n int, err error) {
	if len(p) <= 1 {
		return r.Write(p[0])
	}
	sum := 0
	for i := 0; i < len(p); i++ {
		sum += len(p)
	}

	r.extend(sum - r.Free())
	for i := 0; i < len(p); i++ {
		sum, _ = r.Write(p[i])
		n += sum
	}

	return
}

func (r *ringBuffer) WriteString(str string) (n int, err error) {
	return r.Write([]byte(str))
}

func (r *ringBuffer) PeekN(n int) (head []byte, tail []byte) {

	if r.IsEmpty() || n <= 0 {
		return
	}

	if n > r.Len() {
		n = r.Len()
	}

	if r.read >= r.write {
		head = r.buf[r.read:]
		if len(head) >= n {
			head = head[:n]
			return
		}
		tail = r.buf[:n-len(head)]
		return
	}

	head = r.buf[r.read : r.read+n]

	return
}

func (r *ringBuffer) PeekAll() (head []byte, tail []byte) {
	if r.IsEmpty() {
		return
	}
	if r.read >= r.write {
		head = r.buf[r.read:]
		tail = r.buf[:r.write]
		return
	}
	head = r.buf[r.read:r.write]
	return
}

func (r *ringBuffer) ShiftN(n int) int {
	if r.IsEmpty() || n <= 0 {
		return 0
	}
	if n >= r.Len() {
		n = r.Len()
		r.Reset()
		return n
	}
	if r.read > r.write {
		r.read = (r.read + n) % len(r.buf)
	} else {
		r.read += n
	}
	return n
}

func (r *ringBuffer) Reset() {
	r.read, r.write = 0, 0
	r.isEmpty = true
}

func (r *ringBuffer) ReadWithBytes(p []byte) (bytes *bytebufferpool.ByteBuffer) {
	defer r.Reset()
	bytes = bufferPool.GetBuffer()
	head, tail := r.PeekAll()

	_, _ = bytes.Write(head)
	if len(tail) != 0 {
		_, _ = bytes.Write(tail)
	}
	_, _ = bytes.Write(p)
	return bytes
}

func (r *ringBuffer) PeekAllWithBytes(p []byte) (bytes *bytebufferpool.ByteBuffer) {
	bytes = bufferPool.GetBuffer()
	head, tail := r.PeekAll()

	_, _ = bytes.Write(head)
	if len(tail) != 0 {
		_, _ = bytes.Write(tail)
	}
	if len(p) != 0 {
		_, _ = bytes.Write(p)
	}
	return bytes
}

func (r *ringBuffer) extend(size int) {
	if size <= 0 {
		return
	}

	size = r.Size() + size

	if size > DefaultExtendThreshold {
		newCap := r.Size()
		for newCap < size {
			newCap += newCap / 4
		}
		size = newCap
	} else if size < r.Size()<<1 {
		size = r.Size() << 1
	}

	newBuf := slicePool.Get(size)
	if !r.IsEmpty() {
		if r.read >= r.write {
			copy(newBuf, r.buf[r.read:])
			copy(newBuf[len(r.buf)-r.read:], r.buf[:r.write])
		} else {
			copy(newBuf, r.buf[r.read:r.write])
		}
	}
	r.write = r.Len()
	r.read = 0
	slicePool.Put(r.buf)
	r.buf = newBuf
	return
}

func (r *ringBuffer) Release() {
	r.isEmpty = true
	slicePool.Put(r.buf)
	r.buf = nil
	return
}

func FloorPower2(num int) int {
	num -= 1
	num |= num >> 1
	num |= num >> 2
	num |= num >> 4
	num |= num >> 8
	num |= num >> 16
	num |= num >> 32
	num += 1
	return num
}
