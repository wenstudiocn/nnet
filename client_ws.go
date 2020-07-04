package nnet

import (
	"github.com/gorilla/websocket"
)

// 客户端组
type WsClient struct {
	*Hub
	addr 	string
	pos   int // 指示当前连接第几个 addr
}

func NewWsClient(cf *HubConfig, cb ISessionCallback, p IProtocol, addr string) *WsClient {
	return &WsClient{
		Hub:   newHub(cf, cb, p),
		addr: addr,
		pos:   0,
	}
}

func (self *WsClient) Start() error {
	conn, _, err := websocket.DefaultDialer.Dial(self.addr, nil)
	if err != nil {
		return err
	}
	self.wg.Add(1)
	go func() {
		ses := newSession(NewWsConn(conn), self)
		ses.Do()
		self.wg.Done()
	}()

	return nil
}

func (self *WsClient) DoJob(int) {

}

//func (self *WsClient) ConnectRand(addrs []string) error {
//	src := rand.NewSource(time.Now().UnixNano())
//	rnd := rand.New(src)
//	n := rnd.Intn(len(addrs))
//	addr := addrs[n]
//	return self.Connect(1, addr)
//}

func (self *WsClient) Stop() error {
	close(self.chQuit)
	self.wg.Wait()
	return nil
}
