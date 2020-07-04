package nnet

import (
	//	"fmt"

	"github.com/gorilla/websocket"
)

type WsConn struct {
	*websocket.Conn
}

func NewWsConn(c *websocket.Conn) *WsConn {
	return &WsConn{c}
}

func (self *WsConn) Read(p []byte) (int, error) {
	//_, bytes, err := self.ReadMessage()
	//if nil != err {
	//	return len(bytes), err
	//}
	//if len(bytes) > len(p) {
	//	return len(bytes), ErrBufferSizeInsufficient
	//}
	//copy(p, bytes)
	//return len(bytes), nil
	_, r, err := self.NextReader()
	if nil != err {
		return 0, err
	}
	return r.Read(p)
}

func (self *WsConn) Write(p []byte) (int, error) {
	//err := self.WriteMessage(websocket.BinaryMessage, p)
	//if nil != err {
	//	return 0, err
	//}
	//return len(p), nil
	w, err := self.NextWriter(websocket.BinaryMessage)
	if nil != err {
		return 0, err
	}
	n, err := w.Write(p)
	_ = w.Close()
	return n, err
}
