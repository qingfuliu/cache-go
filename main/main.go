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

var i int32

func (e *echo) React(b []byte, c net.Conn) ([]byte, error) {
	//atomic.AddInt32(&i, 1)
	if string(b) != "lqfhhhasdfffffhhrwtsgy4w756546rstghfgjhn43765sdfhdrjlmlaemkrgklaenrgw90458hg" {
		//zap.L().Debug("err not match")
	}
	//fmt.Println(atomic.LoadInt32(&i))
	return b, nil
}
func main() {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	go func() {
		http.ListenAndServe(":8081", nil)
	}()
	//heapFile, err := os.Create("./pprof/heapFile.pro")
	//if err != nil {
	//	zap.L().Fatal("create heap profile file err", zap.Error(err))
	//	return
	//}
	//defer heapFile.Close()
	//pprof.WriteHeapProfile(heapFile)
	//
	//cpuFile, err := os.Create("./pprof/cpuFile.pro")
	//if err != nil {
	//	zap.L().Fatal("create heap profile file err", zap.Error(err))
	//	return
	//}
	//defer cpuFile.Close()
	//pprof.StartCPUProfile(cpuFile)
	//defer pprof.StopCPUProfile()
	s, err := net.NewServer("tcp", ":5201", net.NewDefaultLengthFieldBasedFrameCodec(), nil, &echo{}, net.SetReuseAddr(1))
	if err != nil {
		zap.L().Fatal("new server err", zap.Error(err))
	}
	err = s.Start(false, 1)
	if err != nil {
		zap.L().Error("err after server", zap.Error(err))
	}
}
