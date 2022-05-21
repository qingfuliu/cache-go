package cache

import "sync/atomic"

type AtomicInt64 int64

func (v *AtomicInt64) Load() int64 {
	return atomic.LoadInt64((*int64)(v))
}

func (v *AtomicInt64) Add(deter int64) int64 {
	return atomic.AddInt64((*int64)(v), deter)
}
