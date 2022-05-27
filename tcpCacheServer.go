package cache_go

import (
	"cache-go/byteString"
	_ "cache-go/logger"
	"cache-go/msg"
	net2 "cache-go/net"
	"context"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net"
	"sync"
	"time"
)

type TcpCacheServer struct {
	net2.DefaultHandler
	s            *net2.Server
	peerPicker   *tcpPeerPicker
	codeC        net2.CodeC
	localAddr    net.Addr
	reactTimeOut time.Duration
	wg           *sync.WaitGroup
}

func (c *TcpCacheServer) React(message []byte, conn net2.Conn) (data []byte, err error) {
	in := &msg.GetRequest{}
	out := &msg.GetResponse{}
	if err = proto.Unmarshal(message, in); err != nil {
		out.Error = err.Error()
		return
	}
	var str string
	skin := byteString.NewStringSkin(&str)
	ctx, cancelFunc := context.WithTimeout(context.Background(), c.reactTimeOut)
	err = Get(ctx, in.CacheName, in.Key, skin)
	defer cancelFunc()
	if err != nil {
		out.Error = err.Error()
	}
	out.Val = str
	data, err = proto.Marshal(out)
	return
}

func NewTcpCacheServer(proto, addr string, codeC net2.CodeC, lb net2.LoadBalance, listenerOpts ...net2.SocketOpt) (tCS *TcpCacheServer, err error) {
	tCS = new(TcpCacheServer)
	tCS.localAddr, err = net.ResolveTCPAddr("proto", addr)

	if err != nil {
		zap.L().Fatal("err", zap.Error(err))
	}

	tCS.s, err = net2.NewServer(proto, addr, codeC, lb, tCS, listenerOpts...)
	if err != nil {
		return nil, err
	}
	tCS.codeC = codeC

	tCS.peerPicker = NewTcpPeerPicker(tCS)
	return
}

func (c *TcpCacheServer) Start(lockOsThread bool, numReactor int, option ...TcpPoolOption) error {
	zap.L().Info("----------------------------------cache Server Start--------------------------------------")
	zap.L().Info("info:", zap.String("localAddr", c.localAddr.String()), zap.Duration("react timeout", c.reactTimeOut))
	if err := c.peerPicker.start(option...); err != nil {
		return err
	}
	c.wg.Add(1)
	err := c.s.Start(lockOsThread, numReactor)
	c.wg.Done()
	return err
}

func (c *TcpCacheServer) AddRemoteAddr(proto, addr string) error {
	return c.peerPicker.addAddr(proto, addr)
}

func (c *TcpCacheServer) Close() {
	c.peerPicker.close()
	c.s.Stop()
	c.wg.Wait()
}
