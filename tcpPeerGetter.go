package cache_go

import (
	"cache-go/msg"
	"cache-go/net"
	"context"
	"github.com/golang/protobuf/proto"
	"time"
)

type message struct {
	msg *msg.GetResponse
	err error
}
type TcpGetterAttachment struct {
	msgChan chan *message
}

type tcpPeerGetter struct {
	pool       *tcpPeerGetterPool
	inUse      bool
	returnAt   time.Time
	createAt   time.Time
	conn       net.Conn
	attachment *TcpGetterAttachment
}

func NewTcpPeerGetter(c net.Conn, timeOut time.Duration) *tcpPeerGetter {
	return &tcpPeerGetter{
		attachment: &TcpGetterAttachment{
			msgChan: make(chan *message, 1),
		},
		conn: c,
	}
}

func (tp *tcpPeerGetter) Close() error {
	close(tp.attachment.msgChan)
	return tp.conn.Close()
}
func (tp *tcpPeerGetter) validation() error {
	now := nowFunc()
	if tp.returnAt.Add(tp.pool.maxIdleTime).Before(now) || tp.createAt.Add(tp.pool.maxIdleTime).Before(now) {
		return net.ErrorGetterExpired
	}
	return nil
}

func (tp *tcpPeerGetter) idleTimeValidation() error {
	now := nowFunc()
	if tp.returnAt.Add(tp.pool.maxIdleTime).Before(now) {
		return net.ErrorTcpGetterTimeout
	}
	return nil
}

func (tp *tcpPeerGetter) timeOutValidation() error {
	now := nowFunc()
	if tp.createAt.Add(tp.pool.maxIdleTime).Before(now) {
		return net.ErrorTcpGetterTimeout
	}
	return nil
}

func (tp *tcpPeerGetter) sendBack() {
	if !tp.pool.putConn(tp) {
		_ = tp.Close()
	}
}

func (tp *tcpPeerGetter) Get(ctx context.Context, in *msg.GetRequest, out *msg.GetResponse) (err error) {
	var data []byte
	if data, err = proto.Marshal(in); err != nil {
		tp.sendBack()
		return
	}
	if err = tp.conn.AsyncWrite(data); err != nil {
		_ = tp.Close()
		return
	}
	select {
	case <-ctx.Done():
		_ = tp.Close()
		return ctx.Err()
	case message := <-tp.attachment.msgChan:
		out.Val = message.msg.Val
		out.Error = message.msg.Error
		tp.sendBack()
		return message.err
	}
}
