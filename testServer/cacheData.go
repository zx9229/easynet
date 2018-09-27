package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
			log.Printf("unknown_command=[%v]", line)
			continue
		}
		sockNm := strings.Trim(fields[0], " \t\r\n")
		sessNm := strings.Trim(fields[1], " \t\r\n")
		action := strings.Trim(fields[2], " \t\r\n")
		mesage := strings.Trim(fields[3], " \t\r\n")
		switch action {
		case "open":
			globalData.openSession(sockNm)
		case "close":
			globalData.closeSessionOrSocket(sockNm, sessNm)
		case "send":
			globalData.sendData(sockNm, sessNm, action, mesage)
		default:
			log.Printf("unknown_command=[%v]", line)
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
	globalData.insertData(sockName, sessName, eSock, eSess)
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
	globalData.deleteData(sockName, sessName, eSock, eSess)
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
		log.Printf("cache does not exist socket = [%v]", sockName)
		return nil
	}
	return curSockInfo.sock.CreateSession() //让回调函数更新cache
}

func (thls *cacheData) closeSessionOrSocket(sockName, sessName string) {
	isSession := (sessName != "0x0")
	var isOk bool
	var curSockInfo *sockInfo
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk {
		log.Printf("cache does not exist socket = [%v]", sockName)
		return
	}
	if !isSession {
		curSockInfo.sock.Close() //让回调函数更新cache
		return
	}
	var curSessData easynet2.EasySession
	if curSessData, isOk = curSockInfo.M[sessName]; !isOk {
		log.Printf("cache does not exist session = [%v]", sessName)
		return
	}
	curSessData.Close(true, true) //让回调函数更新cache
}

func (thls *cacheData) insertData(sockName, sessName string, eSock easynet2.EasySocket, eSess easynet2.EasySession) {
	isSession := (sessName != "0x0")
	var isOk bool
	var curSockInfo *sockInfo
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk { //找不到socket
		if isSession { //更新socket的缓存
			message := fmt.Sprintf("insert session (%v|%v) and no   socket = %p", sockName, sessName, curSockInfo.sock)
			log.Println(message)
			panic(message)
		} else {
			thls.sockCache[sockName] = &sockInfo{name: sockName, sock: eSock, M: make(map[string]easynet2.EasySession)}
		}
	} else { //找到了socket
		if !isSession { //插入socket
			message := fmt.Sprintf("insert socket (%v|%v) and exist socket = %p", sockName, sessName, curSockInfo.sock)
			log.Println(message)
			panic(message)
		} else {
			curSockInfo.M[sessName] = eSess
		}
	}
}

func (thls *cacheData) deleteData(sockName, sessName string, eSock easynet2.EasySocket, eSess easynet2.EasySession) {
	isSession := (sessName != "0x0")
	var isOk bool
	var curSockInfo *sockInfo
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk { //找不到socket
		message := fmt.Sprintf("delete (%v|%v) and no socket = %p", sockName, sessName, curSockInfo.sock)
		log.Println(message)
		panic(message)
	}
	if isSession {
		if _, isOk = curSockInfo.M[sessName]; !isOk {
			message := fmt.Sprintf("delete (%v|%v) and no session = %v", sockName, sessName, sessName)
			log.Println(message)
			panic(message)
		}
		delete(curSockInfo.M, sessName)
	} else {
		delete(thls.sockCache, sockName)
	}
}

func (thls *cacheData) sendData(sockName, sessName, action, msg string) {
	isSession := (sessName != "0x0")
	var isOk bool
	var curSockInfo *sockInfo
	var err error
	if curSockInfo, isOk = thls.sockCache[sockName]; !isOk { //找不到socket
		log.Println("send to (%v|%v) and no socket = %v", sockName, sessName, sockName)
		return
	}
	if isSession {
		var curSession easynet2.EasySession
		if curSession, isOk = curSockInfo.M[sessName]; !isOk {
			log.Println("send to (%v|%v) and no session = %v", sockName, sessName, sessName)
			return
		}
		if err = curSession.Send([]byte(msg)); err != nil {
			log.Println("send to (%v|%v) and err = %v", sockName, sessName, err)
		}
	} else {
		if err = curSockInfo.sock.Send([]byte(msg)); err != nil {
			log.Println("send to (%v|%v) and err = %v", sockName, sessName, err)
		}
	}
}

func initLog(isClient bool) {
	logFilename := time.Now().Format("20060102_150405")
	if isClient {
		logFilename = "log_client_" + logFilename + ".log"
	} else {
		logFilename = "log_server_" + logFilename + ".log"
	}
	logFile, err := os.OpenFile(logFilename, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	easynet2.NetLog.SetOutput(logFile)
	easynet2.NetLog.INFO.Println("init log finish.")
}
