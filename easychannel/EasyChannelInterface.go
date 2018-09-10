package easychannel

//EasyConnected 连接成功的回调函数
type EasyConnected func(eChannel EasyChannel)

//EasyChannel omit
type EasyChannel interface {
	Status() ChannelStatus
	Close(closeSend bool, closeRecv bool)
	Send(data []byte) error
	Recv() (data []byte, err error)
}

//EasyChannelManager 管理器
type EasyChannelManager interface {
	CreateEasyChannel() (EasyChannel, error)
	RegEasyConnected(handler EasyConnected) bool
}
