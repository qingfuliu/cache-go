package net

import (
	"hash/crc32"
	"net"
)

type loadBalance interface {
	Next(addr net.Addr) *eventLoop
	iterator(func(loop *eventLoop) bool) *eventLoop
	register(*eventLoop)
}

type HashFunc func(str string) uint32

var defaultHashFunc HashFunc = func(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

type HashLoadBalance struct {
	loops []*eventLoop
	hash  HashFunc
}

func newDefaultHashBalance() loadBalance {
	return &HashLoadBalance{
		hash:  defaultHashFunc,
		loops: make([]*eventLoop, 0),
	}
}

func newHashBalance(hash HashFunc) loadBalance {
	if hash == nil {
		return nil
	}
	return &HashLoadBalance{
		hash:  hash,
		loops: make([]*eventLoop, 0),
	}
}

func (hash *HashLoadBalance) Next(addr net.Addr) *eventLoop {
	//zap.L().Debug("length of the loadBalance", zap.Int("length", len(hash.loops)))
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
