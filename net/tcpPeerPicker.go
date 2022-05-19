package net

import (
	"cache-go"
	"cache-go/consistentHash"
	"cache-go/msg"
	"context"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

var nowFunc = time.Now

type tcpPeerPicker struct {
	cacheServer     *TcpCacheServer
	cMap            *consistentHash.ConsistentMap
	peerGetterPools map[string]*tcpPeerGetterPool
	remoteAddr      []string
	wg              *sync.WaitGroup
	el              *eventLoop
	poller          *poller

	failedRetry int
}

type tcpPeerPickerEventHandler struct {
	DefaultHandler
}

func (e *tcpPeerPickerEventHandler) React(data []byte, c Conn) ([]byte, error) {
	getResponse := &msg.GetResponse{}
	err := proto.Unmarshal(data, getResponse)
	attachment := c.Context().(*tcpGetterAttachment)
	attachment.mu.Lock()
	select {
	case <-attachment.msgChan:
		attachment.mu.Unlock()
	default:
		attachment.msgChan <- &message{
			err: err,
			msg: getResponse,
		}
		attachment.mu.Unlock()
	}
	return nil, nil
}

func newTcpPeerPicker(cacheServer *TcpCacheServer) *tcpPeerPicker {
	tcpPicker := &tcpPeerPicker{
		cacheServer:     cacheServer,
		cMap:            consistentHash.NewConsistMap(),
		peerGetterPools: make(map[string]*tcpPeerGetterPool),
		remoteAddr:      make([]string, 0),
		wg:              &sync.WaitGroup{},
		failedRetry:     10,
	}
	//tcpPicker.remoteAddr = append(tcpPicker.remoteAddr, cacheServer.localAddr.String())
	cache_go.RegisterGetPeerPickerFunc(tcpPicker)
	return tcpPicker
}

func (tP *tcpPeerPicker) start(options ...TcpPoolOption) (err error) {
	options = append(options, setCodeC(tP.cacheServer.codeC), setLocalAddr(tP.cacheServer.localAddr))
	for i := range tP.remoteAddr {
		if _, ok := tP.peerGetterPools[tP.remoteAddr[i]]; !ok {
			tP.cMap.Add(tP.remoteAddr[i])
			tP.peerGetterPools[tP.remoteAddr[i]] = NewTcpPeerGetterPool(tP, "tcp", tP.remoteAddr[i], options...)
		}
	}

	tP.poller, err = NewPoller()
	if err != nil {
		zap.L().Fatal("start tcpPeerPicker error", zap.Error(err))
	}

	tP.el = newEventLoopWithEventHandle(tP.poller, &tcpPeerPickerEventHandler{})
	zap.L().Info("----------------------------------Peer  Picker Start--------------------------------------")
	zap.L().Info("info:", zap.Int("addr size", len(tP.peerGetterPools)), zap.Strings("remote addr", tP.remoteAddr))
	go func() {
		tP.wg.Add(1)
		if err := tP.el.startSubReactors(false); err != nil {
			zap.L().Error("startSubReactors at tcp peerPicker error", zap.Error(err))
		}
		tP.wg.Done()
	}()
	return
}

func (tP *tcpPeerPicker) addAddr(proto, addr string) error {
	if netAddr, err := net.ResolveTCPAddr(proto, addr); err != nil {
		return err
	} else {
		tP.remoteAddr = append(tP.remoteAddr, netAddr.String())
	}
	return nil
}

func (tP *tcpPeerPicker) register(c *conn) error {
	return tP.el.register(c)
}

func (tP *tcpPeerPicker) GetPeer(key string, ctx context.Context) (pG cache_go.PeerGetter, ok bool) {
	var addr string
	if addr, ok = tP.cMap.Get(key); !ok {
		return nil, false
	}
	if addr == tP.cacheServer.localAddr.String() {
		return nil, false
	}
	var err error
	resume := tP.failedRetry
retry:
	if pG, err = tP.peerGetterPools[addr].Conn(ctx); err != nil {
		resume--
		if resume > 0 {
			goto retry
		}
		return nil, true
	}
	return
}

func (tP *tcpPeerPicker) close() {
	_ = tP.poller.AddUrgentTask(
		func(interface{}) error {
			return ErrorServerShutDown
		},
		nil,
	)
	tP.wg.Wait()
}
