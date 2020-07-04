package nnet

import (
	"net"
)

type TcpConn struct {
	*net.TCPConn
}
