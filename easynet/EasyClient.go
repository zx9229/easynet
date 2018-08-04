package easynet

import (
	"net"
	"time"
)

//EasyClient omit
type EasyClient struct {
	EasySocket
	doReconnect bool
	tcpAddr     string
}

//NewEasyClient omit
func NewEasyClient() *EasyClient {
	curData := new(EasyClient)
	return curData
}

//Connect omit
func (thls *EasyClient) Connect(tcpAddr string, doReconnect bool) error {
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

func (thls *EasyClient) reConnect() error {
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

func (thls *EasyClient) innerSendHeartbeat() {
	emptySlice := make([]byte, 0)
	for thls.innerSend(emptySlice, false) == nil {
		time.Sleep(time.Second * 60 * 2)
	}
}

func (thls *EasyClient) actionWhenDis(eSock *EasySocket) {
	if thls.doReconnect {
		go thls.reConnect()
	}
}
