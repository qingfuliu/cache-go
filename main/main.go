package main

import (
	_ "cache-go/logger"
	"cache-go/net"
	"fmt"
	"go.uber.org/zap"
	"os"
	"runtime/pprof"
	"sync/atomic"
)

type echo struct {
	net.DefaultHandler
}

var i int32

func (e *echo) React(b []byte, c net.Conn) ([]byte, error) {
	atomic.AddInt32(&i, 1)
	fmt.Println(atomic.LoadInt32(&i))
	return b, nil
}
func main() {
	heapFile, err := os.Create("./pprof/heapFile.pro")
	if err != nil {
		zap.L().Fatal("create heap profile file err", zap.Error(err))
		return
	}
	defer heapFile.Close()
	pprof.WriteHeapProfile(heapFile)

	cpuFile, err := os.Create("./pprof/cpuFile.pro")
	if err != nil {
		zap.L().Fatal("create heap profile file err", zap.Error(err))
		return
	}
	defer cpuFile.Close()
	pprof.StartCPUProfile(cpuFile)
	defer pprof.StopCPUProfile()
	s, err := net.NewServer("tcp", ":5201", net.NewDefaultLengthFieldBasedFrameCodec(), nil, &echo{}, net.SetReuseAddr(1))
	if err != nil {
		zap.L().Fatal("new server err", zap.Error(err))
	}
	err = s.Start(false)
	if err != nil {
		zap.L().Error("err after server", zap.Error(err))
	}
}
