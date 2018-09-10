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

type byte4type [4]byte //用于int32相关
type byte3type [3]byte //用于checksum相关

//EasySocketImpl omit
type EasySocketImpl struct {
	onConnected    EasyConnected    //连接成功的回调
	onDisconnected EasyDisconnected //连接断线的回调
	onMessage      EasyMessage      //收到消息的回调
	data           interface{}      //设置的附加信息
	mutex          sync.Mutex
	isAccepted     bool
	sock           net.Conn
}

func newEasySocketImpl(conn net.Conn) *EasySocketImpl {
	curData := new(EasySocketImpl)
	curData.sock = conn
	return curData
}

func (thls *EasySocketImpl) setIsAccepted(value bool) {
	thls.isAccepted = value
}

//RegEasyConnected omit
func (thls *EasySocketImpl) RegEasyConnected(handler EasyConnected) bool {
	if thls.isAccepted {
		return false
	}
	thls.onConnected = handler
	return true
}

//RegEasyDisConnected omit
func (thls *EasySocketImpl) RegEasyDisConnected(handler EasyDisconnected) bool {
	if thls.isAccepted {
		return false
	}
	thls.onDisconnected = handler
	return true
}

//RegEasyMessage omit
func (thls *EasySocketImpl) RegEasyMessage(handler EasyMessage) bool {
	if thls.isAccepted {
		return false
	}
	thls.onMessage = handler
	return true
}

//SetData omit
func (thls *EasySocketImpl) SetData(v interface{}) {
	thls.data = v
}

//GetData omit
func (thls *EasySocketImpl) GetData() interface{} {
	return thls.data
}

//LocalAddr omit
func (thls *EasySocketImpl) LocalAddr() net.Addr {
	curSock := thls.sock
	if curSock != nil {
		return curSock.LocalAddr()
	}
	return nil
}

//RemoteAddr omit
func (thls *EasySocketImpl) RemoteAddr() net.Addr {
	curSock := thls.sock
	if curSock != nil {
		return curSock.RemoteAddr()
	}
	return nil
}

//IsOnline omit
func (thls *EasySocketImpl) IsOnline() bool {
	return (thls.sock != nil)
}

//Close omit
func (thls *EasySocketImpl) Close() {
	thls.mutex.Lock()
	defer thls.mutex.Unlock()
	if thls.sock != nil {
		thls.sock.Close()
		thls.sock = nil
		//等待recv线程执行onDisconnect回调
	}
}

//Send omit
func (thls *EasySocketImpl) Send(data []byte) error {
	return thls.innerSend(data, true)
}

func (thls *EasySocketImpl) innerSend(data []byte, disableEmpty bool) error {
	if data == nil {
		return nil
	}
	if len(data) == 0 && disableEmpty {
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
func (thls *EasySocketImpl) doRecv(conn net.Conn, act func(eSock *EasySocketImpl)) {
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
		if size == 7 {
			continue
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
func EgOnConnected(eSock EasySocket, isAccepted bool) {
	log.Printf("OnCon[v], %p, A=%v, L=%v, R=%v", eSock, isAccepted, eSock.LocalAddr(), eSock.RemoteAddr())
}

//EgOnDisconnected omit
func EgOnDisconnected(eSock EasySocket, err error) {
	log.Printf("OnDis[x], %p, err=%v", eSock, err)
}

//EgOnMessage omit
func EgOnMessage(eSock EasySocket, data []byte) {
	log.Printf("OnMsg[=], %p, data=%v", eSock, string(data))
}
