package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/zx9229/easynet/easynet2"
)

var globalData *cacheData

func jiaoHu() {
	globalData = newCacheData()
	for {
		//socket,session,action,data
		var line string
		fmt.Scanln(&line)
		fields := strings.Split(line, ",")
		if len(fields) != 4 {
			log.Printf("fields=%v", fields)
			continue
		}
		sockNm := strings.Trim(fields[0], " \t\r\n")
		sessNm := strings.Trim(fields[1], " \t\r\n")
		action := strings.Trim(fields[2], " \t\r\n")
		mesage := strings.Trim(fields[3], " \t\r\n")
		if action == "open" {
			globalData.openSession(sockNm)
		} else if action == "close" {
			globalData.closeSession(sockNm, sessNm)
		} else {
			globalData.doAction(sockNm, sessNm, action, mesage, nil, nil)
		}
	}
}

//EgOnConnected omit
func EgOnConnected(eSock easynet2.EasySocket, isAccepted bool, eSess easynet2.EasySession, sessAccepted bool) {
	log.Printf("OnCon[v], %p|%p A=%v|%v, L=%v, R=%v", eSock, eSess, isAccepted, sessAccepted, eSock.LocalAddr(), eSock.RemoteAddr())
	sockName := fmt.Sprintf("%p", eSock)
	if eSock == nil {
		sockName = "0x0"
	}
	sessName := fmt.Sprintf("%p", eSess)
	if eSess == nil {
		sessName = "0x0"
	}
	globalData.doAction(sockName, sessName, "insert", "", eSock, eSess)
}

//EgOnDisconnected omit
func EgOnDisconnected(eSock easynet2.EasySocket, eSess easynet2.EasySession, err error, byDisconnected bool) {
	log.Printf("OnDis[x], %p|%p, err=%v", eSock, eSess, err)
	sockName := fmt.Sprintf("%p", eSock)
	if eSock == nil {
		sockName = "0x0"
	}
	sessName := fmt.Sprintf("%p", eSess)
	if eSess == nil {
		sessName = "0x0"
	}
	globalData.doAction(sockName, sessName, "delete", "", eSock, eSess)
}

//EgOnMessage omit
func EgOnMessage(eSock easynet2.EasySocket, eSess easynet2.EasySession, data []byte) {
	log.Printf("OnMsg[=], %p|%p, data=%v", eSock, eSess, string(data))
}

type sockInfo struct {
	name string
	sock easynet2.EasySocket
	M    map[string]easynet2.EasySession
}

type cacheData struct {
	sockCache map[string]*sockInfo
}

func newCacheData() *cacheData {
	curData := new(cacheData)
	curData.sockCache = make(map[string]*sockInfo)
	return curData
}

func (thls *cacheData) openSession(sockName string) easynet2.EasySession {
	var isOk bool
	var curSockInfo *sockInfo
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk {
		log.Printf("find not socket info [%v]", sockName)
		return nil
	}
	return curSockInfo.sock.CreateSession()
}

func (thls *cacheData) closeSession(sockName, sessName string) {
	isSession := (sessName != "0x0")
	var isOk bool
	var curSockInfo *sockInfo
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk {
		log.Printf("find not socket info [%v]", sockName)
		return
	}
	if !isSession {
		curSockInfo.sock.Close()
		return
	}
	var curSessData easynet2.EasySession
	if curSessData, isOk = curSockInfo.M[sessName]; !isOk {
		log.Printf("find not session info [%v]", sockName)
		return
	}
	curSessData.Close(true, true)
}

func (thls *cacheData) doAction(sockName, sessName, action, msg string, eSock easynet2.EasySocket, eSess easynet2.EasySession) {
	isSession := (sessName != "0x0")
	var isOk bool
	var curSockInfo *sockInfo
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk && isSession {
		log.Printf("find not socket info [%v]", sockName)
		return
	}
	var curSessData easynet2.EasySession
	if isSession {
		if curSessData, isOk = curSockInfo.M[sessName]; !isOk && (action != "insert") {
			log.Printf("find not session info [%v]", sessName)
			return
		}
	}
	switch action {
	case "insert":
		if isSession {
			curSockInfo.M[sessName] = eSess
		} else {
			thls.sockCache[sockName] = &sockInfo{name: sockName, sock: eSock, M: make(map[string]easynet2.EasySession)}
		}
	case "delete":
		if isSession {
			delete(curSockInfo.M, sessName)
		} else {
			delete(thls.sockCache, sockName)
		}
	case "send":
		if isSession {
			curSessData.Send([]byte(msg))
		} else {
			curSockInfo.sock.Send([]byte(msg))
		}
	}
}
