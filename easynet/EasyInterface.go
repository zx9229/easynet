package easynet

import "net"

//EasyConnected 连接成功的回调函数
type EasyConnected func(eSock EasySocket, isAccepted bool)

//EasyDisconnected 连接断开的回调函数
type EasyDisconnected func(eSock EasySocket, err error)

//EasyMessage 收到消息的回调函数
type EasyMessage func(eSock EasySocket, data []byte)

//EasySocket omit
type EasySocket interface {
	RegEasyConnected(handler EasyConnected) bool
	RegEasyDisConnected(handler EasyDisconnected) bool
	RegEasyMessage(handler EasyMessage) bool
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	IsOnline() bool
	Close()
	Send(data []byte) error
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
