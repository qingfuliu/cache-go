package net

import (
	"cache-go/net/pool/slicePool"
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/valyala/bytebufferpool"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"net"
	"time"
)

type Conn interface {
	Get(ctx context.Context, in proto.Message, out proto.Message) error
	Chan() <-chan proto.Message
	Close() error
	PeekAll() []byte
	ShiftN(n int) int
	ResetBuf()
	Len() int
	AsyncWrite(p []byte) error
}

type conn struct {
	fd    int
	codeC CodeC
	//peekAll cache
	cache *bytebufferpool.ByteBuffer
	//event unix.read
	buffer []byte
	//buffer write to inBoundBuffer
	inBoundBuffer         *ringBuffer
	outBoundBuffer        *ringBuffer
	msgChan               chan proto.Message
	localAddr, remoteAddr net.Addr
	unixAddr              unix.Sockaddr
	getExpiration         time.Duration
	closed                bool
	e                     *eventLoop
	writeState            int32
}

func NewConn(fd int, codeC CodeC, unixAddr unix.Sockaddr, localAddr, remoteAddr net.Addr) *conn {
	return &conn{
		fd:             fd,
		codeC:          codeC,
		buffer:         slicePool.Get(1024),
		localAddr:      localAddr,
		remoteAddr:     remoteAddr,
		unixAddr:       unixAddr,
		inBoundBuffer:  NewDefaultRingBuffer(),
		outBoundBuffer: NewDefaultRingBuffer(),
		closed:         true,
	}
}

func (c *conn) FD() int {
	return c.fd
}

func (c *conn) LocalAddr() string {

	return c.localAddr.String()
}

func (c *conn) RemoteAddr() string {
	return c.remoteAddr.String()
}

func (c *conn) Closed() bool {
	return c.closed
}

func (c *conn) Get(ctx context.Context, in proto.Message, out proto.Message) error {

	return nil
}

func (c *conn) Chan() <-chan proto.Message {
	return c.msgChan
}

func (c *conn) write(p []byte) (n int, err error) {
	var bytes []byte
	if bytes, err = c.codeC.Encode(p); err != nil {
		return
	}

	if !c.outBoundBuffer.IsEmpty() {
		return c.outBoundBuffer.Write(bytes)
	}

	defer func() {
		if !c.outBoundBuffer.IsEmpty() {
			c.writeState = 1
			c.e.poller.ModifyReadWrite(c.fd)
		}
	}()

	n, err = unix.Write(c.fd, bytes)

	if err != nil && err == unix.EAGAIN {
		return c.outBoundBuffer.Write(p)
	} else if err != nil {
		return
	}

	if n < len(bytes) {
		_, _ = c.outBoundBuffer.Write(bytes[n:])
	}

	slicePool.Put(bytes)
	return
}

func (c *conn) writeV(p [][]byte) (n int, err error) {

	for i := range p {
		if packet, err := c.codeC.Encode(p[i]); err != nil {
			return 0, err
		} else {
			p[i] = packet
		}
	}

	if !c.outBoundBuffer.IsEmpty() {
		return c.outBoundBuffer.WriteV(p)
	}
	var i int
	var sum int
	for i = 0; i < len(p); i++ {
		sum, err = unix.Write(c.fd, p[i])
		if err != nil && err != unix.EAGAIN {
			return
		} else if err != nil || sum < len(p[i]) {
			break
		}
		slicePool.Put(p[i])
		n += sum
	}

	if sum < len(p[i]) {
		_, _ = c.outBoundBuffer.Write(p[i][sum:])
		n += len(p[i]) - sum
	}
	sum, _ = c.outBoundBuffer.WriteV(p[i+1:])
	n += sum
	slicePool.Put(p[i:]...)
	return
}

func (c *conn) Close() error {
	//zap.L().Debug("conn close!", zap.Int("fd", c.fd))
	c.closed = true
	if c.cache != nil {
		bytebufferpool.Put(c.cache)
		c.cache = nil
	}
	slicePool.Put(c.buffer)
	c.buffer = nil
	c.inBoundBuffer.Release()
	c.outBoundBuffer.Release()
	err := c.e.poller.Delete(c.fd)
	if err != nil {
		zap.L().Fatal("close conn err", zap.Error(err))
	}
	//close(c.msgChan)
	return unix.Close(c.fd)
}

func (c *conn) PeekAll() []byte {
	if c.inBoundBuffer.IsEmpty() {
		return c.buffer
	}
	c.cache = c.inBoundBuffer.ReadWithBytes(c.buffer)
	return c.cache.B
}

func (c *conn) ShiftN(n int) (p int) {
	if n > c.Len() {
		p = c.Len()
		c.ResetBuf()
		return
	}
	p = c.outBoundBuffer.ShiftN(n)
	if p == n {
		return
	}
	c.buffer = c.buffer[n-p:]
	p = n
	return
}

func (c *conn) ResetBuf() {
	bytebufferpool.Put(c.cache)
	c.cache = nil
	c.buffer = nil
	c.inBoundBuffer.Reset()
}
func (c *conn) Len() (n int) {
	n = c.inBoundBuffer.Len()
	if c.buffer != nil {
		n += len(c.buffer)
	}
	return
}

func (c *conn) OutLen() int {
	return c.outBoundBuffer.Len()
}

func (c *conn) read() ([]byte, error) {

	bytes, err := c.codeC.Decode(c)

	return bytes, err
}

func (c *conn) asyncWrite(p interface{}) (err error) {
	bytes := p.([]byte)
	_, err = c.write(bytes)
	return
}

func (c *conn) AsyncWrite(p []byte) error {
	return c.e.poller.AddTask(c.asyncWrite, p)
}

func (c *conn) asyncWriteV(p interface{}) (err error) {
	bytes := p.([][]byte)
	_, err = c.writeV(bytes)
	return
}

func (c *conn) AsyncWriteV(p [][]byte) error {
	return c.e.poller.AddTask(c.asyncWriteV, p)
}
