package easychannel

import (
	"errors"
	"unsafe"
)

type ChannelStatus byte

const (
	ChannelStatus_NA          ChannelStatus = 0
	ChannelStatus_Open        ChannelStatus = (1 << 0) //2^0=1
	ChannelStatus_Working     ChannelStatus = (1 << 1) //2^1=2
	ChannelStatus_CloseSend   ChannelStatus = (1 << 2) //2^2=4   //关闭发送通道
	ChannelStatus_CloseRecv   ChannelStatus = (1 << 3) //2^3=8   //关闭接收通道
	ChannelStatus_CloseBoth   ChannelStatus = 12       //4+8     //关闭双向通道
	ChannelStatus_OpErr       ChannelStatus = (1 << 6) //2^6=64  //操作错误
	ChannelStatus_NotFound    ChannelStatus = ChannelStatus_OpErr + 1
	ChannelStatus_ReOpen      ChannelStatus = ChannelStatus_OpErr + 2
	ChannelStatus_ReCloseSend ChannelStatus = ChannelStatus_OpErr + 3
	ChannelStatus_ReCloseRecv ChannelStatus = ChannelStatus_OpErr + 4
	ChannelStatus_ReCloseBoth ChannelStatus = ChannelStatus_OpErr + 5
	ChannelStatus_IsClient    ChannelStatus = (1 << 7) //2^7=128 //客户端标志
)

//EasyChannelImpl omit
type EasyChannelImpl struct {
	clientFlag ChannelStatus //它是主动创建的,还是被动创建的(true:主动).
	channelIdx int64         //就好像socket的端口号似的.
	status     ChannelStatus
	opChan     chan *operationData
	sendChan   chan []byte
	recvChan   chan []byte
}

func newEasyChannelImpl(id int64, opChan chan *operationData, sendChan chan []byte, isClient bool) *EasyChannelImpl {
	//相当于"创建了一个socket,这个socket没有open"
	curData := new(EasyChannelImpl)
	if isClient {
		curData.clientFlag = ChannelStatus_IsClient
	} else {
		curData.clientFlag = ChannelStatus_NA
	}
	curData.channelIdx = id
	curData.status = ChannelStatus_NA
	curData.opChan = opChan
	curData.sendChan = sendChan
	curData.recvChan = make(chan []byte, 128)
	return curData
}

//ID omit
func (thls *EasyChannelImpl) ID() int64 {
	return thls.channelIdx
}

//Status omit
func (thls *EasyChannelImpl) Status() ChannelStatus {
	return thls.status
}

//Open omit
func (thls *EasyChannelImpl) Open() (err error) {
	//多线程有问题
	if thls.status == ChannelStatus_NA {
		thls.sendChan <- generatePackage(ChannelStatus_Open|thls.clientFlag, thls.channelIdx, nil)
		thls.status = ChannelStatus_Open
	} else {
		err = errors.New("状态不对,无法Open")
	}
	return
}

//OpenAndSend omit
func (thls *EasyChannelImpl) OpenAndSend(data []byte) (err error) {
	//多线程有问题
	if thls.status == ChannelStatus_NA {
		thls.sendChan <- generatePackage(ChannelStatus_Open|thls.clientFlag, thls.channelIdx, data)
		thls.status = ChannelStatus_Open
	} else {
		err = errors.New("状态不对,无法Open")
	}
	return
}

//Send omit
func (thls *EasyChannelImpl) Send(data []byte) (err error) {
	isOpen := thls.status&ChannelStatus_Open == ChannelStatus_Open
	isCloseSend := thls.status&ChannelStatus_CloseSend == ChannelStatus_CloseSend
	isOpErr := thls.status&ChannelStatus_OpErr == ChannelStatus_OpErr
	if isOpen && !isCloseSend && !isOpErr {
		thls.sendChan <- generatePackage(ChannelStatus_Working|thls.clientFlag, thls.channelIdx, data)
	} else {
		err = errors.New("状态不对,无法Send")
	}
	return
}

//SendAndClose omit
func (thls *EasyChannelImpl) SendAndClose(data []byte, closeSend bool, closeRecv bool) (err error) {
	//多线程有问题
	isOpen := thls.status&ChannelStatus_Open == ChannelStatus_Open
	isCloseSend := thls.status&ChannelStatus_CloseSend == ChannelStatus_CloseSend
	isOpErr := thls.status&ChannelStatus_OpErr == ChannelStatus_OpErr
	if isOpen && !isCloseSend && !isOpErr {
		willStatus := ChannelStatus_NA
		if closeSend {
			willStatus &= ChannelStatus_CloseRecv  //关闭对端的接收
			thls.status &= ChannelStatus_CloseSend //关闭本端的发送
		}
		if closeRecv {
			willStatus &= ChannelStatus_CloseSend
			thls.status &= ChannelStatus_CloseRecv
		}
		if willStatus == ChannelStatus_NA {
			willStatus = ChannelStatus_Working
		}
		thls.sendChan <- generatePackage(willStatus|thls.clientFlag, thls.channelIdx, data)
	} else {
		err = errors.New("状态不对,无法Send")
	}
	return
}

//Close omit
func (thls *EasyChannelImpl) Close(closeSend bool, closeRecv bool) {
	//多线程有问题
	isOpen := thls.status&ChannelStatus_Open == ChannelStatus_Open
	isOpErr := thls.status&ChannelStatus_OpErr == ChannelStatus_OpErr
	willStatus := ChannelStatus_NA
	if closeSend { //关闭本端的发送
		willStatus &= ChannelStatus_CloseRecv //关闭对端的接收
	}
	if closeRecv {
		willStatus &= ChannelStatus_CloseSend
	}
	if isOpen && !isOpErr && thls.status&willStatus != willStatus {
		thls.sendChan <- generatePackage(willStatus|thls.clientFlag, thls.channelIdx, nil)
		thls.status &= willStatus
	}
}

//Recv omit
func (thls *EasyChannelImpl) Recv() (data []byte, err error) {
	//先recv出来data，然后执行opCode操作，这样就能比较好的处理了。
	var isOpen, isCloseRecv, isOpErr bool
	var isOk bool
	var statusInData ChannelStatus
	isOpen = true //我们先假定所有channel都是被动创建的,那么它必定打开过了
	for {
		if thls.clientFlag == ChannelStatus_IsClient {
			//如果被动创建,那么肯定open了,只是标志位没有设置过来罢了,
			//如果主动创建,需要它open了,才能接收消息.
			isOpen = thls.status&ChannelStatus_Open == ChannelStatus_Open
		}
		isCloseRecv = thls.status&ChannelStatus_CloseRecv == ChannelStatus_CloseRecv
		isOpErr = thls.status&ChannelStatus_OpErr == ChannelStatus_OpErr
		if (isOpen && !isCloseRecv && !isOpErr) == false {
			err = errors.New("状态不对,无法Recv")
			break
		}
		if data, isOk = <-thls.recvChan; !isOk {
			err = errors.New("chan被关闭了")
			break
		}
		statusInData = *(*ChannelStatus)(unsafe.Pointer(&data[0]))
		if statusInData&ChannelStatus_OpErr == ChannelStatus_OpErr {
			thls.status = thls.status & ChannelStatus_OpErr
			//操作是不应当出错的,这里先不处理,先让它崩溃.
			panic(statusInData)
		}
		switch statusInData {
		case ChannelStatus_Open:
			if thls.status&ChannelStatus_Open == ChannelStatus_Open {
				thls.sendChan <- generatePackageOpErr(ChannelStatus_ReOpen|thls.clientFlag, thls.channelIdx)
				continue
			}
			thls.status = thls.status & ChannelStatus_Open
		case ChannelStatus_Working:
		case ChannelStatus_CloseSend:
			if thls.status&ChannelStatus_CloseSend == ChannelStatus_CloseSend {
				thls.sendChan <- generatePackageOpErr(ChannelStatus_ReCloseSend|thls.clientFlag, thls.channelIdx)
				continue
			}
			thls.status = thls.status & ChannelStatus_CloseSend
		case ChannelStatus_CloseRecv:
			if thls.status&ChannelStatus_CloseRecv == ChannelStatus_CloseRecv {
				thls.sendChan <- generatePackageOpErr(ChannelStatus_CloseRecv|thls.clientFlag, thls.channelIdx)
				continue
			}
			thls.status = thls.status & ChannelStatus_CloseRecv
		case ChannelStatus_CloseBoth:
			if thls.status&ChannelStatus_CloseBoth == ChannelStatus_CloseBoth {
				thls.sendChan <- generatePackageOpErr(ChannelStatus_CloseBoth|thls.clientFlag, thls.channelIdx)
				continue
			}
			thls.status = thls.status & ChannelStatus_CloseBoth
		default:
			panic("")
		}
		data = data[8:]
		if len(data) == 0 {
			continue
		}
		break
	}
	return
}

/*
客户端socket连接服务器，服务器accept连接之后，立即发送数个字节的信息，然后立即close掉这个连接。
客户端socket连接成功后，睡眠数秒，然后recv，可以发现，socket先recv出来信息，err此时为nil，第二次err才有值。
我也准备采取这种方式。
*/
