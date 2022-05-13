package net

import (
	"fmt"
	"net"
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

func TestNewServer(t *testing.T) {
	s, err := NewServer("tcp", ":5201", nil, nil, &echo{}, SetReuseAddr(1))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Start(false)
	t.Log(err)
}

func TestNewServer2(t *testing.T) {
	testString := "lqfhhhasdfffffhhrwtsgy4w756546rstghfgjhn43765sdfhdrjlmlaemkrgklaenrgw90458hg"
	var wg sync.WaitGroup

	var j int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			codeC := NewDefaultLengthFieldBasedFrameCodec()
			conn := &testConn{
				bytes: make([]byte, 0),
			}
			Conn, err := net.Dial("tcp4", "127.0.0.1:5201")
			defer Conn.Close()
			if err != nil {
				return
			}
			bytes, _ := codeC.Encode([]byte(testString))
			Conn.Write([]byte(bytes))
			for {
				//Conn.SetDeadline(0)
				err = Conn.SetReadDeadline(time.Now().Add(time.Minute * 2))
				if err != nil {
					fmt.Println(err)
					break
				}
				n, err := Conn.Read(bytes)
				if err != nil {
					fmt.Println(err)
					break
				}
				if n > 0 {
					conn.bytes = append(conn.bytes, bytes[:n]...)
					fmt.Println(n)
				} else {
					break
				}
				//fmt.Println(len(conn.bytes))
				bytes, err = codeC.Decode(conn)
				if err != nil {
					fmt.Println(err)
					fmt.Println(atomic.LoadInt32(&j))
				} else {
					fmt.Println(string(bytes))
					atomic.AddInt32(&j, 1)
					fmt.Println(atomic.LoadInt32(&j))
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
