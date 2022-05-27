package net

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"net"
	"sync"
	"time"
)

type TcpClient struct {
	el        *eventLoop
	codeC     CodeC
	localAddr net.Addr
	handle    EventHandler
	once      *sync.Once
}

type ClientOptions func(c *TcpClient)

func SetClientCodeC(codeC CodeC) ClientOptions {
	return func(c *TcpClient) {
		c.codeC = codeC
	}
}

func SetClientLocalAddr(addr net.Addr) ClientOptions {
	return func(c *TcpClient) {
		c.localAddr = addr
	}
}

func SetClientHandle(handle EventHandler) ClientOptions {
	return func(c *TcpClient) {
		c.handle = handle
	}
}

func NewTcpClient(options ...ClientOptions) (c *TcpClient, err error) {

	c = &TcpClient{
		codeC:  NewDefaultLengthFieldBasedFrameCodec(),
		handle: NewDefaultHandler(),
		once:   &sync.Once{},
	}
	for i := range options {
		options[i](c)
	}
	var poller *poller
	poller, err = NewPoller()
	if err != nil {
		return
	}

	c.el = newEventLoop(c, poller)
	return
}

func (c *TcpClient) getHandler() EventHandler {
	return c.handle
}

func (c *TcpClient) Connect(ctx context.Context, proto, addr string) (conn Conn, err error) {

	var (
		fd         int
		sa         unix.Sockaddr
		remoteAddr net.Addr
	)
	done := make(chan struct{}, 1)
	defer close(done)

	go func() {
		fd, sa, remoteAddr, err = TcpSocket(proto, addr, false, false, SetSocketNoDelay(1))
		select {
		case <-done:
			_ = unix.Close(fd)
		default:
			done <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		if err == nil {
			C := NewConn(fd, c.codeC, sa, c.localAddr, remoteAddr)
			_ = c.el.asyncRegister(C)
			return C, err
		}
	}
	return
}

func (c *TcpClient) ConnectWithTimeOut(ctx context.Context, proto, addr string, timeOut time.Duration) (conn Conn, err error) {
	deadline := time.Now().Add(timeOut)
	var cancelFunc context.CancelFunc
	if doneTime, ok := ctx.Deadline(); !ok || deadline.Before(doneTime) {
		ctx, cancelFunc = context.WithDeadline(ctx, deadline)
		defer cancelFunc()
	}
	return c.Connect(ctx, proto, addr)
}

func (c *TcpClient) Start() {
	zap.L().Info("----------------------------------Tcp Client Start--------------------------------------")
	zap.L().Info("info:", zap.String("localAddr", c.localAddr.String()), zap.String("CodeC", c.codeC.String()))
	err := c.el.startSubReactors(false)
	zap.L().Info("----------------------------------Tcp Client Stop--------------------------------------")
	zap.L().Info("tcp Client Stop....", zap.Error(err))
}

func (c *TcpClient) Close() {
	_ = c.el.poller.AddUrgentTask(func(interface{}) error {
		return ErrorServerShutDown
	}, nil)
}
