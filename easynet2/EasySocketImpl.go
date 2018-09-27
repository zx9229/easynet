package easynet2

import (
	"errors"
	"log"
	"net"
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
type byte8type [8]byte //用于int64相关

//EasySocketImpl omit
type EasySocketImpl struct {
	onConnected    EasyConnected    //连接成功的回调
	onDisconnected EasyDisconnected //连接断线的回调
	onMessage      EasyMessage      //收到消息的回调
	data           interface{}      //设置的附加信息
	mutex          sync.Mutex
	isAccepted     bool
	sock           net.Conn
	sessManager    *easySessionManager
}

func newEasySocketImpl(conn net.Conn) *EasySocketImpl {
	curData := new(EasySocketImpl)
	curData.sock = conn
	curData.sessManager = newEasySessionManager(curData)
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

func (thls *EasySocketImpl) innerSend2(data []byte, sessionID int64, isAccepted bool, operateData byte) error {
	thls.mutex.Lock()
	defer thls.mutex.Unlock()
	if thls.sock == nil {
		return errOffline
	}
	data2 := zxTmpInfo2Data(data, sessionID, isAccepted, operateData)
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
		thls.onConnected(thls, thls.isAccepted, nil, false)
	}

	doWhenRecvErr := func(err error) {
		conn.Close()
		thls.mutex.Lock()
		thls.sock = nil
		thls.mutex.Unlock()
		if thls.onDisconnected != nil {
			thls.onDisconnected(thls, nil, err, true)
		}
		if act != nil {
			act(thls)
		}
	}
	var err error         //错误
	var num int           //本次读取了多少字节
	var cnt int           //本轮读取了多少字节
	var size int          //传输消息有多少字节
	var data []byte       //传输消息的内容(四字节的长度+内容)
	var byte4 byte4type   //四字节的长度,存储区
	var byte4Ex byte4type //四字节的长度,存储区
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
		byte4Ex = byte4
		byte4Ex[0] &= flag_Mask
		size = int(*(*int32)(unsafe.Pointer(&byte4Ex)))
		if 10240000 < size {
			doWhenRecvErr(errMessageIsTooBig)
			return
		}
		if size == 4 {
			continue
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

		if true {
			var sess *easySessionImpl
			msgData, sessionID, peerIsAccepted, operateData := zxTmpData2Info(data)
			if sessionID != 0 {
				sess = thls.sessManager.operateSession(sessionID, peerIsAccepted, operateData)
			}
			thls.onMessage(thls, sess, msgData)
		}
	}
}

//CreateSession omit
func (thls *EasySocketImpl) CreateSession() EasySession {
	return thls.sessManager.CreateSession()
}

func tmpGetSlice(data []byte) []byte {
	//传输消息的内容(四字节的长度+内容)
	size := int32(len(data))
	size += 4
	b4 := (*byte4type)(unsafe.Pointer(&size))
	return append((*b4)[:], data...)
}

func zxTmpInfo2Data(data []byte, sessionID int64, isAccepted bool, operateData byte) []byte {
	txDataBody := make([]byte, 0)
	txDataSize := int32(len(data))
	if sessionID != 0 {
		txDataSize = 4 + 1 + 8 + txDataSize
		txDataBody = append(txDataBody, (*byte4type)(unsafe.Pointer(&txDataSize))[:]...)
		txDataBody = append(txDataBody, operateData)
		txDataBody = append(txDataBody, (*byte8type)(unsafe.Pointer(&sessionID))[:]...)
		txDataBody = append(txDataBody, data...)
		if isAccepted {
			txDataBody[0] = txDataBody[0] | flag_IsSession | flag_IsAccepted
		} else {
			txDataBody[0] = txDataBody[0] | flag_IsSession
		}
	} else {
		txDataSize = 4 + txDataSize
		txDataBody = append(txDataBody, (*byte4type)(unsafe.Pointer(&txDataSize))[:]...)
		txDataBody = append(txDataBody, data...)
	}
	return txDataBody
}

func zxTmpData2Info(txData []byte) (data []byte, sessionID int64, isAccepted bool, operateData byte) {
	if txData[0]&flag_IsSession == flag_IsSession {
		operateData = txData[4]
		sessionID = *(*int64)(unsafe.Pointer(&txData[5]))
		isAccepted = txData[0]&flag_IsAccepted == flag_IsAccepted
	} else {
		data = txData[4:]
	}
	return
}

//EgOnConnected omit
func EgOnConnected(eSock EasySocket, isAccepted bool, eSess EasySession, sessAccepted bool) {
	log.Printf("OnCon[v], %p|%p A=%v|%v, L=%v, R=%v", eSock, eSess, isAccepted, sessAccepted, eSock.LocalAddr(), eSock.RemoteAddr())
}

//EgOnDisconnected omit
func EgOnDisconnected(eSock EasySocket, eSess EasySession, err error, byDisconnected bool) {
	log.Printf("OnDis[x], %p|%p, err=%v", eSock, eSess, err)
}

//EgOnMessage omit
func EgOnMessage(eSock EasySocket, eSess EasySession, data []byte) {
	log.Printf("OnMsg[=], %p|%p, data=%v", eSock, eSess, string(data))
}
