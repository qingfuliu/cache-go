package lfu

import (
	"cache-go/byteString"
	"cache-go/cache"
)

type CacheLfu struct {
	m                map[string]*treeNode
	tree             *RBTree
	nBytes           int64
	nHit, nGet       cache.AtomicInt64
	nEliminate, nAdd int64
}

func NewLfuCache() cache.GoCache {
	return &CacheLfu{}
}

func (lfu *CacheLfu) Add(key string, value byteString.ByteString) {
	if lfu.tree == nil {
		lfu.tree = NewRBTree()
		lfu.m = make(map[string]*treeNode)
	}
	if node, ok := lfu.m[key]; ok {
		node.val = &value
		lfu.tree.LruAdjust(node)
		return
	}
	node := lfu.tree.LruInsert(key, &value)
	lfu.m[key] = node
	lfu.nBytes += node.Size()
	lfu.nAdd++
}
func (lfu *CacheLfu) Get(key string) (byteString.ByteString, bool) {
	lfu.nGet.Add(1)
	if lfu.tree == nil {
		return byteString.ByteString{}, false
	}
	if node, ok := lfu.m[key]; ok {
		lfu.nHit.Add(1)
		return *node.val, true
	}
	return byteString.ByteString{}, false
}
func (lfu *CacheLfu) Eliminate(nBytes int64) {
	for nBytes > 0 && len(lfu.m) > 0 {
		node := lfu.tree.Eliminate()
		nBytes -= node.Size()
		delete(lfu.m, node.key)
		lfu.nEliminate++
	}
}
func (lfu *CacheLfu) NBytes() int64 {
	return lfu.nBytes
}
