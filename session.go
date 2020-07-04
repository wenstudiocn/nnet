package nnet

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// 关闭原因
const (
	// Close Reason 是一个 int32 型数据，这是系统预置的几个代码
	CLOSE_REASON_READ          = 0
	CLOSE_REASON_WRITE         = 0
	CLOSE_REASON_PROTOCOL      = 1
	CLOSE_REASON_READTIMEOUT   = 4  // HEARTBEAT
	CLOSE_REASON_SERVER_CLOSED = 16 // 本服务器关闭
)

// 长连接
type Session struct {
	id        uint64
	hub       IHub
	conn      IConn
	extraData interface{}
	once      sync.Once // Close once
	closed    int32     // session 是否关闭
	chClose   chan struct{}
	chSend    chan IPacket
	chRecv    chan IPacket
}

func newSession(conn IConn, h IHub) *Session {
	return &Session{
		hub:     h,
		conn:    conn,
		chClose: make(chan struct{}),
		chSend:  make(chan IPacket, h.Conf().SizeOfSendChan),
		chRecv:  make(chan IPacket, h.Conf().SizeOfRecvChan),
	}
}

func (self *Session) GetData() interface{} {
	return self.extraData
}

func (self *Session) SetData(data interface{}) {
	self.extraData = data
}

func (self *Session) GetRawConn() IConn {
	return self.conn
}

func (self *Session) UpdateId(id uint64) {
	self.id = id
	self.hub.PutSession(id, self)
}

func (self *Session) Id() uint64 {
	return self.id
}

func (self *Session) SetId(id uint64) {
	self.id = id
}

func (self *Session) Do() {
	suc, reason := self.hub.Callback().OnConnected(self)
	if !suc {
		//TODO: 这里不 Close 资源能释放吗?
		self.Close(reason)
		return
	}

	asyncDo(self.loopHandle, self.hub.Wg())
	asyncDo(self.loopWrite, self.hub.Wg())
	asyncDo(self.loopRead, self.hub.Wg())
}

func (self *Session) Close(reason int32) {
	self.close(reason)
}

func (self *Session) close(reason int32) {
	self.once.Do(func() {
		atomic.StoreInt32(&self.closed, 1)

		close(self.chClose)
		close(self.chSend)
		close(self.chRecv)

		self.conn.Close()

		self.hub.DelSession(self.id)

		self.hub.Callback().OnClosed(self, reason)
	})
}

func (self *Session) IsClosed() bool {
	return atomic.LoadInt32(&self.closed) != 0
}

func (self *Session) Write(pkt IPacket, timeout time.Duration) error {
	if self.IsClosed() {
		return ErrConnClosing
	}
	if timeout > 0 {
		_ = self.conn.SetWriteDeadline(time.Now().Add(timeout))
	}
	_, err := self.conn.Write(pkt.Serialize())
	return err
}

// public 异步写入
func (self *Session) AWrite(pkt IPacket, timeout time.Duration) (err error) {
	if self.IsClosed() {
		return ErrConnClosing
	}

	defer func() {
		if e := recover(); e != nil {
			err = ErrConnClosing
		}
	}()

	if timeout == 0 {
		select {
		case self.chSend <- pkt:
			return nil
		default:
			return ErrWriteBlocking
		}
	} else {
		select {
		case self.chSend <- pkt:
			return nil
		case <-self.chClose:
			return ErrConnClosing
		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}
}

// 循环从 socket 读取数据，置入 chRecv 通道
func (self *Session) loopRead() {
	var reason int32 = 0

	defer func() {
		self.close(reason)
	}()

	for {
		select {
		case <-self.hub.ChQuit():
			reason = CLOSE_REASON_SERVER_CLOSED
			return
		case <-self.chClose:
			return
		default:
		}
		if self.hub.Conf().ReadTimeout > 0 {
			fmt.Println("timeout:", self.hub.Conf().ReadTimeout)
			self.conn.SetReadDeadline(time.Now().Add(self.hub.Conf().ReadTimeout))
		}
		pkt, err := self.hub.Protocol().ReadPacket(self.conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				reason = CLOSE_REASON_READTIMEOUT
			} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				reason = CLOSE_REASON_READTIMEOUT
			} else {
				reason = CLOSE_REASON_READ
			}
			//
			return
		}
		self.chRecv <- pkt
	}
}

// 循环从 cbSend 通道读取数据，发送到 socket
func (self *Session) loopWrite() {
	var reason int32 = 0

	defer func() {
		//fmt.Println(self.id)
		self.close(reason)
	}()

	ticker := time.NewTicker(self.hub.Conf().Tick)
	for {
		select {
		case <-self.hub.ChQuit():
			reason = CLOSE_REASON_SERVER_CLOSED
			return
		case <-self.chClose:
			return
		case <-ticker.C:
			self.hub.Callback().OnHeartbeat(self)
		case pkt := <-self.chSend:
			if self.IsClosed() {
				return
			}
			data := pkt.Serialize()
			defer pkt.Destroy(data)

			_ = self.conn.SetWriteDeadline(time.Now().Add(self.hub.Conf().Timeout))
			_, err := self.conn.Write(data)
			if err != nil {
				reason = CLOSE_REASON_WRITE
				return
			}
			yes, rsn := pkt.ShouldClose()
			if yes {
				reason = rsn
				return
			}
		}
	}
}

// 循环从 chRecv 读取解析好的数据，回调到实现层
func (self *Session) loopHandle() {
	var reason int32 = 0

	defer func() {
		self.close(reason)
	}()

	for {
		select {
		case <-self.hub.ChQuit():
			reason = CLOSE_REASON_SERVER_CLOSED
			return
		case <-self.chClose:
			return
		case pkt := <-self.chRecv:
			if self.IsClosed() {
				return
			}
			if !self.hub.Callback().OnMessage(self, pkt) {
				reason = CLOSE_REASON_PROTOCOL
				return
			}
		}
	}
}

func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}
