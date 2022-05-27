package cache_go

import (
	"cache-go/byteString"
	net2 "cache-go/net"
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewTcpCacheServer(t *testing.T) {
	count := 1000
	var m = make(map[string]string)
	for i := 0; i < 10; i++ {
		str := "lqf" + strconv.Itoa(i)
		m[str] = str
	}
	var getFunc GetterFunc = func(ctx context.Context, key string, skin byteString.Skin) error {
		if val, ok := m[key]; ok {
			skin.SetString(val)
			return nil
		}
		return KeyDoesNotExists
	}

	NewCacheHub("lqf", getFunc, 1024, SetLfuCache())
	p, err := NewTcpCacheServer("tcp", ":8081", net2.NewDefaultLengthFieldBasedFrameCodec(), net2.NewDefaultHashBalance(), net2.SetReuseAddr(1))
	if err != nil {
		t.Fatal(err)
	}
	//_ = p.AddRemoteAddr("tcp", ":5201")
	go p.Start(false, -1, SetMaxOpen(20), SetMaxLife(time.Second*10), SetMaxIdle(10))

	time.Sleep(time.Second)
	wg := &sync.WaitGroup{}
	for i := 0; i < count; i++ {
		if i%100 == 0 {
			time.Sleep(time.Second)
		}
		wg.Add(1)
		go func() {
			var s string
			err = Get(context.Background(), "lqf", "lqf7", byteString.NewStringSkin(&s))
			fmt.Println("err", err, "handlerGetter", s)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestNewTcpCacheServer2(t *testing.T) {
	dial, err := net.Dial("tcp", "192.168.1.103:8081")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = dial.Close()
}
