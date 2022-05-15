package net

import (
	_ "cache-go/logger"
	"go.uber.org/zap"
	"runtime"
	"sync"
)

type Server struct {
	listener      *listener
	mainEventLoop *eventLoop
	lb            loadBalance
	eventHandle   EventHandler
	wg            sync.WaitGroup
	cond          *sync.Cond
	codeC         CodeC
	//trickCtx      context.Context
	//trickFunc     context.CancelFunc
	//peerGetters   map[string]cache_go.PeerGetter
	//consistentMap *consistentHash.ConsistentMap
}

func NewServer(proto, addr string, codeC CodeC, lb loadBalance, handler EventHandler, listenerOpts ...SocketOpt) (s *Server, err error) {
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
		s.lb = newDefaultHashBalance()
	}

	listener, err := newListener(proto, addr, listenerOpts...)
	if err != nil {
		zap.L().Error("init listener err", zap.Error(err))
		return nil, err
	}
	s.listener = listener

	s.wg = sync.WaitGroup{}
	s.cond = sync.NewCond(&sync.Mutex{})

	return
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
		_ = el.startMainReactor()
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
