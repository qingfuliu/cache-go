package net

import (
	"cache-go"
	"cache-go/byteString"
	_ "cache-go/logger"
	"cache-go/msg"
	"context"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	net2 "net"
	"time"
)

type TcpCacheServer struct {
	DefaultHandler
	s            *Server
	peerPicker   *tcpPeerPicker
	codeC        CodeC
	localAddr    net2.Addr
	reactTimeOut time.Duration
}

func (c *TcpCacheServer) React(message []byte, conn Conn) (data []byte, err error) {
	in := &msg.GetRequest{}
	out := &msg.GetResponse{}
	if err = proto.Unmarshal(message, in); err != nil {
		out.Error = err.Error()
		return
	}
	var str string
	skin := byteString.NewStringSkin(&str)
	ctx, cancelFunc := context.WithTimeout(context.Background(), c.reactTimeOut)
	err = cache_go.Get(ctx, in.CacheName, in.Key, skin)
	defer cancelFunc()
	if err != nil {
		out.Error = err.Error()
	}
	out.Val = str
	data, err = proto.Marshal(out)
	return
}

func NewTcpCacheServer(proto, addr string, codeC CodeC, lb LoadBalance, listenerOpts ...SocketOpt) (tCS *TcpCacheServer, err error) {
	tCS = new(TcpCacheServer)
	tCS.s, err = NewServer(proto, addr, codeC, lb, tCS, listenerOpts...)
	if err != nil {
		return nil, err
	}
	tCS.codeC = codeC
	tCS.localAddr = tCS.s.listener.localAddr
	tCS.peerPicker = newTcpPeerPicker(tCS)
	return
}

func (c *TcpCacheServer) Start(lockOsThread bool, numReactor int, option ...TcpPoolOption) error {
	zap.L().Info("----------------------------------cache Server Start--------------------------------------")
	zap.L().Info("info:", zap.String("localAddr", c.localAddr.String()), zap.Duration("react timeout", c.reactTimeOut))
	if err := c.peerPicker.start(option...); err != nil {
		return err
	}
	return c.s.Start(lockOsThread, numReactor)
}

func (c *TcpCacheServer) AddRemoteAddr(proto, addr string) error {
	return c.peerPicker.addAddr(proto, addr)
}
