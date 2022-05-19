package net

import (
	"fmt"
	"golang.org/x/sys/unix"
	"testing"
	"time"
)

func TestTcpSocket(t *testing.T) {
	fd, sa, remoteAddr, err := TcpSocket("tcp", "127.0.0.1:5200", true, true, SetSocketNoDelay(1), SetRecvBuffer(1200), SetSendBuffer(1200))
	if err != nil {
		return
	}
	defer unix.Close(fd)
	fmt.Println(sa, remoteAddr)
	time.Sleep(time.Second * 69)
}
