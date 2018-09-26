package easynet2

var (
	action_NA          byte = 0
	action_Open        byte = (1 << 0) //2^0=1
	action_Working     byte = (1 << 1) //2^1=2
	action_CloseSend   byte = (1 << 2) //2^2=4   //关闭发送通道
	action_CloseRecv   byte = (1 << 3) //2^3=8   //关闭接收通道
	action_CloseBoth   byte = 4 + 8    //4+8     //关闭双向通道
	action_OpErr       byte = (1 << 6) //2^6=64  //操作错误
	action_ReOpen      byte = action_OpErr + 1
	action_NotFound    byte = action_OpErr + 2    //误报几率很高
	action_ReCloseSend byte = action_OpErr + 3    //误报几率很高
	action_ReCloseRecv byte = action_OpErr + 4    //误报几率很高
	action_ReCloseBoth byte = action_OpErr + 5    //误报几率很高
	flag_IsSession     byte = (1 << 7)            //2^7=128  //标志这个消息包是session的
	flag_IsAccepted    byte = (1 << 6)            //2^6=64   //标志这个消息包是server类别的
	flag_Mask          byte = flag_IsAccepted - 1 //2^6=64   //标志掩码(获取真实数据的掩码)
)

//EasySessionImpl omit
type easySessionImpl struct {
	sock        *EasySocketImpl
	id          int64
	isAccepted  bool
	isOpen      bool
	isCloseSend bool
	isCloseRecv bool
}

func newEasySessionImpl(eSock *EasySocketImpl, idx int64, isAccepted bool) *easySessionImpl {
	curData := new(easySessionImpl)
	curData.sock = eSock
	curData.id = idx
	curData.isAccepted = isAccepted
	curData.isOpen = false
	curData.isCloseSend = false
	curData.isCloseRecv = false
	return curData
}

//ID omit
func (thls *easySessionImpl) ID() int64 {
	return thls.id
}

//IsAccepted omit
func (thls *easySessionImpl) IsAccepted() bool {
	return thls.isAccepted
}

//Close omit
func (thls *easySessionImpl) Close(closeSend bool, closeRecv bool) {
	if !thls.isOpen {
		return
	}
	if thls.isCloseRecv && thls.isCloseSend {
		return
	}
	curAction := action_NA
	if !thls.isCloseRecv && closeRecv {
		curAction |= action_CloseRecv
		thls.isCloseRecv = true
	}
	if !thls.isCloseSend && closeSend {
		curAction |= action_CloseSend
		thls.isCloseSend = true
	}
	if curAction != action_NA {
		thls.sock.innerSend2(nil, thls.id, thls.isAccepted, curAction)
	}
	if thls.isCloseRecv && thls.isCloseSend {
		thls.sock.sessManager.deleteSession(thls.id, thls.isAccepted)
	}
}

//Send omit
func (thls *easySessionImpl) Send(data []byte) error {
	if thls.isOpen && !thls.isCloseSend {
		return thls.sock.innerSend2(data, thls.id, thls.isAccepted, action_Working)
	} else {
		return nil //TODO:
	}
}

//Recv omit
func (thls *easySessionImpl) Recv() (data []byte, err error) {
	panic("")
	return
}

//Socket omit
func (thls *easySessionImpl) Socket() EasySocket {
	return thls.sock
}
