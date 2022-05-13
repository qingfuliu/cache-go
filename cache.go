package cache_go

import (
	"cache-go/byteString"
	"cache-go/lru"
)

type cacheState struct {
}

const DefaultMaxBytes = 4096

type cache struct {
	*lru.CacheLRU
	nBytes           int64
	nHit, nGet       AtomicInt64
	nEliminate, nAdd int64
}

func NewCache() (c *cache) {
	c = &cache{}
	return c
}

func (c *cache) add(key string, value byteString.ByteString) {
	if c.CacheLRU == nil {
		c.CacheLRU = lru.NewCacheLru(-1, func(key, value interface{}) {
			val := value.(byteString.ByteString)
			c.nEliminate++
			c.nBytes -= int64(val.Len())
		})
	}
	c.Add(key, value)
	c.nBytes += int64(value.Len())
	c.nAdd++
}

func (c *cache) get(key string) (byteString.ByteString, bool) {
	if c.CacheLRU == nil {
		return byteString.ByteString{}, false
	}
	c.nGet.add(1)
	val, ok := c.Get(key)
	if !ok {
		return byteString.ByteString{}, false
	}
	c.nHit.add(1)
	return val.(byteString.ByteString), true
}

func (c *cache) eliminate(nBytes int64) {
	for nBytes > 0 {
		e := c.CacheLRU.RemoveOldest()
		val := e.(byteString.ByteString)
		c.nBytes -= int64(val.Len())
		nBytes -= int64(val.Len())
		c.nEliminate++
	}
}
