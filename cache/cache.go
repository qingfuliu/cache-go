package cache

import (
	"cache-go/byteString"
	"cache-go/cache/lru"
)

type GoCache interface {
	Add(key string, value byteString.ByteString)
	Get(key string) (byteString.ByteString, bool)
	Eliminate(nBytes int64)
	NBytes() int64
}

const DefaultMaxBytes = 4096

type lruCache struct {
	*lru.CacheLRU
	nBytes           int64
	nHit, nGet       AtomicInt64
	nEliminate, nAdd int64
}

func NewLruCache() GoCache {
	return &lruCache{}
}

func (c *lruCache) Add(key string, value byteString.ByteString) {
	if c.CacheLRU == nil {
		c.CacheLRU = lru.NewCacheLru(-1, func(key, value interface{}) {
			val := value.(byteString.ByteString)
			c.nEliminate++
			c.nBytes -= int64(val.Len())
		})
	}
	c.CacheLRU.Add(key, value)
	c.nBytes += int64(value.Len())
	c.nAdd++
}

func (c *lruCache) Get(key string) (byteString.ByteString, bool) {
	c.nGet.Add(1)
	if c.CacheLRU == nil {
		return byteString.ByteString{}, false
	}
	val, ok := c.CacheLRU.Get(key)
	if !ok {
		return byteString.ByteString{}, false
	}
	c.nHit.Add(1)
	return val.(byteString.ByteString), true
}

func (c *lruCache) Eliminate(nBytes int64) {
	for nBytes > 0 && c.CacheLRU.Len() > 0 {
		e := c.CacheLRU.RemoveOldest()
		val := e.(byteString.ByteString)
		c.nBytes -= int64(val.Len())
		nBytes -= int64(val.Len())
		c.nEliminate++
	}
}
func (c *lruCache) NBytes() int64 {
	return c.nBytes
}
