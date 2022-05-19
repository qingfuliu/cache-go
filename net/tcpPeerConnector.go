package net

import (
	"context"
	"net"
)

type tcpPeerConnector struct {
	pool        *tcpPeerGetterPool
	proto, addr string
	localAddr   net.Addr
	codeC       CodeC
}

func neTcpPeerConnector(proto, addr string) *tcpPeerConnector {
	return &tcpPeerConnector{
		proto: proto,
		addr:  addr,
	}
}

func (tPC *tcpPeerConnector) Connect(ctx context.Context) (*tcpPeerGetter, error) {
	fd, sa, remoteAddr, err := TcpSocket(tPC.proto, tPC.addr, false, false, SetSocketNoDelay(1))
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err != nil {
			return nil, err
		}
	}

	conn := NewConn(fd, tPC.codeC, sa, tPC.localAddr, remoteAddr)
	if err = tPC.pool.tcpPeerPicker.register(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &tcpPeerGetter{
		pool:     tPC.pool,
		conn:     conn,
		createAt: nowFunc(),
		attachment: &tcpGetterAttachment{
			msgChan: make(chan *message, 1),
			wait:    0,
		},
	}, nil
}
