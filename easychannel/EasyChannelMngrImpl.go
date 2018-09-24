package easychannel

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/zx9229/easynet/easynet"
)

type operationData struct {
	isClient    bool  //客户端类别的标志(client类别的channel,主动创建的channel)
	id          int64 //
	isCloseRecv bool  //关闭接收(关闭接收chan)
	isCloseBoth bool  //关闭双向(移除这个sockChan)
}

type mapWithLock struct {
	sync.RWMutex
	M map[int64]*EasyChannelImpl
}

//EasyChannelMngrImpl omit
type EasyChannelMngrImpl struct {
	sendToSock   chan []byte
	recvFromSock chan []byte
	opChan       chan *operationData
	eSock        easynet.EasySocket
	srvChannel   *mapWithLock // map[int64]*EasyChannelImpl //server类别的channel,被动创建的channel
	cliChannel   *mapWithLock // map[int64]*EasyChannelImpl //client类别的channel,主动创建的channel
	cliChanIdx   int64
	cbConnected  EasyConnected
}

func newEasyChannelMngrImpl(eSock easynet.EasySocket) EasyChannelManager {
	curData := new(EasyChannelMngrImpl)
	curData.sendToSock = make(chan []byte, 128)
	curData.recvFromSock = make(chan []byte, 128)
	curData.opChan = make(chan *operationData, 2)
	curData.eSock = eSock
	curData.cliChanIdx = 0
	curData.srvChannel = &mapWithLock{M: make(map[int64]*EasyChannelImpl)}
	curData.cliChannel = &mapWithLock{M: make(map[int64]*EasyChannelImpl)}
	curData.cbConnected = nil
	return curData
}

//RegEasyConnected omit
func (thls *EasyChannelMngrImpl) RegEasyConnected(handler EasyConnected) bool {
	if thls.cbConnected != nil {
		return false
	}
	thls.cbConnected = handler
	return true
}

//CreateEasyChannel omit
func (thls *EasyChannelMngrImpl) CreateEasyChannel() (eChannel EasyChannel, err error) {
	cliChannelIdx := atomic.AddInt64(&thls.cliChanIdx, 1)
	curSockChan := newEasyChannelImpl(cliChannelIdx, thls.opChan, thls.sendToSock, true)
	thls.cliChannel.Lock()
	thls.cliChannel.M[cliChannelIdx] = curSockChan
	thls.cliChannel.Unlock()
	eChannel = curSockChan
	return
}

func (thls *EasyChannelMngrImpl) doOneGoroutine() {
	var data []byte
	var opData *operationData
	select {
	case data = <-thls.recvFromSock:
		thls.doRecvData(data)
	case opData = <-thls.opChan:
		thls.doOperate(opData)
	}
}

func (thls *EasyChannelMngrImpl) doOperate(opData *operationData) {
	var mapSafeData *mapWithLock
	if opData.isClient {
		mapSafeData = thls.cliChannel
	} else {
		mapSafeData = thls.srvChannel
	}

	if opData.isCloseRecv {
		mapSafeData.RLock()
		close(mapSafeData.M[opData.id].recvChan)
		mapSafeData.M[opData.id].recvChan = nil
		mapSafeData.RUnlock()
	}
	if opData.isCloseBoth {
		mapSafeData.Lock()
		//TODO:其他清理操作
		delete(mapSafeData.M, opData.id)
		mapSafeData.Unlock()
	}
}

func (thls *EasyChannelMngrImpl) doRecvData(data []byte) {
	//byte(ChannelStatus)|int64(channelIndex)|messageData
	if len(data) < 9 {
		panic("logic_error")
	}
	var mapSafeData *mapWithLock
	curStatus := *(*ChannelStatus)(unsafe.Pointer(&data[0]))
	isClient := curStatus&ChannelStatus_IsClient == ChannelStatus_IsClient
	if isClient {
		mapSafeData = thls.srvChannel       //对端是客户端,本端就是服务端.
		curStatus -= ChannelStatus_IsClient //TODO:不知对错.
	} else {
		mapSafeData = thls.cliChannel
	}

	channelIndex := *(*int64)(unsafe.Pointer(&data[1]))
	mapSafeData.Lock()
	eSockChan, isOk := mapSafeData.M[channelIndex]
	mapSafeData.Unlock()
	if !isOk {
		if isClient && *(*ChannelStatus)(unsafe.Pointer(&data[0])) == ChannelStatus_Open {
			eSockChan = newEasyChannelImpl(channelIndex, thls.opChan, thls.sendToSock, false)
			mapSafeData.Lock()
			mapSafeData.M[channelIndex] = eSockChan
			mapSafeData.Unlock()
			if thls.cbConnected != nil {
				thls.cbConnected(eSockChan)
			}
		} else {
			thls.sendToSock <- generatePackageOpErr(ChannelStatus_NotFound, channelIndex)
		}
	}
	if eSockChan != nil {
		tmpChan := eSockChan.recvChan
		if tmpChan != nil {
			tmpChan <- data
		}
	}
}

func generatePackage(status ChannelStatus, id int64, data []byte) []byte {
	byteSlice := make([]byte, 0)
	byteSlice = append(byteSlice, byte(status))
	byteSlice = append(byteSlice, (*(*[8]byte)(unsafe.Pointer(&id)))[:]...)
	if data != nil {
		byteSlice = append(byteSlice, data...)
	}
	return byteSlice
}

func generatePackageOpErr(status ChannelStatus, id int64) []byte {
	if status&ChannelStatus_OpErr != ChannelStatus_OpErr {
		panic("logic_error")
	}
	return generatePackage(status, id, nil)
}
