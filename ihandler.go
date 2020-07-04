package nnet

type IHandler interface {
	Dispatch(ISession, IPacket) bool
}
