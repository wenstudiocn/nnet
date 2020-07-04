package nnet

// Default implementation of IPacket and IProtocol

type Packet struct {
	msg []byte
	ifClose int32 //
	reason int32
}

func NewPacket(data []byte, params ...int32) *Packet {
	var ifClose int32 = 0
	var reason int32 = 0
	if len(params) > 0 {
		ifClose = params[0]
	}
	if len(params) > 1 {
		reason = params[1]
	}
	return &Packet{
		msg: data,
		ifClose: ifClose,
		reason: reason,
	}
}

func (self *Packet) Msg() []byte {
	return self.msg
}

func (self *Packet) Serialize() []byte {
	return self.msg
}

func (self *Packet) Destroy(b []byte) {

}

func (self *Packet) ShouldClose() (bool, int32) {
	return self.ifClose != 0, self.reason
}

type Protocol struct {
	BufSize int
}

func NewProtocol(buf_size int) *Protocol {
	return &Protocol{BufSize: buf_size}
}

func (self *Protocol) ReadPacket(conn IConn) (IPacket, error) {
	buf := make([]byte, self.BufSize)

	n, err := conn.Read(buf)
	if nil != err {
		return nil, err
	}
	return NewPacket(buf[:n]), nil
}