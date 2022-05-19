package net

import (
	cache_go "cache-go"
	"cache-go/byteString"
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestNewTcpCacheServer(t *testing.T) {

	var m = make(map[string]string)
	for i := 0; i < 10; i++ {
		str := "lqf" + strconv.Itoa(i)
		m[str] = str
	}
	var getFunc cache_go.GetterFunc = func(ctx context.Context, key string, skin byteString.Skin) error {
		if val, ok := m[key]; ok {
			skin.SetString(val)
			return nil
		}
		return cache_go.KeyDoesNotExists
	}

	cache_go.NewCacheHub("lqf", getFunc, 1024)
	p, err := NewTcpCacheServer("tcp", "127.0.0.1:8081", NewDefaultLengthFieldBasedFrameCodec(), newDefaultHashBalance(), SetReuseAddr(1))
	if err != nil {
		t.Fatal(err)
	}
	//_ = p.AddRemoteAddr("tcp", "192.168.1.103:5200")
	go p.Start(false, -1, SetMaxOpen(20), SetMaxLife(time.Second*10), SetMaxIdle(10))

	var s string
	err = cache_go.Get(context.Background(), "lqf", "lqf", byteString.NewStringSkin(&s))
	fmt.Println(err, s)
}

func TestNewTcpCacheServer2(t *testing.T) {

}
