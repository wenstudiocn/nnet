package nnet

import "time"

var (
	DEF_SESSION_ID     uint64 = 0
	RUBBISH_SESSION_ID uint64 = 1
)

type ISession interface {
	Do() // session 开始工作

	Close(int32) // 停止所有工作
	IsClosed() bool

	Write(IPacket, time.Duration) error
	AWrite(IPacket, time.Duration) error // async write

	GetData() interface{} // data field
	SetData(interface{})

	UpdateId(uint64) // update session id
	Id() uint64
	SetId(uint64)

	GetRawConn() IConn
}

type ISessionCallback interface {
	OnClosed(ISession, int32)
	OnConnected(ISession) (bool, int32)
	OnMessage(ISession, IPacket) bool
	OnHeartbeat(ISession) bool
}
