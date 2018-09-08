package easynet

import (
	"net"
	"time"
)

//EasyClient3 omit
type EasyClient3 struct {
	EasySocket3
	doReconnect bool
	tcpAddr     string
}

//NewEasyClient3 omit
func NewEasyClient3() EasyClient {
	curData := new(EasyClient3)
	return curData
}

//Connect omit
func (thls *EasyClient3) Connect(tcpAddr string, doReconnect bool) error {
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

func (thls *EasyClient3) reConnect() error {
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

func (thls *EasyClient3) innerSendHeartbeat() {
	emptySlice := make([]byte, 0)
	for thls.innerSend(emptySlice, false) == nil {
		time.Sleep(time.Second * 60 * 2)
	}
}

func (thls *EasyClient3) actionWhenDis(eSock *EasySocket3) {
	if thls.doReconnect {
		go thls.reConnect()
	}
}
