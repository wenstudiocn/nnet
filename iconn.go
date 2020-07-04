package nnet

import (
	"io"
	"net"
	"time"
)

type IConn interface {
	io.ReadWriteCloser
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}
