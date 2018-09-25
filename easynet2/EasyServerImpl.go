package easynet2

import (
	"net"
	"sync"
)

type safeSockSet struct {
	sync.Mutex
	M map[EasySocket]bool
}

//EasyServerImpl omit
type EasyServerImpl struct {
	onConnected    EasyConnected    //连接成功的回调
	onDisconnected EasyDisconnected //连接断线的回调
	onMessage      EasyMessage      //收到消息的回调
	listener       net.Listener
	cache          *safeSockSet
}

//NewEasyServerImpl omit
func NewEasyServerImpl() *EasyServerImpl {
	curData := new(EasyServerImpl)
	curData.cache = &safeSockSet{M: make(map[EasySocket]bool)}
	return curData
}

//RegEasyConnected omit
func (thls *EasyServerImpl) RegEasyConnected(handler EasyConnected) bool {
	thls.onConnected = handler
	return true
}

//RegEasyDisConnected omit
func (thls *EasyServerImpl) RegEasyDisConnected(handler EasyDisconnected) bool {
	thls.onDisconnected = handler
	return true
}

//RegEasyMessage omit
func (thls *EasyServerImpl) RegEasyMessage(handler EasyMessage) bool {
	thls.onMessage = handler
	return true
}

//Run omit
func (thls *EasyServerImpl) Run(tcpAddr string) error {
	var err error
	if thls.listener, err = net.Listen("tcp", tcpAddr); err != nil {
		return err
	}
	var conn net.Conn
	for {
		if conn, err = thls.listener.Accept(); err != nil {
			return err
		}
		eSock := newEasySocketImpl(conn)
		eSock.RegEasyConnected(thls.onConnected)
		eSock.RegEasyDisConnected(thls.onDisconnected)
		eSock.RegEasyMessage(thls.onMessage)
		eSock.setIsAccepted(true)
		thls.cache.Lock()
		thls.cache.M[eSock] = true
		thls.cache.Unlock()
		go eSock.doRecv(conn, thls.actionWhenDis)
	}
}

func (thls *EasyServerImpl) actionWhenDis(eSock *EasySocketImpl) {
	thls.cache.Lock()
	delete(thls.cache.M, eSock)
	thls.cache.Unlock()
}
