package cache_go

import (
	"cache-go/byteString"
	"cache-go/cache"
	"cache-go/cache/lfu"
	"cache-go/msg"
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
)

func Get(ctx context.Context, groupName, key string, skin byteString.Skin) (err error) {
	mu.RLock()
	if cacheHub, ok := gableGetter[groupName]; ok {
		err = cacheHub.Get(ctx, key, skin)
	} else {
		err = GroupDoesNotExists
	}
	mu.RUnlock()
	return
}

type Getter interface {
	Get(ctx context.Context, key string, skin byteString.Skin) error
}

type GetterFunc func(ctx context.Context, key string, skin byteString.Skin) error

func (gFn GetterFunc) Get(ctx context.Context, key string, skin byteString.Skin) error {
	return gFn(ctx, key, skin)
}

type Setter interface {
	Set(ctx context.Context, key string, skin byteString.Skin) error
}

var (
	gableGetter = make(map[string]Getter)
	mu          sync.RWMutex
)

func getGetter(key string) (g Getter, ok bool) {
	g, ok = gableGetter[key]
	return
}

func NewCacheHub(name string, getter Getter, maxBytes int64, options ...CacheHubOption) *CacheHub {
	if getter == nil {
		return nil
	}
	mu.Lock()
	if _, ok := gableGetter[name]; ok {
		mu.Unlock()
		return nil
	}
	cacheHub := &CacheHub{
		name:     name,
		maxBytes: maxBytes,
		barrier:  NewBarrier(),
		getter:   getter,
		state:    NewCacheState(),
	}
	for _, val := range options {
		val(cacheHub)
	}
	if cacheHub.localCache == nil {
		cacheHub.localCache = cache.NewLruCache()
		cacheHub.hotCache = cache.NewLruCache()
	}
	gableGetter[name] = cacheHub
	mu.Unlock()
	return cacheHub
}

type CacheHubOption func(hub *CacheHub)

func SetLruCache() CacheHubOption {
	return func(hub *CacheHub) {
		hub.localCache = cache.NewLruCache()
		hub.hotCache = cache.NewLruCache()
	}
}

func SetLfuCache() CacheHubOption {
	return func(hub *CacheHub) {
		hub.localCache = lfu.NewLfuCache()
		hub.hotCache = lfu.NewLfuCache()
	}
}

type CacheHub struct {
	name       string
	localCache cache.GoCache
	hotCache   cache.GoCache
	picker     PeerPicker
	mu         sync.RWMutex
	barrier    Barrier
	maxBytes   int64
	once       sync.Once
	getter     Getter
	state      *CacheState
}

func (c *CacheHub) getFromLocal(key string) (val byteString.ByteString, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val, ok = c.localCache.Get(key); ok {
		return
	}
	if val, ok = c.hotCache.Get(key); ok {
		return
	}
	return
}

func (c *CacheHub) load(ctx context.Context, key string) (val byteString.ByteString, err error) {
	var view interface{}
	view, err = c.barrier.Execute(key, func() (val interface{}, err error) {
		var (
			ok    bool
			peer  PeerGetter
			view_ byteString.ByteString
		)
		if val, ok = c.getFromLocal(key); ok {
			return
		}
		c.once.Do(func() {
			c.picker = getPeerPicker()
		})
		peer, ok = c.picker.GetPeer(key, ctx)
		if ok {
			if peer == nil {
				err = ErrorServerBusy
				return
			}
			view_, err = c.getFromPeer(ctx, peer, key)
		} else {
			view_, err = c.getLocally(ctx, key)
		}

		if err != nil {
			return nil, KeyDoesNotExists
		}
		defer func() {
			val = view_
		}()
		c.mu.Lock()
		defer c.mu.Unlock()
		if !ok {
			c.localCache.Add(key, view_)
		} else if rand.Intn(10) == 0 {
			c.hotCache.Add(key, view_)
		}

		totalBytes := c.hotCache.NBytes() + c.localCache.NBytes()
		if totalBytes < c.maxBytes {
			return
		}
		victim := c.hotCache
		if c.localCache.NBytes() > victim.NBytes() {
			victim = c.localCache
		}
		victim.Eliminate(totalBytes - c.maxBytes)
		return
	})
	if err == nil {
		val = view.(byteString.ByteString)
	}
	return
}

func (c *CacheHub) getLocally(ctx context.Context, key string) (byteString.ByteString, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	skin := byteString.NewByteStringSkin()
	err := c.getter.Get(ctx, key, skin)
	if err != nil {
		return byteString.ByteString{}, err
	}
	bs := skin.View()
	c.localCache.Add(key, bs)
	return bs, nil
}

func (c *CacheHub) getFromPeer(ctx context.Context, peer PeerGetter, key string) (val byteString.ByteString, err error) {
	c.state.addNumPeerGet()
	in := &msg.GetRequest{
		Key:       key,
		CacheName: c.name,
	}
	out := &msg.GetResponse{}
	err = peer.Get(ctx, in, out)
	if err != nil {
		c.state.addNumsPeerMiss()
		return byteString.ByteString{}, err
	} else if out.Error != "" {
		c.state.addNumsPeerMiss()
		return byteString.ByteString{}, errors.New(out.Error)
	}

	bs := byteString.NewByteStringSkin()
	bs.SetString(out.Val)
	c.state.addNumsPeerHit()
	return bs.View(), err
}

func (c *CacheHub) Get(ctx context.Context, key string, skin byteString.Skin) error {
	c.state.addNumsGet()
	if val, ok := c.getFromLocal(key); ok {
		c.state.addNumsHit()
		skin.SetByteString(val)
		return nil
	}
	val, err := c.load(ctx, key)
	if err != nil {
		c.state.addNumsMiss()
		return err
	}
	c.state.addNumsHit()
	skin.SetByteString(val)
	return nil
}

type CacheState struct {
	numsGet      cache.AtomicInt64 //nums of the total request
	numsPeerGet  cache.AtomicInt64 //the number of times the local cache was miss but the peer search was hit
	numsHit      cache.AtomicInt64 //nums of the total hit
	numsPeerHit  cache.AtomicInt64 //the number of times the peer search was  hit
	numsMiss     cache.AtomicInt64 //the number of times the search was missed both local and peer
	numsPeerMiss cache.AtomicInt64
}

func NewCacheState() *CacheState {
	return &CacheState{}
}

func (c *CacheHub) State() *CacheState {
	return c.state
}

func (CS *CacheState) NumsGet() int64 {
	return atomic.LoadInt64((*int64)(&CS.numsGet))
}
func (CS *CacheState) NumPeerGet() int64 {
	return atomic.LoadInt64((*int64)(&CS.numsPeerGet))
}
func (CS *CacheState) NumsHit() int64 {
	return atomic.LoadInt64((*int64)(&CS.numsHit))
}
func (CS *CacheState) NumsPeerHit() int64 {
	return atomic.LoadInt64((*int64)(&CS.numsPeerHit))
}
func (CS *CacheState) NumsMiss() int64 {
	return atomic.LoadInt64((*int64)(&CS.numsMiss))
}
func (CS *CacheState) NumsPeerMiss() int64 {
	return atomic.LoadInt64((*int64)(&CS.numsPeerMiss))
}

func (CS *CacheState) addNumsGet() int64 {
	return atomic.AddInt64((*int64)(&CS.numsGet), 1)
}
func (CS *CacheState) addNumPeerGet() int64 {
	return atomic.AddInt64((*int64)(&CS.numsPeerGet), 1)
}
func (CS *CacheState) addNumsHit() int64 {
	return atomic.AddInt64((*int64)(&CS.numsHit), 1)
}
func (CS *CacheState) addNumsPeerHit() int64 {
	return atomic.AddInt64((*int64)(&CS.numsPeerHit), 1)
}
func (CS *CacheState) addNumsMiss() int64 {
	return atomic.AddInt64((*int64)(&CS.numsMiss), 1)
}
func (CS *CacheState) addNumsPeerMiss() int64 {
	return atomic.AddInt64((*int64)(&CS.numsPeerMiss), 1)
}
