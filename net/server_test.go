package net

import (
	"cache-go/net/pool/slicePool"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

type echo struct {
	DefaultHandler
}

func (e *echo) React(b []byte, c Conn) ([]byte, error) {
	return b, nil
}

//func TestNewServer(t *testing.T) {
//	s, err := NewServer("tcp", ":5201", nil, nil, &echo{}, SetReuseAddr(1))
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = s.Start(false, -1)
//	t.Log(err)
//}

func TestNewServer2(t *testing.T) {
	testString := "1asasdfasdfsdafsdadfsadfsadsdf"
	var wg sync.WaitGroup

	var j int32
	for i := 0; i < 1500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codeC := NewDefaultLengthFieldBasedFrameCodec()
			conn := &testConn{
				bytes: make([]byte, 0),
			}
			Conn, err := net.Dial("tcp4", "127.0.0.1:5201")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer func() {
				if Conn != nil {
					_ = Conn.Close()
				}
			}()

			bytes, _ := codeC.Encode([]byte(testString))
			Conn.Write(bytes)
			slicePool.Put(bytes)
			bytes = nil
			for {
				//Conn.SetDeadline(0)
				//err = Conn.SetReadDeadline(0)
				if err != nil {
					fmt.Println(err)
					break
				}
				bytes = slicePool.Get(len(testString) * 2)
				n, err := Conn.Read(bytes)
				if err != nil {
					fmt.Println(err)
					break
				}
				if n > 0 {
					conn.bytes = append(conn.bytes, bytes[:n]...)
					fmt.Println(n)
					slicePool.Put(bytes)
				} else {
					slicePool.Put(bytes)
					break
				}
				//fmt.Println(len(conn.bytes))
				bytes, err = codeC.Decode(conn)
				if err != nil {
					fmt.Println(err)
					fmt.Println(atomic.LoadInt32(&j))
					slicePool.Put(bytes)
				} else {
					if string(bytes) != testString {
						t.Fatal("testString should be same with the bytes", string(bytes))
					}
					fmt.Println(string(bytes))
					atomic.AddInt32(&j, 1)
					fmt.Println(atomic.LoadInt32(&j))
					slicePool.Put(bytes)
					break
				}
			}
		}()
	}
	wg.Wait()
}
