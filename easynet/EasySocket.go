package easynet

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"unsafe"
)

var (
	errOffline         = errors.New("offline")
	errSendHalfMessage = errors.New("send half message")
	errMessageIsTooBig = errors.New("message is too big")
	errChecksumIsWrong = errors.New("checksum is wrong")
	errPlaceholder     = errors.New("placeholder")
)

//CbConnected 连接成功的回调函数
type CbConnected func(eSock *EasySocket, isAccepted bool)

//CbDisconnected 连接断开的回调函数
type CbDisconnected func(eSock *EasySocket, err error)

//CbMessage 收到消息的回调函数
type CbMessage func(eSock *EasySocket, data []byte)

type byte4type [4]byte //用于int32相关
type byte3type [3]byte //用于checksum相关

//EasySocket omit
type EasySocket struct {
	onConnected    CbConnected    //连接成功的回调
	onDisconnected CbDisconnected //连接断线的回调
	onMessage      CbMessage      //收到消息的回调
	mutex          sync.Mutex
	isAccepted     bool
	sock           net.Conn
}

func newEasySocket(conn net.Conn) *EasySocket {
	curData := new(EasySocket)
	curData.sock = conn
	return curData
}

func (thls *EasySocket) setIsAccepted(value bool) {
	thls.isAccepted = value
}

//RegCbConnected omit
func (thls *EasySocket) RegCbConnected(handler CbConnected) bool {
	if thls.isAccepted {
		return false
	}
	thls.onConnected = handler
	return true
}

//RegCbDisConnected omit
func (thls *EasySocket) RegCbDisConnected(handler CbDisconnected) bool {
	if thls.isAccepted {
		return false
	}
	thls.onDisconnected = handler
	return true
}

//RegCbMessage omit
func (thls *EasySocket) RegCbMessage(handler CbMessage) bool {
	if thls.isAccepted {
		return false
	}
	thls.onMessage = handler
	return true
}

//LocalAddr omit
func (thls *EasySocket) LocalAddr() net.Addr {
	curSock := thls.sock
	if curSock != nil {
		return curSock.LocalAddr()
	}
	return nil
}

//RemoteAddr omit
func (thls *EasySocket) RemoteAddr() net.Addr {
	curSock := thls.sock
	if curSock != nil {
		return curSock.RemoteAddr()
	}
	return nil
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
		thls.sock = nil
		//等待recv线程执行onDisconnect回调
	}
}

//Send omit
func (thls *EasySocket) Send(data []byte) error {
	if data == nil || len(data) == 0 {
		return nil
	}
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
		//等待recv线程执行onDisconnect回调
		return err
	}
	if num != len(data2) {
		thls.sock.Close()
		thls.sock = nil
		//等待recv线程执行onDisconnect回调
		return errSendHalfMessage
	}
	return nil
}

//doRecv omit
func (thls *EasySocket) doRecv(conn net.Conn, act func(eSock *EasySocket)) {
	thls.mutex.Lock()
	thls.sock = conn
	thls.mutex.Unlock()

	if thls.onConnected != nil {
		thls.onConnected(thls, thls.isAccepted)
	}

	doWhenRecvErr := func(err error) {
		conn.Close()
		thls.mutex.Lock()
		thls.sock = nil
		thls.mutex.Unlock()
		if thls.onDisconnected != nil {
			thls.onDisconnected(thls, err)
		}
		if act != nil {
			act(thls)
		}
	}
	var err error            //错误
	var num int              //本次读取了多少字节
	var cnt int              //本轮读取了多少字节
	var size int             //传输消息有多少字节
	var data []byte          //传输消息的内容(四字节的长度+内容+三字节的校验和)
	var checksumValue string //传输消息的校验和
	var byte4 byte4type      //四字节的长度,存储区
	for {
		cnt = 0
		for cnt < 4 {
			if num, err = conn.Read(byte4[cnt:]); err == nil {
				cnt += num
			} else {
				doWhenRecvErr(err)
				return
			}
		}
		size = int(*(*int32)(unsafe.Pointer(&byte4)) + 3)
		if 10240 < size {
			doWhenRecvErr(errMessageIsTooBig)
			return
		}
		data = make([]byte, size)
		*(*byte4type)(unsafe.Pointer(&data[0])) = byte4
		for cnt < size {
			if num, err = conn.Read(data[cnt:]); err == nil {
				cnt += num
			} else {
				doWhenRecvErr(err)
				return
			}
		}
		checksumValue = checksum(data[:size-3])
		//fmt.Println("checksumValue", checksumValue, string(data[size-3:]))
		if *(*byte3type)(unsafe.Pointer(&data[size-3])) != *(*byte3type)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&checksumValue)).Data)) {
			doWhenRecvErr(errChecksumIsWrong)
			return
		}
		thls.onMessage(thls, data[4:size-3])
	}
}

func checksum(byteSlice []byte) string {
	var sum byte //0~255
	for _, x := range byteSlice {
		sum += x
	}
	//fmt.Println(reflect.StringHeader{}, reflect.SliceHeader{})//string和slice的内存布局.
	return fmt.Sprintf("%03v", sum)
}

func tmpGetSlice(data []byte) []byte {
	//传输消息的内容(四字节的长度+内容+三字节的校验和)
	size := int32(len(data))
	size += 4
	b4 := (*byte4type)(unsafe.Pointer(&size))
	data2 := append((*b4)[:], data...)
	sum := checksum(data2)
	data2 = append(data2, sum[:]...)
	return data2
}

//EgOnConnected omit
func EgOnConnected(eSock *EasySocket, isAccepted bool) {
	log.Printf("OnCon[v], %p, A=%v, L=%v, R=%v", eSock, isAccepted, eSock.LocalAddr(), eSock.RemoteAddr())
}

//EgOnDisconnected omit
func EgOnDisconnected(eSock *EasySocket, err error) {
	log.Printf("OnDis[x], %p, err=%v", eSock, err)
}

//EgOnMessage omit
func EgOnMessage(eSock *EasySocket, data []byte) {
	log.Printf("OnMsg[=], %p, data=%v", eSock, string(data))
}
