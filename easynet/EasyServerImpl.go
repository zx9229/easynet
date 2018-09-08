package easynet

import (
	"net"
)

//EasyServer3 omit
type EasyServer3 struct {
	onConnected    EasyConnected    //连接成功的回调
	onDisconnected EasyDisconnected //连接断线的回调
	onMessage      EasyMessage      //收到消息的回调
	listener       net.Listener
	cache          *safeSet
}

//NewEasyServer3 omit
func NewEasyServer3() *EasyServer3 {
	curData := new(EasyServer3)
	curData.cache = newSafeSet()
	return curData
}

//RegEasyConnected omit
func (thls *EasyServer3) RegEasyConnected(handler EasyConnected) bool {
	thls.onConnected = handler
	return true
}

//RegEasyDisConnected omit
func (thls *EasyServer3) RegEasyDisConnected(handler EasyDisconnected) bool {
	thls.onDisconnected = handler
	return true
}

//RegEasyMessage omit
func (thls *EasyServer3) RegEasyMessage(handler EasyMessage) bool {
	thls.onMessage = handler
	return true
}

//Run omit
func (thls *EasyServer3) Run(tcpAddr string) error {
	var err error
	if thls.listener, err = net.Listen("tcp", tcpAddr); err != nil {
		return err
	}
	var conn net.Conn
	for {
		if conn, err = thls.listener.Accept(); err != nil {
			return err
		}
		eSock := newEasySocket3(conn)
		eSock.RegEasyConnected(thls.onConnected)
		eSock.RegEasyDisConnected(thls.onDisconnected)
		eSock.RegEasyMessage(thls.onMessage)
		eSock.setIsAccepted(true)
		thls.cache.Add(eSock)
		go eSock.doRecv(conn, thls.actionWhenDis)
	}
}

func (thls *EasyServer3) actionWhenDis(eSock *EasySocket3) {
	thls.cache.Del(eSock)
}
