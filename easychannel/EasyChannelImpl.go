package easychannel

import (
	"errors"
)

type ChannelStatus byte

const (
	ChannelStatus_NA        ChannelStatus = 0
	ChannelStatus_Open      ChannelStatus = (1 << 0) //2^0=1
	ChannelStatus_Working   ChannelStatus = (1 << 1) //2^1=2
	ChannelStatus_CloseSend ChannelStatus = (1 << 2) //2^2=4  //关闭发送通道
	ChannelStatus_CloseRecv ChannelStatus = (1 << 3) //2^3=8  //关闭接收通道
	ChannelStatus_CloseBoth ChannelStatus = 12       //4+8    //关闭双向通道
	//ChannelStatus_CloseBoth ChannelStatus = (1 << 4) //2^4=16 //关闭双向通道
	ChannelStatus_NotFound ChannelStatus = (1 << 5) //2^5=32 //(对端)找不到这个channel.
	ChannelStatus_Recreate ChannelStatus = (1 << 6) //2^6=64 //(对端)已经存在此channel的ID,这次是重复创建.
)

//EasyChannelImpl omit
type EasyChannelImpl struct {
	//positive   bool  //它是主动创建的,还是被动创建的(true:主动).
	channelIdx int64 //就好像socket的端口号似的.
	status     ChannelStatus
	sendChan   chan []byte
	recvChan   chan []byte
}

func newEasyChannelImpl(id int64, sendChan chan []byte) *EasyChannelImpl {
	curData := new(EasyChannelImpl)
	curData.channelIdx = id
	curData.sendChan = sendChan
	curData.recvChan = make(chan []byte, 128)
	return curData
}

func (thls *EasyChannelImpl) ID() int64 {
	return thls.channelIdx
}

func (thls *EasyChannelImpl) Status() ChannelStatus {
	return thls.status
}

func (thls *EasyChannelImpl) Open() (err error) {
	if thls.status == ChannelStatus_NA {
		thls.sendChan <- generatePackage(ChannelStatus_Open, thls.channelIdx, nil)
		thls.status = ChannelStatus_Open
	} else {
		err = errors.New("状态不对,无法Open")
	}
	return
}

func (thls *EasyChannelImpl) OpenAndSend(data []byte) (err error) {
	//多线程有问题
	if thls.status == ChannelStatus_NA {
		thls.sendChan <- generatePackage(ChannelStatus_Open, thls.channelIdx, data)
		thls.status = ChannelStatus_Open
	} else {
		err = errors.New("状态不对,无法Open")
	}
	return
}

func (thls *EasyChannelImpl) Send(data []byte) (err error) {
	if thls.status&ChannelStatus_CloseSend == ChannelStatus_NA {
		thls.sendChan <- generatePackage(ChannelStatus_Working, thls.channelIdx, data)
	} else {
		err = errors.New("状态不对,无法Send")
	}
	return
}

func (thls *EasyChannelImpl) SendAndClose(data []byte, closeSend bool, closeRecv bool) (err error) {
	//多线程有问题
	if thls.status&ChannelStatus_CloseSend == ChannelStatus_NA {
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
		thls.sendChan <- generatePackage(willStatus, thls.channelIdx, data)
	} else {
		err = errors.New("状态不对,无法Send")
	}
	return
}

func (thls *EasyChannelImpl) Close(closeSend bool, closeRecv bool) {
	//多线程有问题
	willStatus := ChannelStatus_NA
	if closeSend { //关闭本端的发送
		willStatus &= ChannelStatus_CloseRecv //关闭对端的接收
	}
	if closeRecv {
		willStatus &= ChannelStatus_CloseSend
	}
	if thls.status&willStatus != willStatus {
		thls.sendChan <- generatePackage(willStatus, thls.channelIdx, nil)
		thls.status &= willStatus
	}
}

func (thls *EasyChannelImpl) Recv() (data []byte, err error) {
	if thls.status&ChannelStatus_CloseRecv != ChannelStatus_CloseRecv {
		var isOk bool
		if data, isOk = <-thls.recvChan; !isOk {
			err = errors.New("chan被关闭了")
			return
		}

		panic("")
		//manager判断了ChannelStatus这里就不用再判断了.
		if 9 <= len(data) {

		}
	} else {
		err = errors.New("状态不对,无法Send")
	}
	return
}
