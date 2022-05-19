package net

import (
	"cache-go/net/pool/slicePool"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"runtime"
	"sync/atomic"
)

var (
	MaximumQuantityPreTime = 8192
)

type eventLoop struct {
	s           *Server
	poller      *poller
	connections map[int]*conn
	buffer      []byte
	eventHandle EventHandler
}

func newEventLoopWithEventHandle(p *poller, handle EventHandler) (el *eventLoop) {
	el = &eventLoop{
		eventHandle: handle,
		poller:      p,
		connections: make(map[int]*conn),
		buffer:      make([]byte, MaximumQuantityPreTime), //slicePool.get(MaxTasksPreLoop),
	}
	p.el = el
	return
}

func newEventLoop(s *Server, p *poller) (el *eventLoop) {
	el = &eventLoop{
		s:           s,
		poller:      p,
		connections: make(map[int]*conn),
		buffer:      make([]byte, MaximumQuantityPreTime), //slicePool.get(MaxTasksPreLoop),
	}
	p.el = el
	if s != nil {
		el.eventHandle = s.eventHandle
	}
	return
}

func (el *eventLoop) closeAllConn() {
	for _, val := range el.connections {
		_ = val.Close()
		_ = el.poller.Delete(val.fd)
	}
	el.connections = nil
}

func (el *eventLoop) closeConn(fd int) (err error) {
	if c, ok := el.connections[fd]; ok {
		delete(el.connections, c.fd)
		_ = el.poller.Delete(c.fd)
		err = c.close()
	}
	return
}

func (el *eventLoop) Close() (err error) {
	slicePool.Put(el.buffer)
	el.buffer = nil
	el.closeAllConn()
	err = el.poller.Close()
	return
}

func (el *eventLoop) register(c *conn) error {
	el.connections[c.FD()] = c
	if err := el.poller.AddRead(c.FD()); err != nil {
		return err
	}
	c.e = el
	c.closed = false
	return nil
}

func (el *eventLoop) asyncRegister(c *conn) error {
	return el.poller.AddUrgentTask(func(c interface{}) error {
		return el.register(c.(*conn))
	}, c)
}

var testNum int32

func (el *eventLoop) read(c *conn) error {
	n, err := unix.Read(c.fd, el.buffer)
	if err != nil {
		switch err {
		case unix.EAGAIN, unix.EINTR:
			err = nil
			fallthrough
		default:
			return err
		}
	} else if n == 0 {
		return c.e.closeConn(c.fd)
	}

	c.buffer = el.buffer[:n]
	var buf, bytes []byte
	for bytes, err = c.read(); err == nil; bytes, err = c.read() {
		buf, err = el.eventHandle.React(bytes, c)
		bytes = nil
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
	}

	if err == ErrorBytesLengthTooShort {
		err = nil
	}

	if len(c.buffer) != 0 {
		_, _ = c.inBoundBuffer.Write(c.buffer)
	}

	return nil
}

func (el *eventLoop) write(c *conn) (err error) {
	data := make([][]byte, 2)
	data[0], data[1] = c.outBoundBuffer.PeekN(MaximumQuantityPreTime)

	var n int
	for i := range data {
		n, err = unix.Write(c.fd, data[i])
		switch err {
		case nil:
		case unix.EAGAIN:
			err = nil
			fallthrough
		default:
			return
		}
		c.outBoundBuffer.ShiftN(n)
		if n < len(data[i]) {
			break
		}
	}

	if c.outBoundBuffer.IsEmpty() && c.writeState != 0 {
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
				if err = el.read(conn); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		zap.L().Fatal("subReactors error:", zap.Error(err))
	}
	zap.L().Info("subReactor shut down")
	return
}

func (el *eventLoop) startMainReactor() (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = el.poller.Poller(func(fd int32, event uint32) (err error) {
		return el.s.accept(int(fd))
	}); err != nil {
		zap.L().Fatal("mainReactor err:", zap.Error(err))
	}
	zap.L().Info("main Reactor shut down")
	return
}
