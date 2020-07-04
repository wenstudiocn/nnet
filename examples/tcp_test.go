package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
	"github.com/wenstudiocn/nnet"
)
var s *nnet.WsServer
var c *nnet.WsClient

var tcp_server_conf = &nnet.HubConfig{
	SizeOfSendChan: 1024,
	SizeOfRecvChan: 1024,
	ReadBufSize: 1024,
	WriteBufSize: 1024,
	Timeout: 5 * time.Second, // timeout of sending, receiving etc.
	Tick:           30 * time.Second, // interval of timed callback
	ReadTimeout:    12 * time.Second,
}

var tcp_client_conf = &nnet.HubConfig{
	SizeOfSendChan: 1024,
	SizeOfRecvChan: 1024,
	ReadBufSize: 1024,
	WriteBufSize: 1024,
	Timeout: 5 * time.Second,
	Tick:           30 * time.Second,
	ReadTimeout:    12 * time.Second,
}

///////////////////////////////////////////////////////
type TcpServerCb struct {}

func (self *TcpServerCb)OnClosed(ses nnet.ISession, reason int32) {
	fmt.Println("client closed ", ses.Id(), reason)
	ws.DelSession(ses.Id())
}

func (self *TcpServerCb)OnConnected(ses nnet.ISession) (bool, int32) {
	fmt.Println("a client Connected ", ses.Id())
	ses.UpdateId(1)
	return true, 0
}

func (self *TcpServerCb)OnMessage(ses nnet.ISession, pkt nnet.IPacket) bool {
	fmt.Println("received a package")
	return true
}

func (self *TcpServerCb)OnHeartbeat(nnet.ISession) bool {
	fmt.Println("Heartbeat ticker")
	return true
}

///////////////////////////

type TcpServerProtocol struct {}

func (self *TcpServerProtocol)ReadPacket(conn nnet.IConn) (nnet.IPacket, error) {
	//TODO: mem pool
	buf := make([]byte, 1024)

	n, err := io.ReadFull(conn, buf[:2])
	if err != nil {
		fmt.Println("readfull:", err)
		return nil, err
	}
	if n < 2 {
		fmt.Println(nnet.ErrBufferSizeInsufficient)
		return nil, nnet.ErrBufferSizeInsufficient
	}

	data_len := binary.LittleEndian.Uint16(buf[:2])
	if data_len > 0 {
		n, err = io.ReadFull(conn, buf[2:data_len])
		if nil != err {
			fmt.Println(err)
			return nil, err
		}
		if n != int(data_len) {
			fmt.Println(nnet.ErrBufferSizeInsufficient)
			return nil, nnet.ErrBufferSizeInsufficient
		}
	}

	return NewWsServerPacket(data_len, buf), nil
}

///////////////////////////
type TcpServerPacket struct {
	len uint16
	buf []byte
}

func NewTcpServerPacket(length uint16, buf []byte) nnet.IPacket {
	b := make([]byte, 1024)
	copy(b[2:len(buf)], buf)
	return &WsServerPacket{
		len: length,
		buf: b,
	}
}

// serialize to binary format to be sent
func (self *TcpServerPacket)Serialize() []byte {
	binary.LittleEndian.PutUint16(self.buf, self.len)
	return self.buf
}

// free the memory if needs
func (self *TcpServerPacket)Destroy([]byte) {

}
// if need to close socket after sending this packet
func (self *TcpServerPacket)ShouldClose() (bool, int32) {
	return false, 0
}


///////////////////////////////// Client side /////////////

type TcpClientCb struct {}

func (self *TcpClientCb)OnClosed(ses nnet.ISession, reason int32) {
	fmt.Println("closed:", ses.Id(), reason)
	wsc.DelSession(ses.Id())
}

func (self *TcpClientCb)OnConnected(ses nnet.ISession) (bool, int32) {
	fmt.Println("Connected ", ses.Id())
	ses.UpdateId(1)
	fmt.Println("sessions:", wsc.GetSessionNum())
	return true, 0
}

func (self *TcpClientCb)OnMessage(ses nnet.ISession, pkt nnet.IPacket) bool {
	return true
}

func (self *TcpClientCb)OnHeartbeat(nnet.ISession) bool {
	fmt.Println("Heartbeat ticker")
	return true
}

func TestTcp(t *testing.T) {
	ln, err := net.ListenTCP("tcp", ":3001")
	// server
	s = nnet.NewTcpServer(tcp_server_conf, new(TcpServerCb), new(TcpServerProtocol), ln)
	go s.Start()
	time.Sleep(1 * time.Second)

	// client
	c = nnet.NewWsClient(tcp_client_conf, new(TcpClientCb), new(TcpServerProtocol), "127.0.0.1:3000")
	c.Start()

	time.Sleep(1 * time.Second)

	fmt.Println("all sessions:", c.GetSessionNum())
	ses, err := c.GetSession(1)
	if err != nil {
		t.Error("error2:", err)
	}

	sentence := []byte("Hello")
	err = ses.AWrite(NewTcpServerPacket(uint16(len(sentence)), sentence), 0)
	if nil != err {
		t.Error(err)
	}
	t.Log("sent")
	time.Sleep(15 * time.Second)

	c.Stop()
	s.Stop()
}