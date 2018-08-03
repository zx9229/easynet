package easynet

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"unsafe"
)

//EasyConn omit
type EasyConn interface {
	RegCbConnected(handler CbConnected) bool
	RegCbDisConnected(handler CbDisconnected) bool
	RegCbMessage(handler CbMessage) bool
	IsOnline() bool
	Close()
	Send(data []byte) error
}

var (
	errOffline         = errors.New("Offline")
	errSendHalfMessage = errors.New("send half message")
	errMessageIsTooBig = errors.New("message packet is too big")
	errChecksumIsWrong = errors.New("checksum is wrong")
	errPlaceholder     = errors.New("placeholder")
)

//CbConnected 连接成功的回调函数
type CbConnected func(eSock *EasySocket, isAccepted bool)

//CbDisconnected 连接断开的回调函数
type CbDisconnected func(eSock *EasySocket, err error)

//CbMessage 收到消息的回调函数
type CbMessage func(eSock *EasySocket, data []byte)

type extraAction func(eSock *EasySocket, err error)

type byte4type [4]byte

//EasySocket omit
type EasySocket struct {
	onConnected    CbConnected    //连接成功的回调
	onDisconnected CbDisconnected //连接断线的回调
	onMessage      CbMessage      //收到消息的回调
	sock           net.Conn
	mutex          sync.Mutex
}

func newEasySocket(conn net.Conn) *EasySocket {
	curData := new(EasySocket)
	curData.sock = conn
	return curData
}

//RegCbConnected omit
func (thls *EasySocket) RegCbConnected(handler CbConnected) bool {
	thls.onConnected = handler
	return true
}

//RegCbDisConnected omit
func (thls *EasySocket) RegCbDisConnected(handler CbDisconnected) bool {
	thls.onDisconnected = handler
	return true
}

//RegCbMessage omit
func (thls *EasySocket) RegCbMessage(handler CbMessage) bool {
	thls.onMessage = handler
	return true
}

//IsOnline omit
func (thls *EasySocket) IsOnline() bool {
	return (thls.sock != nil)
}

//Close omit
func (thls *EasySocket) Close() {
	thls.mutex.Lock()
	defer thls.mutex.Unlock()
	if thls.sock != nil {
		thls.sock.Close()
	}
}

//Send omit
func (thls *EasySocket) Send(data []byte) error {
	thls.mutex.Lock()
	defer thls.mutex.Unlock()
	if thls.sock == nil {
		return errOffline
	}
	data2 := tmpGetSlice(data)
	num, err := thls.sock.Write(data2)
	if err != nil {
		thls.sock.Close()
		thls.sock = nil
		return err
	}
	if num != len(data2) {
		thls.sock.Close()
		thls.sock = nil
		return errSendHalfMessage
	}
	return nil
}

//doRecv omit
func (thls *EasySocket) doRecv(conn net.Conn, isAccepted bool, act extraAction) {
	thls.mutex.Lock()
	thls.sock = conn
	thls.mutex.Unlock()

	if thls.onConnected != nil {
		thls.onConnected(thls, isAccepted)
	}

	doReXyz := func(err error) {
		conn.Close()
		thls.mutex.Lock()
		thls.sock = nil
		thls.mutex.Unlock()
		if thls.onDisconnected != nil {
			thls.onDisconnected(thls, err)
		}
		if act != nil {
			act(thls, err)
		}
	}

	var err error
	var num int
	var cnt int
	byte4 := byte4type{}
	for {
		cnt = 0
		for cnt < 4 {
			if num, err = conn.Read(byte4[cnt:]); err == nil {
				cnt += num
			} else {
				doReXyz(err)
				return
			}
		}
		size := *(*int32)(unsafe.Pointer(&byte4))
		if 10240 < size {
			doReXyz(errMessageIsTooBig)
			return
		}
		data := make([]byte, size+3) //数据+3位的checksum
		for i, b := range byte4 {
			data[i] = b
		}
		cnt = 4
		for cnt < int(size+3) {
			if num, err = conn.Read(data[cnt:]); err == nil {
				cnt += num
			} else {
				doReXyz(err)
				return
			}
		}
		csum := checksum(data[:size])
		if string(data[size:]) != csum {
			doReXyz(errChecksumIsWrong)
			return
		}
		thls.onMessage(thls, data[4:size])
	}
}

func checksum(byteSlice []byte) string {
	var sum byte //0~255
	for _, x := range byteSlice {
		sum += x
	}
	return fmt.Sprintf("%03v", sum)
}

func tmpGetSlice(data []byte) []byte {
	size := int32(len(data))
	size += 4
	b4 := (*byte4type)(unsafe.Pointer(&size))
	data2 := append((*b4)[:], data...)
	sum := checksum(data2)
	data2 = append(data2, []byte(sum)...)
	return data2
}

//EgOnConnected omit
func EgOnConnected(eSock *EasySocket, isAccepted bool) {
	log.Printf("OnCon[v], %p, isAccepted=%v", eSock, isAccepted)
}

//EgOnDisconnected omit
func EgOnDisconnected(eSock *EasySocket, err error) {
	log.Printf("OnDis[x], %p, err=%v", eSock, err)
}

//EgOnMessage omit
func EgOnMessage(eSock *EasySocket, data []byte) {
	log.Printf("OnMsg[=], %p, data=%v", eSock, string(data))
}
