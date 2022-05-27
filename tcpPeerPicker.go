package cache_go

import (
	"cache-go/consistentHash"
	"cache-go/msg"
	net2 "cache-go/net"
	"context"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

var nowFunc = time.Now

type tcpPeerPicker struct {
	cacheServer *TcpCacheServer

	client          *net2.TcpClient
	cMap            *consistentHash.ConsistentMap
	peerGetterPools map[string]*tcpPeerGetterPool
	remoteAddr      []string
	wg              *sync.WaitGroup
	failedRetry     int
	dialTimeOut     time.Duration
}

type TcpPeerPickerOption func(peerPicker *tcpPeerPicker)

func SetMaxRetry(failedRetry int) TcpPeerPickerOption {
	return func(peerPicker *tcpPeerPicker) {
		peerPicker.failedRetry = failedRetry
	}
}

func SetDialTimeOut(dialTimeOut time.Duration) TcpPeerPickerOption {
	return func(peerPicker *tcpPeerPicker) {
		peerPicker.dialTimeOut = dialTimeOut
	}
}

type tcpPeerPickerEventHandler struct {
	net2.DefaultHandler
}

func (e *tcpPeerPickerEventHandler) React(data []byte, c net2.Conn) ([]byte, error) {
	getResponse := &msg.GetResponse{}
	err := proto.Unmarshal(data, getResponse)
	attachment := c.Context().(*TcpGetterAttachment)
	select {
	case <-attachment.msgChan:
	default:
		attachment.msgChan <- &message{
			err: err,
			msg: getResponse,
		}
	}
	return nil, nil
}

func NewTcpPeerPicker(cacheServer *TcpCacheServer) *tcpPeerPicker {
	tcpPicker := &tcpPeerPicker{
		cacheServer:     cacheServer,
		cMap:            consistentHash.NewConsistMap(),
		peerGetterPools: make(map[string]*tcpPeerGetterPool),
		remoteAddr:      make([]string, 0),
		wg:              &sync.WaitGroup{},
		failedRetry:     10,
	}
	tcpPicker.remoteAddr = append(tcpPicker.remoteAddr, cacheServer.localAddr.String())
	RegisterGetPeerPickerFunc(tcpPicker)
	return tcpPicker
}

func (tP *tcpPeerPicker) addAddr(proto, addr string) error {
	if netAddr, err := net.ResolveTCPAddr(proto, addr); err != nil {
		return err
	} else {
		tP.remoteAddr = append(tP.remoteAddr, netAddr.String())
	}
	return nil
}

func (tP *tcpPeerPicker) GetPeer(key string, ctx context.Context) (pG PeerGetter, ok bool) {
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

func (tP *tcpPeerPicker) start(options ...TcpPoolOption) (err error) {
	options = append(options)
	for i := range tP.remoteAddr {
		if _, ok := tP.peerGetterPools[tP.remoteAddr[i]]; !ok {
			tP.cMap.Add(tP.remoteAddr[i])
			tP.peerGetterPools[tP.remoteAddr[i]] = NewTcpPeerGetterPool(tP, "tcp", tP.remoteAddr[i], options...)
		}
	}

	tP.client, err = net2.NewTcpClient(net2.SetClientHandle(&tcpPeerPickerEventHandler{}),
		net2.SetClientLocalAddr(tP.cacheServer.localAddr),
		net2.SetClientCodeC(tP.cacheServer.codeC))

	zap.L().Info("----------------------------------Peer Picker Start--------------------------------------")
	zap.L().Info("info:", zap.Int("numAddr", len(tP.peerGetterPools)), zap.Strings("remote addr", tP.remoteAddr))
	tP.wg.Add(1)
	go func() {
		tP.client.Start()
		tP.wg.Done()
	}()
	return
}

func (tP *tcpPeerPicker) close() {
	tP.client.Close()
	tP.wg.Wait()
	zap.L().Info("----------------------------------Peer Picker Close--------------------------------------")
}

type TcpConnector interface {
	Connect(ctx context.Context, proto, addr string) (*tcpPeerGetter, error)
}

func (tP *tcpPeerPicker) Connect(ctx context.Context, proto, addr string) (*tcpPeerGetter, error) {
	c, err := tP.client.Connect(ctx, proto, addr)
	if err != nil {
		return nil, err
	}
	peerGetter := NewTcpPeerGetter(c, tP.dialTimeOut)
	return peerGetter, nil
}
