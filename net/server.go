package net

import (
	_ "cache-go/logger"
	"go.uber.org/zap"
	"runtime"
	"sync"
)

type HandlerContainer interface {
	getHandler() EventHandler
}

type Server struct {
	listener      *listener
	mainEventLoop *eventLoop
	lb            LoadBalance
	eventHandle   EventHandler
	wg            sync.WaitGroup
	cond          *sync.Cond
	codeC         CodeC
}

func NewDefaultServer(addr string, handler EventHandler) (*Server, error) {
	return NewServer(
		"tcp",
		addr,
		NewDefaultLengthFieldBasedFrameCodec(),
		NewDefaultHashBalance(),
		handler,
		SetReuseAddr(1))
}

func NewServer(proto, addr string, codeC CodeC, lb LoadBalance, handler EventHandler, listenerOpts ...SocketOpt) (s *Server, err error) {
	s = new(Server)
	s.codeC = codeC
	s.lb = lb
	s.eventHandle = handler

	if codeC == nil {
		s.codeC = NewNothingCodec()
	}
	if handler == nil {
		s.eventHandle = NewDefaultHandler()
	}
	if lb == nil {
		s.lb = NewDefaultHashBalance()
	}

	listener, err := newListener(proto, addr, listenerOpts...)
	if err != nil {
		zap.L().Fatal("init listener err", zap.Error(err))
		return nil, err
	}
	s.listener = listener

	s.wg = sync.WaitGroup{}
	s.cond = sync.NewCond(&sync.Mutex{})

	return
}

func (s *Server) getHandler() EventHandler {
	return s.eventHandle
}

func (s *Server) Start(lockOsThread bool, numReactors int) (err error) {
	if numReactors <= 0 {
		numReactors = runtime.NumCPU()
	}

	var pl *poller
	for i := 0; i < numReactors; i++ {
		if pl, err = NewPoller(); err != nil {
			return err
		}
		el := newEventLoop(s, pl)
		s.lb.register(el)
	}
	zap.L().Info("subReactors starting ...")
	s.startSubReactors(lockOsThread)

	if pl, err = NewPoller(); err != nil {
		s.Stop()
		return err
	}

	el := newEventLoop(s, pl)
	s.mainEventLoop = el
	if err = s.listener.open(pl); err != nil {
		zap.L().Error("listener open failed", zap.Error(err))
		return
	}

	zap.L().Info("mainReactors starting ...")
	go func() {
		s.wg.Add(1)
		_ = s.startMainReactor(el)
		s.wg.Done()
	}()
	return s.afterShutDown()
}

func (s *Server) afterShutDown() (err error) {
	s.waitForShutDown()

	s.lb.iterator(func(loop *eventLoop) bool {
		_ = loop.poller.AddUrgentTask(func(i interface{}) error {
			return ErrorServerShutDown
		}, nil)
		return true
	})

	if err = s.listener.close(); err != nil {
		zap.L().Error("close listener err", zap.Error(err))
	}

	s.wg.Wait()
	return nil
}

func (s *Server) waitForShutDown() {
	s.cond.L.Lock()
	s.cond.Wait()
	s.cond.L.Unlock()
}

func (s *Server) Stop() {
	s.cond.L.Lock()
	s.cond.Signal()
	s.cond.L.Unlock()
}

func (s *Server) startSubReactors(lockOsThread bool) {
	s.lb.iterator(func(loop *eventLoop) bool {
		s.wg.Add(1)
		go func() {
			_ = loop.startSubReactors(lockOsThread)
			s.wg.Done()
		}()
		return true
	})
}

func (s *Server) startMainReactor(el *eventLoop) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = el.poller.Poller(func(fd int32, event uint32) (err error) {
		return s.accept(int(fd))
	}); err != nil {
		zap.L().Fatal("mainReactor err:", zap.Error(err))
	}
	zap.L().Info("main Reactor shut down")
	return
}
