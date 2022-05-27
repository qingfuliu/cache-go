package net

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type fakehandle struct {
	DefaultHandler
}

func (f *fakehandle) OnEstablish(c Conn) error {
	fmt.Println("ok")
	return nil
}

func TestNewClient(t *testing.T) {
	client, err := NewTcpClient(SetClientHandle(&fakehandle{}))
	if err != nil {
		t.Fatal(err)
	}
	count := 100
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			_, _ = client.ConnectWithTimeOut(context.Background(), "tcp", "192.168.1.103:5200", time.Second*10)
			//fmt.Println(err)
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("come here")
	time.Sleep(time.Second * 3)
}
