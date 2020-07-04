package nnet

type IPacket interface {
	// serialize to binary format to be sent
	Serialize() []byte
	// free the memory if needs
	Destroy([]byte)
	// if need to close socket after sending this packet
	ShouldClose() (bool, int32)
}

type IProtocol interface {
	// parse by raw data to a packet
	ReadPacket(conn IConn) (IPacket, error)
}

type IMiddleware interface {
	BeforeSend([]byte) []byte
	AfterReceived([]byte) []byte
}