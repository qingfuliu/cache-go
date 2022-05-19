package main

import (
	_ "cache-go/logger"
	"cache-go/net"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

type echo struct {
	net.DefaultHandler
}

func (e *echo) React(b []byte, c net.Conn) ([]byte, error) {

	return nil, c.Close()
}

func main() {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	go func() {
		_ = http.ListenAndServe(":8081", nil)
	}()
	s, err := net.NewServer("tcp", ":5201", net.NewDefaultLengthFieldBasedFrameCodec(), nil, &echo{}, net.SetReuseAddr(1))
	if err != nil {
		zap.L().Fatal("new server err", zap.Error(err))
	}
	//time.AfterFunc(time.Second*15, func() {
	//	s.Stop()
	//})
	err = s.Start(false, 0)
	if err != nil {
		zap.L().Error("err after server", zap.Error(err))
	}
}
