package net

import (
	cache_go "cache-go"
	"cache-go/msg"
	"context"
	"github.com/golang/protobuf/proto"
	"sync"
	"time"
)

type tcpPeerGetter struct {
	pool       *tcpPeerGetterPool
	inUse      bool
	returnAt   time.Time
	createAt   time.Time
	conn       Conn
	attachment *tcpGetterAttachment
}

func NewTcpPeerGetter(c Conn, timeOut time.Duration) cache_go.PeerGetter {
	return &tcpPeerGetter{
		attachment: &tcpGetterAttachment{
			msgChan: make(chan *message, 1),
			wait:    1,
			mu:      &sync.Mutex{},
		},
		conn: c,
	}
}
func (tp *tcpPeerGetter) Close() error {
	tp.attachment.mu.Lock()
	close(tp.attachment.msgChan)
	tp.attachment.mu.Unlock()
	return tp.conn.Close()
}
func (tp *tcpPeerGetter) validation() error {
	now := nowFunc()
	if tp.returnAt.Add(tp.pool.maxIdleTime).Before(now) || tp.createAt.Add(tp.pool.maxIdleTime).Before(now) {
		return ErrorGetterExpired
	}
	return nil
}

func (tp *tcpPeerGetter) idleTimeValidation() error {
	now := nowFunc()
	if tp.returnAt.Add(tp.pool.maxIdleTime).Before(now) {
		return ErrorTcpGetterTimeout
	}
	return nil
}

func (tp *tcpPeerGetter) timeOutValidation() error {
	now := nowFunc()
	if tp.createAt.Add(tp.pool.maxIdleTime).Before(now) {
		return ErrorTcpGetterTimeout
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
