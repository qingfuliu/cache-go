package net

import (
	_ "cache-go/logger"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoller(t *testing.T) {
	var wg sync.WaitGroup
	closePoller := func(interface{}) error {
		wg.Done()
		return ErrorServerShutDown
	}

	poller, err := NewPoller()
	if err != nil {
		zap.L().Fatal("poller create fatal", zap.Error(err))
	}
	go poller.Poller(nil)
	wg.Add(1)

	testFunc := func(arg interface{}) error {
		i := arg.(*int64)
		atomic.AddInt64(i, 1)
		zap.L().Debug("value of i", zap.Int64("i", atomic.LoadInt64(i)))
		return nil
	}
	var k int64
	nums := 3000
	for i := 0; i < nums; i++ {
		err = poller.AddTask(testFunc, &k)
		if err != nil {
			zap.L().Fatal("addTaskError", zap.Error(err))
		}
	}

	//for i := 0; i < 1000; i++ {
	//	err = poller.AddUrgentTask(testFunc, &k)
	//	if err != nil {
	//		zap.L().Fatal("addTaskError", zap.Error(err))
	//	}
	//}

	time.Sleep(time.Second * 3)
	poller.AddUrgentTask(closePoller, nil)
	wg.Wait()

	if k != int64(nums) {
		zap.L().Fatal("k should be nums,but is", zap.Int64("k", k), zap.Int("nums", nums))
	}
}
