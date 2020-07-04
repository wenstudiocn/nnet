package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"testing"
	"time"
	"github.com/wenstudiocn/nnet"
)
var ws *nnet.WsServer
var wsc *nnet.WsClient

var ws_server_conf = &nnet.HubConfig{
	SizeOfSendChan: 1024,
	SizeOfRecvChan: 1024,
	ReadBufSize: 1024,
	WriteBufSize: 1024,
	Timeout: 5 * time.Second,
	Tick:           30 * time.Second,
	ReadTimeout:    12 * time.Second,
}

var ws_client_conf = &nnet.HubConfig{
	SizeOfSendChan: 1024,
	SizeOfRecvChan: 1024,
	ReadBufSize: 1024,
	WriteBufSize: 1024,
	Timeout: 5 * time.Second, // 发送等超时
	Tick:           30 * time.Second, // 定时回调
	ReadTimeout:    12 * time.Second,
}

///////////////////////////////////////////////////////
type WsServerCb struct {}

func (self *WsServerCb)OnClosed(ses nnet.ISession, reason int32) {
	fmt.Println("client closed ", ses.Id(), reason)
	ws.DelSession(ses.Id())
}

func (self *WsServerCb)OnConnected(ses nnet.ISession) (bool, int32) {
	fmt.Println("a client Connected ", ses.Id())
	ses.UpdateId(1)
	return true, 0
}

func (self *WsServerCb)OnMessage(ses nnet.ISession, pkt nnet.IPacket) bool {
	fmt.Println("received a package")
	return true
}

func (self *WsServerCb)OnHeartbeat(nnet.ISession) bool {
	fmt.Println("Heartbeat ticker")
	return true
}

///////////////////////////

type WsServerProtocol struct {}

func (self *WsServerProtocol)ReadPacket(conn nnet.IConn) (nnet.IPacket, error) {
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
type WsServerPacket struct {
	len uint16
	buf []byte
}

func NewWsServerPacket(length uint16, buf []byte) nnet.IPacket {
	b := make([]byte, 1024)
	copy(b[2:len(buf)], buf)
	return &WsServerPacket{
		len: length,
		buf: b,
	}
}

// serialize to binary format to be sent
func (self *WsServerPacket)Serialize() []byte {
	binary.LittleEndian.PutUint16(self.buf, self.len)
	return self.buf
}

// free the memory if needs
func (self *WsServerPacket)Destroy([]byte) {

}
// if need to close socket after sending this packet
func (self *WsServerPacket)ShouldClose() (bool, int32) {
	return false, 0
}


///////////////////////////////// Client side /////////////

type WsClientCb struct {}

func (self *WsClientCb)OnClosed(ses nnet.ISession, reason int32) {
	fmt.Println("closed:", ses.Id(), reason)
	wsc.DelSession(ses.Id())
}

func (self *WsClientCb)OnConnected(ses nnet.ISession) (bool, int32) {
	fmt.Println("Connected ", ses.Id())
	ses.UpdateId(1)
	fmt.Println("sessions:", wsc.GetSessionNum())
	return true, 0
}

func (self *WsClientCb)OnMessage(ses nnet.ISession, pkt nnet.IPacket) bool {
	return true
}

func (self *WsClientCb)OnHeartbeat(nnet.ISession) bool {
	fmt.Println("Heartbeat ticker")
	return true
}

func TestWs(t *testing.T) {
	// server
	ws = nnet.NewWsServer(ws_server_conf, new(WsServerCb), new(WsServerProtocol), "127.0.0.1:3000", "", nil)
	go ws.Start()
	time.Sleep(1 * time.Second)

	// client
	wsc = nnet.NewWsClient(ws_client_conf, new(WsClientCb), new(WsServerProtocol), "ws://127.0.0.1:3000/ws")
	wsc.Start()

	time.Sleep(1 * time.Second)

	fmt.Println("all sessions:", wsc.GetSessionNum())
	ses, err := wsc.GetSession(1)
	if err != nil {
		t.Error("error2:", err)
	}

	sentence := []byte("Hello")
	err = ses.AWrite(NewWsServerPacket(uint16(len(sentence)), sentence), 0)
	if nil != err {
		t.Error(err)
	}
	t.Log("sent")
	time.Sleep(15 * time.Second)

	wsc.Stop()
	ws.Stop()
}