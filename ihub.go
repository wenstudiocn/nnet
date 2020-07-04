package nnet

import (
	"sync"
	"time"
)

var (
// Hub 状态
)

type HubConfig struct {
	SizeOfSendChan uint32
	SizeOfRecvChan uint32
	ReadBufSize    int
	WriteBufSize   int
	Timeout        time.Duration // 发送等超时
	Tick           time.Duration // 定时回调
	ReadTimeout    time.Duration // 讀超時，如果為0，則無限等待。超時到達，意味著客戶端心跳丟失
}

type IHub interface {
	Lock() // support locker semantics
	Unlock()

	Wg() *sync.WaitGroup        // object
	ChQuit() <-chan struct{}    // 返回一个通道，用于退出 hub 循环
	Conf() *HubConfig           // 返回配置信息
	Callback() ISessionCallback // 返回回调对象
	Protocol() IProtocol        // 返回数据协议

	Start() error // 启动 hub
	Stop() error  // 停止 hub
	DoJob(int)    // 执行 hub 中其他任务

	PutSession(uint64, ISession) error // session 管理，这里的 session 必须基于　id
	DelSession(uint64) error
	GetSession(uint64) (ISession, error)
	PeekSession(uint64) (ISession, error)
	RandSession() (ISession, error)
	GetSessionNum() int
	GetAllSessions() map[uint64]ISession
}
