package net

import (
	"cache-go/net/pool/bufferPool"
	"cache-go/net/pool/slicePool"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"runtime"
	"sync/atomic"
)

var (
	MaximumQuantityPreTime = 2048
)

type eventLoop struct {
	s           *Server
	poller      *poller
	connections map[int]*conn
	buffer      []byte
	eventHandle EventHandler
}

func newEventLoop(s *Server, p *poller) (el *eventLoop) {
	el = &eventLoop{
		s:           s,
		poller:      p,
		connections: make(map[int]*conn),
		buffer:      slicePool.Get(MaxTasksPreLoop),
		eventHandle: s.eventHandle,
	}
	p.el = el
	return
}

func (el *eventLoop) closeAllConn() {
	for key, val := range el.connections {
		_ = val.Close()
		delete(el.connections, key)
	}
}

func (el *eventLoop) closeConn(c *conn) error {
	if c, ok := el.connections[c.fd]; ok {
		delete(el.connections, c.fd)
		return c.Close()
	}
	return nil
}

func (el *eventLoop) Close() {
	slicePool.Put(el.buffer)
	el.buffer = nil
	el.closeAllConn()
}

func (el *eventLoop) register(c *conn) error {
	el.connections[c.FD()] = c
	if err := el.poller.AddRead(c.FD()); err != nil {
		return err
	}
	c.e = el
	return nil
}

func (el *eventLoop) asyncRegister(c *conn) error {
	return el.poller.AddUrgentTask(func(c interface{}) error {
		return el.register(c.(*conn))
	}, c)
}

var testNum int32

func (el *eventLoop) read(c *conn) error {

	n, err := unix.Read(c.FD(), el.buffer)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return err
	} else if n == 0 {
		return el.closeConn(c)
	}
	c.buffer = el.buffer[:n]
	var buf, bytes []byte
	for bytes, err = c.read(); err == nil; bytes, err = c.read() {
		buf, err = el.eventHandle.React(bytes, c)
		if err != nil {
			return nil
		}
		if buf != nil {
			_, err = c.write(buf)
			if err != nil {
				return err
			}
		}
		slicePool.Put(bytes)
		//bufferPool.PutBuffer(&bytebufferpool.ByteBuffer{B: bytes})
		if c.cache != nil {
			bufferPool.PutBuffer(c.cache)
		}
	}
	if err != nil {
		if err == ErrorBytesLengthTooShort {
			err = nil
		} else {
			return err
		}
	}
	//slicePool.Put(bytes)

	c.cache = nil
	if len(c.buffer) != 0 {
		_, _ = c.inBoundBuffer.Write(c.buffer)
	}
	return nil
}

func (el *eventLoop) write(c *conn) (err error) {
	head, tail := c.outBoundBuffer.PeekN(MaximumQuantityPreTime)
	var n int
	n, err = unix.Write(c.FD(), head)

	if err != nil {
		if err == unix.EAGAIN {
			err = nil
		}
		return
	}

	if n == len(head) {
		var n2 int
		n2, err = unix.Write(c.FD(), tail)
		if err != nil {
			if err == unix.EAGAIN {
				err = nil
			}
			return
		}
		n += n2
	}
	c.outBoundBuffer.ShiftN(n)
	if c.OutLen() == 0 && c.writeState != 0 {
		err = el.poller.ModifyRead(c.FD())
		c.writeState = 0
	}
	return
}

func (el *eventLoop) startSubReactors(lockOsThread bool) (err error) {
	if lockOsThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}
	if err = el.poller.Poller(func(fd int32, event uint32) (err error) {
		if conn, ok := el.connections[int(fd)]; ok {

			if event&writeEvent != 0 && !conn.outBoundBuffer.IsEmpty() {
				if err = el.write(conn); err != nil {
					return err
				}
			}

			if event&readEvent != 0 {
				atomic.AddInt32(&testNum, 1)
				zap.L().Debug("testNums", zap.Int32("testNum", testNum))
				if err = el.read(conn); err != nil {
					return err
				}
				zap.L().Debug("after testNums", zap.Int32("testNum", testNum))
			}
		}
		return nil
	}); err != nil {
		zap.L().Error("startSubReactors error:", zap.Error(err))
	}
	return
}

func (el *eventLoop) startMainReactor() (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = el.poller.Poller(func(fd int32, event uint32) (err error) {
		return el.s.accept(int(fd))
	}); err != nil {
		zap.L().Error("startMainReactor err:", zap.Error(err))
	}
	return
}
