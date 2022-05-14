package net

import (
	_ "cache-go/logger"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func (s *Server) accept(fd int) (err error) {
	var (
		cfd int
		sa  unix.Sockaddr
	)

	cfd, sa, err = unix.Accept(fd)
	if err != nil {
		zap.L().Warn("accept err", zap.Int("fd", fd), zap.Error(err))
		if err == unix.EAGAIN {
			err = nil
		}
		return
	}

	var flags int
	if flags, err = unix.FcntlInt(uintptr(cfd), unix.F_GETFD, 0); err != nil {
		zap.L().Error("f cntl Int err", zap.Error(err))
		return
	} else {
		if _, err = unix.FcntlInt(uintptr(cfd), unix.F_SETFD, flags|unix.SOCK_NONBLOCK); err != nil {
			zap.L().Error("set SOCK_NONBLOCK err", zap.Error(err))
			return
		}
	}

	remoteAddr := SockaddrToTCPOrUnixAddr(sa)
	conn := NewConn(cfd, s.codeC, sa, remoteAddr, s.listener.localAddr)
	el := s.lb.Next(remoteAddr)
	if err = el.asyncRegister(conn); err != nil {
		return err
	}
	return nil
}
