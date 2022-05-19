package net

import (
	"bufio"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"strconv"
	"strings"
)

type SocketOpt func(fd int) error

func (s SocketOpt) Apply(fd int) error {
	return s(fd)
}

func SetReuseAddr(val int) SocketOpt {
	return func(fd int) error {
		return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, val)
	}
}

func SetSocketNoDelay(val int) SocketOpt {
	return func(fd int) error {
		return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, val)
	}
}

func SetRecvBuffer(size int) SocketOpt {
	return func(fd int) error {
		return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, size)
	}
}

func SetSendBuffer(size int) SocketOpt {
	return func(fd int) error {
		return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, size)
	}
}

func SetTcpKeepAlive(val int) SocketOpt {
	return func(fd int) error {
		return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, val)
	}
}

func GetNetAddrAndSaAddr(proto, addr string) (family int, remoteAddr net.Addr, sa unix.Sockaddr, err error) {
	var (
		tcpAddress *net.TCPAddr
		version    string
	)
	tcpAddress, err = net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return
	}
	remoteAddr = tcpAddress
	version, err = handleTcpVersion(proto, tcpAddress)
	if err != nil {
		return
	}
	switch version {
	case "tcp", "tcp4":
		family = unix.AF_INET
		sa4 := &unix.SockaddrInet4{
			Port: tcpAddress.Port,
		}
		if len(tcpAddress.IP) == 16 {
			copy(sa4.Addr[:], tcpAddress.IP[12:16])
		} else {
			copy(sa4.Addr[:], tcpAddress.IP)
		}
		sa = sa4
		return
	case "tcp6":
		family = unix.AF_INET6
		sa4 := &unix.SockaddrInet6{
			Port: tcpAddress.Port,
		}
		copy(sa4.Addr[:], tcpAddress.IP)
		sa = sa4
		return
	}
	err = ErrorUnsupportedTCPProtocol
	return
}

func handleTcpVersion(proto string, remoteAddr *net.TCPAddr) (version string, err error) {
	if remoteAddr.IP.To4() == nil {
		return "tcp", nil
	}
	if remoteAddr.IP.To16() == nil {
		return "tcp6", nil
	}
	switch proto {
	case "tcp":
		return "tcp", nil
	case "tcp4":
		return "tcp4", nil
	case "tcp6":
		return "tcp6", nil
	}
	return "", ErrorUnsupportedTCPProtocol
}

var MaxSocketListenNums = 1024

func init() {
	file, err := os.Open("/proc/sys/net/core/somaxconn")
	if err != nil {
		zap.L().Fatal("open somaxConn file err", zap.Error(err))
	}
	reader := bufio.NewReader(file)
	val, err := reader.ReadString('\n')
	if err != nil {
		zap.L().Fatal("read val err", zap.Error(err))
	}
	val = strings.Trim(val, "\n")
	if maxSocketListtenNums, err := strconv.Atoi(val); err != nil {
		zap.L().Fatal("atio fatal", zap.Error(err))
	} else {
		MaxSocketListenNums = maxSocketListtenNums
	}
	zap.L().Info("max socket listen Num is ", zap.Int("num", MaxSocketListenNums))
}
func TcpSocket(proto, addr string, immediate bool, passive bool, opts ...SocketOpt) (fd int, sa unix.Sockaddr, remoteAddr net.Addr, err error) {
	var family int
	if family, remoteAddr, sa, err = GetNetAddrAndSaAddr(proto, addr); err != nil {
		return
	}
	if fd, err = unix.Socket(family, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP); err != nil {
		zap.L().Error("tcp socket err when  create socket fd", zap.Error(err))
		return
	}

	for _, val := range opts {
		if err = val(fd); err != nil {
			zap.L().Error("tcp socket err when  set SocketOpt ", zap.Error(err))
			return
		}
	}

	if !passive {
		err = unix.Connect(fd, sa)
		if err != nil {
			zap.L().Error("connect fatal!", zap.Error(err))
		}
		return
	}

	if err = unix.Bind(fd, sa); err != nil {
		zap.L().Error("tcp socket err when  binding ", zap.Error(err))
		return
	}

	if immediate {
		if err = unix.Listen(fd, MaxSocketListenNums); err != nil {
			zap.L().Error("tcp socket err when  listening ", zap.Error(err))
			return
		}
	}

	return
}
