package net

import (
	"golang.org/x/sys/unix"
	"net"
)

type listener struct {
	fd        int
	localAddr net.Addr
	sa        unix.Sockaddr
	closed    bool
	poller    *poller
}

func newListener(proto, addr string, opts ...SocketOpt) (l *listener, err error) {
	l = new(listener)

	if l.fd, l.sa, l.localAddr, err = TcpSocket(proto, addr, false, true, opts...); err != nil {
		return nil, err
	}

	l.closed = true
	return
}

func (ls *listener) open(poller *poller) (err error) {
	if !ls.closed {
		return
	}
	ls.poller = poller
	ls.closed = false
	if err = ls.poller.AddRead(ls.fd); err != nil {
		return
	}

	if err = unix.Listen(ls.fd, MaxSocketListenNums); err != nil {
		return
	}

	return
}

func (ls *listener) close() (err error) {
	ls.closed = true
	_ = ls.poller.Delete(ls.fd)
	_ = ls.poller.AddUrgentTask(func(i interface{}) error {
		return ErrorServerShutDown
	}, nil)

	err = unix.Close(ls.fd)

	return
}

func (ls *listener) isClosed() bool {
	return ls.closed
}
