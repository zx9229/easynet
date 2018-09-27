package main

import (
	"github.com/zx9229/easynet/easynet2"
)

func main() {
	initLog(true)
	tcpAddr := "localhost:54321"
	eClient := easynet2.NewEasyClientImpl()
	eClient.RegEasyConnected(EgOnConnected)
	eClient.RegEasyDisConnected(EgOnDisconnected)
	eClient.RegEasyMessage(EgOnMessage)
	eClient.Connect(tcpAddr, true)
	jiaoHu()
}
