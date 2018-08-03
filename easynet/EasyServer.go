package easynet

import (
	"net"
)

//EasyServer omit
type EasyServer struct {
	onConnected    CbConnected    //连接成功的回调
	onDisconnected CbDisconnected //连接断线的回调
	onMessage      CbMessage      //收到消息的回调
	listener       net.Listener
	cache          *safeSet
}

//NewEasyServer omit
func NewEasyServer() *EasyServer {
	curData := new(EasyServer)
	curData.cache = newSafeSet()
	return curData
}

//RegCbConnected omit
func (thls *EasyServer) RegCbConnected(handler CbConnected) bool {
	thls.onConnected = handler
	return true
}

//RegCbDisConnected omit
func (thls *EasyServer) RegCbDisConnected(handler CbDisconnected) bool {
	thls.onDisconnected = handler
	return true
}

//RegCbMessage omit
func (thls *EasyServer) RegCbMessage(handler CbMessage) bool {
	thls.onMessage = handler
	return true
}

//Run omit
func (thls *EasyServer) Run(tcpAddr string) error {
	var err error
	if thls.listener, err = net.Listen("tcp", tcpAddr); err != nil {
		return err
	}
	var conn net.Conn
	for {
		if conn, err = thls.listener.Accept(); err != nil {
			return err
		}
		eSock := newEasySocket(conn)
		eSock.RegCbConnected(thls.onConnected)
		eSock.RegCbDisConnected(thls.onDisconnected)
		eSock.RegCbMessage(thls.onMessage)
		thls.cache.Add(eSock)
		eSock.doRecv(conn, true, thls.actionWhenDis)
	}
}

func (thls *EasyServer) actionWhenDis(eSock *EasySocket, err error) {
	thls.cache.Del(eSock)
	if thls.onDisconnected != nil {
		thls.onDisconnected(eSock, err)
	}
}
