package nnet

import (
	"math/rand"
	"sync"
	"time"
)

type Hub struct {
	sync.Mutex
	conf   *HubConfig
	cbSes  ISessionCallback
	prot   IProtocol
	chQuit chan struct{}
	wg     *sync.WaitGroup

	sess map[uint64]ISession
}

func newHub(cf *HubConfig, cb ISessionCallback, p IProtocol) *Hub {
	return &Hub{
		conf:   cf,
		cbSes:  cb,
		prot:   p,
		chQuit: make(chan struct{}),
		wg:     &sync.WaitGroup{},
		sess:   make(map[uint64]ISession),
	}
}

func (self *Hub) Wg() *sync.WaitGroup {
	return self.wg
}

func (self *Hub) ChQuit() <-chan struct{} {
	return self.chQuit
}

func (self *Hub) Conf() *HubConfig {
	return self.conf
}

func (self *Hub) Callback() ISessionCallback {
	return self.cbSes
}

func (self *Hub) Protocol() IProtocol {
	return self.prot
}

func (self *Hub) PutSession(id uint64, ses ISession) error {
	self.Lock()
	//@Notice: replace
	self.sess[id] = ses
	self.Unlock()
	return nil
}

func (self *Hub) DelSession(id uint64) error {
	self.Lock()
	defer self.Unlock()

	if _, ok := self.sess[id]; !ok {
		return ErrNotExists
	}
	delete(self.sess, id)
	return nil
}

func (self *Hub) PeekSession(id uint64) (ISession, error) {
	self.Lock()
	defer self.Unlock()

	s, ok := self.sess[id]
	if !ok {
		return nil, ErrNotExists
	}

	delete(self.sess, id)

	return s, nil
}

func (self *Hub) GetSession(id uint64) (ISession, error) {
	self.Lock()
	defer self.Unlock()

	s, ok := self.sess[id]
	if !ok {
		return nil, ErrNotExists
	}
	return s, nil
}

func (self *Hub) GetAllSessions() map[uint64]ISession {
	self.Lock()
	defer self.Unlock()

	return self.sess
}

var (
	s = rand.NewSource(time.Now().UnixNano())
	r = rand.New(s)
)

func Intn(a int) int {
	return r.Intn(a)
}

func (self *Hub) RandSession() (ISession, error) {
	self.Lock()
	defer self.Unlock()

	sz := len(self.sess)

	sel := Intn(sz)
	counter := 0
	for _, ses := range self.sess {
		if counter == sel {
			return ses, nil
		}
		counter += 1
	}
	return nil, ErrNotExists
}

func (self *Hub) GetSessionNum() int {
	self.Lock()
	defer self.Unlock()
	return len(self.sess)
}

