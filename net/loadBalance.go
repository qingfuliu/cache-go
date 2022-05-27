package net

import (
	"hash/crc32"
	"net"
)

type LoadBalance interface {
	Next(addr net.Addr) *eventLoop
	iterator(func(loop *eventLoop) bool) *eventLoop
	register(*eventLoop)
}

type HashFunc func(str string) uint32

var defaultHashFunc HashFunc = func(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

type (
	HashLoadBalance struct {
		loops []*eventLoop
		hash  HashFunc
	}

	IteratorLoadBalance struct {
		loops []*eventLoop
		next  int
	}

	LeastLoadBalance struct {
		loops []*eventLoop
	}
)

//==================================================hash loadBalance================================================//

func NewDefaultHashBalance() LoadBalance {
	return &HashLoadBalance{
		hash:  defaultHashFunc,
		loops: make([]*eventLoop, 0),
	}
}

func NewHashBalance(hash HashFunc) LoadBalance {
	if hash == nil {
		return nil
	}
	return &HashLoadBalance{
		hash:  hash,
		loops: make([]*eventLoop, 0),
	}
}

func (hash *HashLoadBalance) Next(addr net.Addr) *eventLoop {
	//zap.L().Debug("length of the LoadBalance", zap.Int("length", len(hash.loops)))
	index := int(hash.hash(addr.String())) % len(hash.loops)
	return hash.loops[index]
}
func (hash *HashLoadBalance) iterator(fun func(loop *eventLoop) bool) *eventLoop {
	for i := range hash.loops {
		if !fun(hash.loops[i]) {
			return hash.loops[i]
		}
	}
	return nil
}
func (hash *HashLoadBalance) register(el *eventLoop) {
	hash.loops = append(hash.loops, el)
}

//==================================================iterator loadBalance================================================//

func (iterator *IteratorLoadBalance) Next(addr net.Addr) (loop *eventLoop) {
	loop = iterator.loops[iterator.next]
	iterator.next++
	if iterator.next == len(iterator.loops) {
		iterator.next = 0
	}
	return
}
func (iterator *IteratorLoadBalance) iterator(fun func(loop *eventLoop) bool) *eventLoop {
	for i := range iterator.loops {
		if !fun(iterator.loops[i]) {
			return iterator.loops[i]
		}
	}
	return nil
}
func (iterator *IteratorLoadBalance) register(el *eventLoop) {
	iterator.loops = append(iterator.loops, el)
}

//===================================================
func (least *LeastLoadBalance) Next(addr net.Addr) (loop *eventLoop) {
	min := least.loops[0].loadCountConn()
	loop = least.loops[0]
	for i := range least.loops {
		temp := least.loops[i].connCount
		if min > temp {
			min = temp
			loop = least.loops[i]
		}
	}
	return
}
func (least *LeastLoadBalance) iterator(fun func(loop *eventLoop) bool) *eventLoop {
	for i := range least.loops {
		if !fun(least.loops[i]) {
			return least.loops[i]
		}
	}
	return nil
}

func (least *LeastLoadBalance) register(el *eventLoop) {
	least.loops = append(least.loops, el)
}
