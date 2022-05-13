package slicePool

import (
	"math"
	"math/bits"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

type Pool struct {
	pools [32]sync.Pool
}

var builtinPool Pool

func Get(size int) []byte {
	return builtinPool.Get(size)
}

func Put(bytes ...[]byte) {
	for idx := range bytes {
		builtinPool.Put(bytes[idx])
	}
}

func (p *Pool) Get(size int) (buf []byte) {
	if size <= 0 {
		return nil
	}
	if size >= math.MaxInt32 {
		return make([]byte, size)
	}
	idx := index(size)
	ptr, _ := p.pools[idx].Get().(unsafe.Pointer)

	if ptr == nil {
		return make([]byte, 1<<idx)[:size]
	}

	bytes := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	bytes.Cap = 1 << idx
	bytes.Len = size
	bytes.Data = uintptr(ptr)
	runtime.KeepAlive(ptr)
	return
}

func (p *Pool) Put(bytes []byte) {
	length := cap(bytes)
	if length == 0 || length > math.MaxInt32 {
		return
	}

	idx := index(length)
	if length != 1<<idx {
		idx--
	}

	p.pools[idx].Put(unsafe.Pointer(&bytes[:1][0]))
}

func index(size int) int {
	return bits.Len(uint(size - 1))
}
