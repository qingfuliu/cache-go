package net

import (
	_ "cache-go/logger"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"log"
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

	time.Sleep(time.Second * 3)
	poller.AddUrgentTask(closePoller, nil)
	wg.Wait()

	if k != int64(nums) {
		zap.L().Fatal("k should be nums,but is", zap.Int64("k", k), zap.Int("nums", nums))
	}

	timerFd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_CLOEXEC|unix.TFD_NONBLOCK)
	if err != nil {
		log.Fatal(err)
	}
	err = unix.TimerfdSettime(timerFd, 0, &unix.ItimerSpec{
		Value: unix.Timespec{
			Sec: 5,
		},
	}, &unix.ItimerSpec{
		Interval: unix.Timespec{
			Sec:  5,
			Nsec: 0,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
func TestPoller2(t *testing.T) {
	//unix.CLOCK_MONOTONIC  unix.CLOCK_REALTIME
	timerFd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_CLOEXEC|unix.TFD_NONBLOCK)
	if err != nil {
		log.Fatal(err)
	}
	//flags：1代表设置的是绝对时间；为0代表相对时间。
	//new val set the timer val
	err = unix.TimerfdSettime(timerFd, 0, &unix.ItimerSpec{
		Value: unix.Timespec{
			Sec: 5,
		},
		Interval: unix.Timespec{
			Sec: 5,
		},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	var poller *poller
	if poller, err = NewPoller(); err != nil {
		log.Fatal(err)
	}
	if err = poller.AddRead(timerFd); err != nil {
		log.Fatal(err)
	}

	time.AfterFunc(time.Second*15, func() {
		poller.AddUrgentTask(func(interface{}) error {
			return ErrorServerShutDown
		}, nil)
	})

	timerBUf := make([]byte, 8)
	err = poller.Poller(func(fd int32, event uint32) error {
		zap.L().Debug("timer fd coming", zap.Int32("timerfd", fd))
		_, err := unix.Read(int(fd), timerBUf)
		if err == unix.EAGAIN {
			err = nil
		}
		return err
	})
	zap.L().Error("err is", zap.Error(err))
}
