package easynet2

import "net"

//EasySession omit
type EasySession interface {
	ID() int64
	IsAccepted() bool
	Close(closeSend bool, closeRecv bool)
	Send(data []byte) error
	Recv() (data []byte, err error)
	Socket() EasySocket
}

//EasyConnected 连接成功的回调函数
type EasyConnected func(eSock EasySocket, isAccepted bool, eSess EasySession, sessAccepted bool)

//EasyDisconnected 连接断开的回调函数
type EasyDisconnected func(eSock EasySocket, eSess EasySession, err error, byDisconnected bool)

//EasyMessage 收到消息的回调函数
type EasyMessage func(eSock EasySocket, eSess EasySession, data []byte)

//EasySocket omit
type EasySocket interface {
	RegEasyConnected(handler EasyConnected) bool
	RegEasyDisConnected(handler EasyDisconnected) bool
	RegEasyMessage(handler EasyMessage) bool
	SetData(v interface{})
	GetData() interface{}
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	IsOnline() bool
	Close()
	Send(data []byte) error
	CreateSession() EasySession
}

//EasyClient omit
type EasyClient interface {
	EasySocket
	Connect(tcpAddr string, doReconnect bool) error
}

//EasyServer omit
type EasyServer interface {
	RegEasyConnected(handler EasyConnected) bool
	RegEasyDisConnected(handler EasyDisconnected) bool
	RegEasyMessage(handler EasyMessage) bool
	Run(tcpAddr string) error
}
