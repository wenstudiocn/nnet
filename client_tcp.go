package nnet

import (
	"net"
)

type TCPClient struct {
	*Hub
	addr string
}

func NewTCPClient(hc *HubConfig,  cb ISessionCallback, p IProtocol, addr string) *TCPClient {
	return &TCPClient{
		Hub: newHub(hc, cb, p),
		addr: addr,
	}
}

func (self *TCPClient) Start() error {
	dialer := net.Dialer{Timeout: self.conf.Timeout}
	conn, err := dialer.Dial("tcp",  self.addr)
	if nil != err {
		return err
	}

	tconn, ok := conn.(*net.TCPConn)
	if !ok {
		return ErrInterface
	}

	self.wg.Add(1)
	go func(){
		ses := newSession(TcpConn{tconn}, self)
		ses.Do()
		self.wg.Done()
	}()
	return nil
}

func (self *TCPClient) DoJob( p int ) {

}

func (self *TCPClient) Stop() error {
	close(self.chQuit)
	self.wg.Wait()
	return nil
}