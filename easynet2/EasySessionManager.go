package easynet2

import (
	"sync"
	"sync/atomic"
)

type mapWithLock struct {
	sync.RWMutex
	M map[int64]*easySessionImpl
}

//easySessionManager omit
type easySessionManager struct {
	sock       *EasySocketImpl
	srvSession *mapWithLock
	cliSession *mapWithLock
	cliSessIdx int64
}

func newEasySessionManager(eSock *EasySocketImpl) *easySessionManager {
	curData := new(easySessionManager)
	curData.sock = eSock
	curData.srvSession = &mapWithLock{M: make(map[int64]*easySessionImpl)}
	curData.cliSession = &mapWithLock{M: make(map[int64]*easySessionImpl)}
	curData.cliSessIdx = 0
	return curData
}

func (thls *easySessionManager) CreateSession() *easySessionImpl {
	//TODO:如果socket已经断线了呢
	cliChannelIdx := atomic.AddInt64(&thls.cliSessIdx, 1)
	sess := newEasySessionImpl(thls.sock, cliChannelIdx, false)
	thls.cliSession.Lock()
	thls.cliSession.M[cliChannelIdx] = sess
	thls.cliSession.Unlock()
	sess.sock.innerSend2(nil, sess.id, sess.isAccepted, action_Open)
	thls.sock.onConnected(thls.sock, thls.sock.isAccepted, sess, sess.isAccepted)
	return sess
}

func (thls *easySessionManager) deleteSession(idx int64, isAccepted bool) {
	var safeData *mapWithLock
	if isAccepted {
		safeData = thls.srvSession
	} else {
		safeData = thls.cliSession
	}
	safeData.Lock()
	delete(safeData.M, idx)
	safeData.Unlock()
}

func (thls *easySessionManager) operateSession(idx int64, peerIsAccepted bool, action byte) *easySessionImpl {
	var safeData *mapWithLock
	if peerIsAccepted { //对端是被动创建的,那么,本端是主动创建的.
		safeData = thls.cliSession
	} else {
		safeData = thls.srvSession
	}

	var sess *easySessionImpl
	var isOk bool
	safeData.RLock()
	sess, isOk = safeData.M[idx]
	safeData.RUnlock()

	switch action {
	case action_NA:
		thls.sock.innerSend2(nil, idx, !peerIsAccepted, action_OpErr)
	case action_Open:
		if peerIsAccepted { //对端是被动创建的,本端是主动创建的,逻辑上不会有open本端的操作.
			thls.sock.innerSend2(nil, idx, !peerIsAccepted, action_OpErr)
			break
		}
		if isOk {
			thls.sock.innerSend2(nil, idx, !peerIsAccepted, action_ReOpen)
			break
		}
		sess = newEasySessionImpl(thls.sock, idx, !peerIsAccepted)
		safeData.Lock()
		safeData.M[idx] = sess
		safeData.Unlock()
		thls.sock.onConnected(thls.sock, thls.sock.isAccepted, sess, sess.isAccepted)
	case action_Working:
		//nothing.
	case action_CloseSend:
		if !isOk {
			//参考socket,一段时间之内,还会返回错误,时间过长,就直接丢弃了,这里可以选择直接丢弃的.
		}
		if sess.isCloseSend {
			break
		}
		sess.isCloseSend = true
		if sess.isCloseRecv && sess.isCloseSend {
			thls.deleteSession(idx, !peerIsAccepted)
			thls.sock.onDisconnected(thls.sock, sess, nil, false)
		}
	case action_CloseRecv:
		if !isOk {
			//参考socket,一段时间之内,还会返回错误,时间过长,就直接丢弃了,这里可以选择直接丢弃的.
		}
		if sess.isCloseRecv {
			break
		}
		sess.isCloseRecv = true
		if sess.isCloseRecv && sess.isCloseSend {
			thls.deleteSession(idx, !peerIsAccepted)
			thls.sock.onDisconnected(thls.sock, sess, nil, false)
		}
	case action_CloseBoth:
		if !isOk {
			//参考socket,一段时间之内,还会返回错误,时间过长,就直接丢弃了,这里可以选择直接丢弃的.
		}
		if sess.isCloseSend && sess.isCloseRecv {
			break
		}
		sess.isCloseSend = true
		sess.isCloseRecv = true
		thls.deleteSession(idx, !peerIsAccepted)
		thls.sock.onDisconnected(thls.sock, sess, nil, false)
	case action_OpErr:
		//本端发出操作,对端返回[操作错误],本端处理这个情况.
		if !isOk {
			//参考socket,一段时间之内,还会返回错误,时间过长,就直接丢弃了,这里可以选择直接丢弃的.
		}
		panic("操作错误")
	case action_ReOpen:
		//本端发出操作,对端返回[操作错误],本端处理这个情况.
		if !isOk {
			//参考socket,一段时间之内,还会返回错误,时间过长,就直接丢弃了,这里可以选择直接丢弃的.
		}
		panic("操作错误")
	default:
		if !isOk {
			//参考socket,一段时间之内,还会返回错误,时间过长,就直接丢弃了,这里可以选择直接丢弃的.
		}
		panic("逻辑异常")
	}
	return sess
}
