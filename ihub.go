package nnet

import (
	"sync"
	"time"
)

var (
// Hub state
)

type HubConfig struct {
	SizeOfSendChan uint32
	SizeOfRecvChan uint32
	ReadBufSize    int
	WriteBufSize   int
	Timeout        time.Duration // timeout of sending, receiving etc.
	Tick           time.Duration // for timed callback
	ReadTimeout    time.Duration // infinite if = 0, it means client connection lost if timeout
}

type IHub interface {
	Lock() // support locker semantics
	Unlock()

	Wg() *sync.WaitGroup        // object
	ChQuit() <-chan struct{}    // return a channel used to quit hub loop
	Conf() *HubConfig           // return config object
	Callback() ISessionCallback // return callback object
	Protocol() IProtocol        // return protocol

	Start() error // start hub
	Stop() error  // stop hub
	DoJob(int)    // do other jobs

	/// session(id based) manager function
	PutSession(uint64, ISession) error
	DelSession(uint64) error
	GetSession(uint64) (ISession, error)
	PeekSession(uint64) (ISession, error)
	RandSession() (ISession, error)
	GetSessionNum() int
	GetAllSessions() map[uint64]ISession
}
