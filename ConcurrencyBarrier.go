package cache_go

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	err error
	val interface{}
}

type Barrier interface {
	Execute(key string, fn func() (interface{}, error)) (interface{}, error)
}

func NewBarrier() Barrier {
	return &concurrencyBarrier{
		m: make(map[string]*call),
	}
}

type concurrencyBarrier struct {
	m  map[string]*call
	mu sync.Mutex
}

func (cb *concurrencyBarrier) Execute(key string, fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	if cl, ok := cb.m[key]; ok {
		cb.mu.Unlock()
		cl.wg.Wait()
		return cl.val, cl.err
	}
	cl := &call{}
	cl.wg.Add(1)
	cb.m[key] = cl
	cb.mu.Unlock()
	cl.val, cl.err = fn()

	cl.wg.Done()
	cb.mu.Lock()
	delete(cb.m, key)
	cb.mu.Unlock()
	return cl.val, cl.err
}
