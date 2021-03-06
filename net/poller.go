package net

import (
	_ "cache-go/logger"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"runtime"
	"sync/atomic"
	"unsafe"
)

var (
	DefaultEpollSize = 4096
	MaxTasksPreLoop  = 1024
)

type EpollEventFunc func(fd int32, event uint32) error
type poller struct {
	fd          int
	weakUpFd    int
	weakState   int32
	tasks       *AsyncQueue
	urgentTasks *AsyncQueue
	eventBuf    []unix.EpollEvent
	weakUpBuf   []byte
	state       int32
	el          *eventLoop
}

func NewPoller() (p *poller, err error) {
	p = new(poller)
	if p.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
		return nil, err
	}

	if p.weakUpFd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC); err != nil {
		_ = p.Close()
		return nil, err
	}

	if err = p.AddRead(p.weakUpFd); err != nil {
		_ = p.Close()
		return nil, err
	}

	p.tasks = NewAsyncQueue()
	p.urgentTasks = NewAsyncQueue()
	p.eventBuf = make([]unix.EpollEvent, DefaultEpollSize)
	p.weakUpBuf = make([]byte, 8)
	return
}

func (p *poller) Close() (err error) {
	if err = unix.Close(p.fd); err != nil {
		return
	}
	if err = unix.Close(p.weakUpFd); err != nil {
		return
	}
	return
}

func (p *poller) Poller(fn EpollEventFunc) (err error) {
	var n int
	var doTasks bool
	for {
		doTasks = false
		n, err = unix.EpollWait(p.fd, p.eventBuf, -1)
		if n < 0 || err != nil && err == unix.EINTR {
			runtime.Gosched()
			continue
		} else if err != nil {
			zap.L().Fatal("Poller err", zap.Error(err))
			return err
		}

		for i := 0; i < n; i++ {
			if int(p.eventBuf[i].Fd) != p.weakUpFd {
				err = fn(p.eventBuf[i].Fd, p.eventBuf[i].Events)
				switch err {
				case nil:
				default:
					_ = p.el.closeConn(int(p.eventBuf[i].Fd))
					zap.L().Error("poller err when excaute epollEventFunc", zap.Error(err))
				}
			} else {
				if _, err := unix.Read(p.weakUpFd, p.weakUpBuf); err != nil {
					zap.L().Fatal("err when do urgent extra task", zap.Error(err))
				}
				atomic.StoreInt32(&p.weakState, 0)
				doTasks = true
			}
		}

		//----------------extra tasks------------------------------//

		if doTasks {
			//----------------urgent tasks------------------------------//
			task := p.urgentTasks.Pop()
			for ; task != nil; task = p.urgentTasks.Pop() {
				err = task.Run()
				switch err {
				case nil:
				case ErrorServerShutDown:
					return p.el.Close()
				default:
					zap.L().Error("err when do urgent extra task", zap.Error(err))
				}
			}
			//----------------common tasks------------------------------//
			task = p.tasks.Pop()
			for i := 0; task != nil && i < MaxTasksPreLoop; i++ {
				err = task.Run()
				switch err {
				case nil:
				default:
					zap.L().Error("err when do common  extra task", zap.Error(err))
				}
				task = p.tasks.Pop()
			}

			if (!p.urgentTasks.IsEmpty() || !p.tasks.IsEmpty()) && atomic.CompareAndSwapInt32(&p.weakState, 0, 1) {
				_, err = unix.Write(p.weakUpFd, weakBytes)
				switch err {
				case nil:
				case unix.EAGAIN, unix.EINTR:
					atomic.CompareAndSwapInt32(&p.weakState, 1, 0)
				default:
					zap.L().Fatal("weak up err after do extra tasks!", zap.Error(err))
				}
			}
		}

		if n == len(p.eventBuf) {
			p.eventBuf = make([]unix.EpollEvent, n<<1)
		} else if n <= len(p.eventBuf)>>1 {
			p.eventBuf = make([]unix.EpollEvent, len(p.eventBuf)>>1)
		}

	}
}

func (p *poller) AddTask(fn func(interface{}) error, arg interface{}) error {
	p.tasks.Push(fn, arg)
	return p.Ticker()
}

func (p *poller) AddUrgentTask(fn func(interface{}) error, arg interface{}) error {
	p.urgentTasks.Push(fn, arg)
	return p.Ticker()
}

var weakVar int64 = 1
var weakBytes = (*((*[8]byte)(unsafe.Pointer(&weakVar))))[:]

func (p *poller) Ticker() (err error) {
	if atomic.CompareAndSwapInt32(&p.weakState, 0, 1) {
	retry:
		_, err = unix.Write(p.weakUpFd, weakBytes)
		if err == unix.EAGAIN {
			goto retry
		}
		if err != nil {
			zap.L().Fatal("ticker fatal", zap.Error(err))
		}
	}
	return
}

var (
	readEvent      uint32 = unix.EPOLLPRI | unix.EPOLLIN | unix.EPOLLRDHUP
	writeEvent     uint32 = unix.EPOLLOUT
	readWriteEvent        = readEvent | writeEvent
)

func (p *poller) ModifyRead(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: readEvent,
	})
}
func (p *poller) ModifyReadWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: readWriteEvent,
	})
}
func (p *poller) AddRead(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: readEvent,
	})
}

func (p *poller) AddWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: writeEvent,
	})
}

func (p *poller) AddReadWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: readWriteEvent,
	})
}

func (p *poller) Delete(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd, nil)
}
