package easychannel

import (
	"unsafe"

	"github.com/zx9229/easynet/easynet"
)

//EasyChannelMngrImpl omit
type EasyChannelMngrImpl struct {
	sendChan    chan []byte
	eSock       easynet.EasySocket
	channelIdx  int64
	srvChannel  map[int64]*EasyChannelImpl //server类别的channel,被动创建的channel
	cliChannel  map[int64]*EasyChannelImpl //client类别的channel,主动创建的channel
	cbConnected EasyConnected
}

func newEasyChannelMngrImpl(eSock easynet.EasySocket) EasyChannelManager {
	curData := new(EasyChannelMngrImpl)
	curData.sendChan = make(chan []byte, 128)
	curData.eSock = eSock
	curData.channelIdx = 9
	curData.cliChannel = make(map[int64]*EasyChannelImpl)
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
	thls.channelIdx++
	curSockChan := newEasyChannelImpl(thls.channelIdx, thls.sendChan)
	thls.cliChannel[thls.channelIdx] = curSockChan
	eChannel = curSockChan
	return
}

func (thls *EasyChannelMngrImpl) doRecvData(data []byte) {
	//发过来的数据中,第一个字符,表征"这个消息包是做什么用的"
	switch ChannelStatus(data[0]) {
	case ChannelStatus_NA:
		panic("")
	case ChannelStatus_Open:
		thls.doChannelStatusOpen(data)
	case ChannelStatus_Working:
		thls.doChannelStatusWorking(data)
	case ChannelStatus_CloseSend:
	case ChannelStatus_CloseRecv:
	case ChannelStatus_CloseBoth:
	case ChannelStatus_NotFound:
	case ChannelStatus_Recreate:
	default:
		panic("")
	}
}

func (thls *EasyChannelMngrImpl) doChannelStatusWorking(data []byte) {
	channelIdx := *(*int64)(unsafe.Pointer(&data[1]))
	var isOk bool
	var eSockChan EasyChannel
	if eSockChan, isOk = thls.srvChannel[channelIdx]; !isOk {
		thls.sendChan <- generatePackage(ChannelStatus_NotFound, channelIdx, nil)
		return
	}
	if eSockChan.Status()&ChannelStatus_CloseRecv == ChannelStatus_CloseRecv {
		//这一端关闭了接收通道,按照逻辑,已经发送了"关闭对端的发送通道"消息,只是这个消息还没有被对端处理,
		//这里不再重复发送对应的消息,以规避冗余消息
		//thls.sendChan <- generatePackage(ChannelStatus_CloseSend, channelIdx, nil)
		return
	}

}

func (thls *EasyChannelMngrImpl) doChannelStatusOpen(data []byte) {
	channelIdx := *(*int64)(unsafe.Pointer(&data[1]))
	if _, ok := thls.srvChannel[channelIdx]; ok {
		thls.sendChan <- generatePackage(ChannelStatus_Recreate, channelIdx, nil)
		return
	}
	eSockChan := newEasyChannelImpl(channelIdx, thls.sendChan)
	thls.srvChannel[channelIdx] = eSockChan
	if thls.cbConnected != nil {
		thls.cbConnected(eSockChan)
	}
}

func generatePackage(status ChannelStatus, id int64, data []byte) []byte {
	byteSlice := make([]byte, 0)
	byteSlice = append(byteSlice, byte(status))
	//byteSlice = append(byteSlice, )
	byteSlice = append(byteSlice, (*(*[8]byte)(unsafe.Pointer(&id)))[:]...)
	if data != nil {
		byteSlice = append(byteSlice, data...)
	}
	return byteSlice
}
