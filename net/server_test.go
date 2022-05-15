package net

import (
	"bytes"
	"cache-go/net/pool/slicePool"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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
	file, err := os.Open("/home/lqf/MPackage/git-2.17.1.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	testString, err := ioutil.ReadAll(file)
	//testString = []byte("asdfijhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhuiogfaerrrrrrrrrrrrrrrrr")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	fmt.Println(len(testString))
	var j int32
	for i := 0; i < 100; i++ {
		wg.Add(1)
		if i%100 == 0 {
			time.Sleep(time.Second)
		}
		go func() {
			defer wg.Done()
			codeC := NewDefaultLengthFieldBasedFrameCodec()
			conn := &testConn{
				bytes: make([]byte, 0),
			}
			data, _ := codeC.Encode(testString)
			fmt.Println("len(data)", len(data))
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

			_, _ = Conn.Write(data)
			slicePool.Put(data)
			data = nil
			for {
				//Conn.SetDeadline(0)
				//err = Conn.SetReadDeadline(0)
				if err != nil {
					fmt.Println(err)
					break
				}
				data = slicePool.Get(len(testString) * 2)
				n, err := Conn.Read(data)
				if err != nil {
					fmt.Println(err)
					break
				}
				if n > 0 {
					conn.bytes = append(conn.bytes, data[:n]...)
					fmt.Println(n)
					slicePool.Put(data)
				} else {
					slicePool.Put(data)
					break
				}
				data, err = codeC.Decode(conn)
				if err != nil {
					fmt.Println(err)
					fmt.Println(atomic.LoadInt32(&j))
					slicePool.Put(data)
				} else {
					if !bytes.Equal(data, testString) {
						fmt.Println(len(data) == len(testString))
						fmt.Println(testString)
						fmt.Println(data)
						t.Fatal("testString should be same with the data")
					}
					//fmt.Println(string(data))
					atomic.AddInt32(&j, 1)
					fmt.Println(atomic.LoadInt32(&j))
					slicePool.Put(data)
					break
				}
			}
		}()
	}
	wg.Wait()
}
