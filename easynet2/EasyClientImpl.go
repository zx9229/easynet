package easynet2

import (
	"net"
	"time"
)

//EasyClientImpl omit
type EasyClientImpl struct {
	EasySocketImpl
	doReconnect bool
	tcpAddr     string
}

//NewEasyClientImpl omit
func NewEasyClientImpl() EasyClient {
	curData := new(EasyClientImpl)
	curData.EasySocketImpl.sessManager = newEasySessionManager(&curData.EasySocketImpl)
	return curData
}

//Connect omit
func (thls *EasyClientImpl) Connect(tcpAddr string, doReconnect bool) error {
	var err error
	thls.tcpAddr = tcpAddr
	thls.doReconnect = doReconnect
	if thls.doReconnect {
		go thls.reConnect()
	} else {
		err = thls.reConnect()
	}
	return err
}

func (thls *EasyClientImpl) reConnect() error {
	var conn net.Conn
	err := errPlaceholder
	for err != nil {
		if conn, err = net.Dial("tcp", thls.tcpAddr); err != nil {
			time.Sleep(time.Second * 5)
		} else {
			go thls.doRecv(conn, thls.actionWhenDis)
			go thls.innerSendHeartbeat()
		}
		if !thls.doReconnect {
			break
		}
	}
	return err
}

func (thls *EasyClientImpl) innerSendHeartbeat() {
	emptySlice := make([]byte, 0)
	for thls.innerSend(emptySlice, false) == nil {
		time.Sleep(time.Second * 60 * 2)
	}
}

func (thls *EasyClientImpl) actionWhenDis(eSock *EasySocketImpl) {
	if thls.doReconnect {
		go thls.reConnect()
	}
}
