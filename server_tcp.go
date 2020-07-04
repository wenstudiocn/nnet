package nnet

import (
	"net"
	"time"
)

type TcpServer struct {
	*Hub
	listener *net.TCPListener
}

func NewTcpServer(cf *HubConfig, cb ISessionCallback, p IProtocol, ls *net.TCPListener) *TcpServer {
	return &TcpServer{
		Hub:      newHub(cf, cb, p),
		listener: ls,
	}
}

func (self *TcpServer) Start() error {
	self.wg.Add(1)
	defer func() {
		self.listener.Close()
		self.wg.Done()
	}()

	for {
		select {
		case <-self.chQuit:
			return nil
		default:
		}
		self.listener.SetDeadline(time.Now().Add(self.conf.Timeout))

		conn, err := self.listener.AcceptTCP()
		if err != nil {
			continue
		}
		self.wg.Add(1)
		go func() {
			ses := newSession(conn, self)
			ses.Do()
			self.wg.Done()
		}()
	}
	return nil
}

func (self *TcpServer) DoJob(int) {

}

func (self *TcpServer) Stop() error {
	close(self.chQuit)
	self.wg.Wait() //保证 Start 完全退出
	return nil
}
