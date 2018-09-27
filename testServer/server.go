package main

import (
	"os"

	"github.com/zx9229/easynet/easynet2"
)

func main() {
	toLogSet()
	tcpAddr := "localhost:54321"
	eServer := easynet2.NewEasyServerImpl()
	eServer.RegEasyConnected(EgOnConnected)
	eServer.RegEasyDisConnected(EgOnDisconnected)
	eServer.RegEasyMessage(EgOnMessage)
	go eServer.Run(tcpAddr)

	jiaoHu()
}
func toLogSet() {
	logFile, err := os.OpenFile("./server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	easynet2.NetLog.SetOutput(logFile)
	easynet2.NetLog.INFO.Println("SERVER_BEGIN")
}
