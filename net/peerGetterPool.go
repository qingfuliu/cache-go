package net

import (
	"cache-go"
	"context"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type peerGetterContainerStack struct {
	getters []*tcpPeerGetter
}

func newPeerGetterContainerStack(size int) *peerGetterContainerStack {
	return &peerGetterContainerStack{
		getters: make([]*tcpPeerGetter, size),
	}
}

func (pg *peerGetterContainerStack) size() int {
	return len(pg.getters)
}
func (pg *peerGetterContainerStack) get() (p *tcpPeerGetter) {
	if len(pg.getters) == 0 {
		return nil
	}
	p = pg.getters[len(pg.getters)-1]
	pg.getters = pg.getters[:len(pg.getters)-1]
	return
}

func (pg *peerGetterContainerStack) Insert(p *tcpPeerGetter) bool {
	pg.getters = append(pg.getters, p)
	return true
}

func (pg *peerGetterContainerStack) iterator(fn func(*tcpPeerGetter) bool) {
	nums := len(pg.getters)
	for i := 0; i < nums; {
		if fn(pg.getters[i]) {
			pg.getters[i], pg.getters[nums-1] = pg.getters[nums-1], pg.getters[i]
			nums--
		} else {
			i++
		}
	}
	pg.getters = pg.getters[:nums]
}

func (pg *peerGetterContainerStack) binarySearch(duration time.Duration) (getters []*tcpPeerGetter) {
	now := nowFunc()
	idx := sort.Search(len(pg.getters), func(i int) bool {
		if pg.getters[i].returnAt.Add(duration).Before(now) {
			return false
		}
		return true
	})
	getters = make([]*tcpPeerGetter, idx)
	copy(getters, pg.getters)
	pg.getters = pg.getters[idx:]
	return
}

type tcpPeerGetterPool struct {
	tcpPeerPicker *tcpPeerPicker

	mu        *sync.Mutex
	container *peerGetterContainerStack
	closed    bool

	connector *tcpPeerConnector
	cleaner   chan struct{}
	opener    chan struct{}

	numOpen  int64
	numClose int64
	maxOpen  int64

	maxIdle       int
	maxIdleClosed int64

	maxLifeTimeClosed int64
	maxLifetime       time.Duration

	maxIdleTimeClosed int64
	maxIdleTime       time.Duration

	waitDuration        int64
	waiting             int64
	nextRequestSequence int64
	requests            map[int64]chan *tcpPeerGetter
}

type TcpPoolOption func(pool *tcpPeerGetterPool)

func SetMaxIdle(maxIdle int) TcpPoolOption {
	return func(pool *tcpPeerGetterPool) {
		pool.maxIdle = maxIdle
	}
}

func SetMaxIdleTime(maxIdleTime time.Duration) TcpPoolOption {
	return func(pool *tcpPeerGetterPool) {
		pool.maxIdleTime = maxIdleTime
	}
}

func SetMaxOpen(maxOpen int64) TcpPoolOption {
	return func(pool *tcpPeerGetterPool) {
		pool.maxOpen = maxOpen
	}
}

func SetMaxLife(maxLife time.Duration) TcpPoolOption {
	return func(pool *tcpPeerGetterPool) {
		pool.maxLifetime = maxLife
	}
}

func setCodeC(codec CodeC) TcpPoolOption {
	return func(pool *tcpPeerGetterPool) {
		pool.connector.codeC = codec
	}
}

func setLocalAddr(localAddr net.Addr) TcpPoolOption {
	return func(pool *tcpPeerGetterPool) {
		pool.connector.localAddr = localAddr
	}
}

func NewTcpPeerGetterPool(tcpPeerPicker *tcpPeerPicker, proto, addr string, options ...TcpPoolOption) (p *tcpPeerGetterPool) {
	p = &tcpPeerGetterPool{
		tcpPeerPicker: tcpPeerPicker,
		mu:            &sync.Mutex{},
		container: &peerGetterContainerStack{
			getters: make([]*tcpPeerGetter, 0),
		},
		connector: neTcpPeerConnector(proto, addr),
		opener:    make(chan struct{}, 1),
		requests:  make(map[int64]chan *tcpPeerGetter),
	}
	p.connector.pool = p
	for _, fn := range options {
		fn(p)
	}

	if p.maxIdle <= 0 {
		p.maxIdle = -1
	}
	if p.maxOpen <= 0 {
		p.maxOpen = 5
	}
	if p.maxLifetime == 0 {
		p.maxLifetime = time.Minute * 10
	}
	if p.maxIdle == 0 {
		p.maxIdleTime = time.Minute * 5
	}
	go p.Opener()
	return
}

func (GP *tcpPeerGetterPool) Conn(ctx context.Context) (cache_go.PeerGetter, error) {
	GP.mu.Lock()
	if GP.closed {
		GP.mu.Unlock()
		return nil, ErrorPoolClosed
	}

	select {
	default:
	case <-ctx.Done():
		GP.mu.Unlock()
		return nil, ctx.Err()
	}

	var (
		peerGetter *tcpPeerGetter
		err        error
		ok         bool
	)

	idx := GP.container.size()
	if idx > 0 {
		peerGetter = GP.container.get()
		if err = peerGetter.idleTimeValidation(); err != nil {
			GP.maxIdleTimeClosed++
			GP.mu.Unlock()
			_ = peerGetter.Close()
			return nil, ErrorBadGetters
		}
		GP.mu.Unlock()
		if peerGetter.inUse {
			panic("conn is still in use")
		}
		peerGetter.inUse = true
		return peerGetter, nil
	}

	if GP.maxOpen > 0 && GP.numOpen >= GP.maxOpen {
		requestChan := make(chan *tcpPeerGetter, 1)
		key := GP.nextRequestSequence
		GP.requests[key] = requestChan
		GP.nextRequestSequence++
		GP.mu.Unlock()
		waitStart := nowFunc()
		select {
		case <-ctx.Done():
			GP.mu.Lock()
			delete(GP.requests, key)
			GP.mu.Unlock()
			atomic.AddInt64(&GP.waitDuration, int64(time.Since(waitStart)))

			select {
			default:
			case peerGetter, ok = <-requestChan:
				if ok && peerGetter != nil {
					if ok = GP.putConnLocked(peerGetter); !ok {
						_ = peerGetter.Close()
					}
				}
			}

			return nil, ctx.Err()
		case peerGetter, ok = <-requestChan:
			atomic.AddInt64(&GP.waitDuration, int64(time.Since(waitStart)))

			if !ok || peerGetter == nil {
				return nil, ErrorPoolClosed
			}

			if err := peerGetter.timeOutValidation(); err != nil {
				_ = peerGetter.Close()
				return nil, err
			}

			if peerGetter.inUse {
				panic("conn is still in use")
			}

			peerGetter.inUse = true
			return peerGetter, nil
		}
	}

	GP.numOpen++
	GP.mu.Unlock()

	var conn *tcpPeerGetter
	if conn, err = GP.newConnection(ctx); err != nil {
		GP.mu.Lock()
		GP.numOpen--
		GP.mu.Unlock()
		GP.maybeOpenNewConnection()
		return nil, err
	}
	conn.inUse = true
	return conn, nil
}

func (GP *tcpPeerGetterPool) putConnLocked(tcpGetter *tcpPeerGetter) bool {
	if GP.closed {
		return false
	}

	if err := tcpGetter.timeOutValidation(); err != nil {
		GP.maxLifeTimeClosed++
		return false
	}
	tcpGetter.inUse = false
	if len(GP.requests) > 0 {
		var key int64
		for key = range GP.requests {
			break
		}
		requestChan := GP.requests[key]
		delete(GP.requests, key)
		requestChan <- tcpGetter
		return true
	}

	if GP.maxIdle <= GP.container.size() {
		GP.maxIdleClosed++
		return false
	}
	tcpGetter.returnAt = nowFunc()
	GP.container.Insert(tcpGetter)
	GP.maybeOpenCleaner()
	return true
}

func (GP *tcpPeerGetterPool) Close() {
	GP.mu.Lock()
	defer GP.mu.Unlock()
	GP.closed = true
	close(GP.opener)

	if GP.opener != nil {
		close(GP.opener)
	}

	for _, reqCh := range GP.requests {
		close(reqCh)
	}

	GP.container.iterator(func(getter *tcpPeerGetter) bool {
		_ = getter.Close()
		return true
	})
}

func (GP *tcpPeerGetterPool) newConnection(ctx context.Context) (tG *tcpPeerGetter, err error) {
	if GP.closed {
		return nil, ErrorPoolClosed
	}
	if tG, err = GP.connector.Connect(ctx); err != nil {
		return nil, err
	}
	return tG, err
}

func (GP *tcpPeerGetterPool) Opener() {
	select {
	case _ = <-GP.opener:
		if GP.closed {
			return
		}
		tcpGetter, err := GP.newConnection(context.Background())
		if err != nil {
			GP.mu.Lock()
			GP.numOpen--
			GP.maybeOpenNewConnection()
			GP.mu.Unlock()
		} else {
			GP.mu.Lock()
			ok := GP.putConnLocked(tcpGetter)
			if !ok {
				GP.numOpen--
				_ = tcpGetter.Close()
			}
			GP.mu.Unlock()
		}
	}
}

func (GP *tcpPeerGetterPool) maybeOpenNewConnection() {
	if GP.closed {
		return
	}
	nums := int64(len(GP.requests))
	if GP.maxOpen > 0 {
		numCanOpen := GP.maxOpen - GP.numOpen
		if nums > numCanOpen {
			nums = numCanOpen
		}
	}
	for nums > 0 {
		GP.numOpen++
		GP.opener <- struct{}{}
	}
}

func (GP *tcpPeerGetterPool) maybeOpenCleaner() {
	if GP.numOpen != 0 && GP.cleaner == nil {
		GP.cleaner = make(chan struct{}, 1)
		go GP.cleanerGoroutine()
	}
}

func (GP *tcpPeerGetterPool) cleanerGoroutine() {
	ticker := time.NewTimer(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-GP.cleaner:
		case <-ticker.C:
			GP.mu.Lock()
			if GP.closed {
				GP.cleaner = nil
				return
			}
			getters := GP.container.binarySearch(GP.maxIdleTime)
			for i := range getters {
				_ = getters[i].Close()
			}
			GP.numOpen -= int64(len(getters))
			now := nowFunc()
			GP.container.iterator(func(getter *tcpPeerGetter) bool {
				if getter.createAt.Add(GP.maxLifetime).Before(now) {
					_ = getter.Close()
					GP.numOpen--
					return true
				}
				return false
			})

			if GP.numOpen == 0 {
				close(GP.cleaner)
				GP.cleaner = nil
				GP.mu.Unlock()
				return
			}
			GP.mu.Unlock()
			ticker.Reset(time.Second)
		}
	}

}

func (GP *tcpPeerGetterPool) putConn(tcpGetter *tcpPeerGetter) bool {
	GP.mu.Lock()
	defer GP.mu.Lock()
	return GP.putConnLocked(tcpGetter)
}
