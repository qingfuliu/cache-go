package net

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/valyala/bytebufferpool"
)

type Conn interface {
	Get(ctx context.Context, in proto.Message, out proto.Message) error
	Chan() <-chan proto.Message
	Write(p []byte) (n int, err error)
	Close() error
	PeekAll() []byte
	ShiftN(n int)
}

type conn struct {
	fd             int
	codeC          CodeC
	cache          *bytebufferpool.ByteBuffer
	inBoundBuffer  ringBuffer
	outBoundBuffer ringBuffer
	msgChan        chan proto.Message
}

func (c *conn) Get(ctx context.Context, in proto.Message, out proto.Message) error {

}
func (c *conn) Chan() <-chan proto.Message {
	return c.msgChan
}
func (c *conn) Write(p []byte) (n int, err error) {

}
func (c *conn) Close() error {

}
func (c *conn) PeekAll() []byte {

}
func (c *conn) ShiftN(n int) {

}
