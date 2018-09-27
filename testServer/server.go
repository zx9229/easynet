package main

import (
	"github.com/zx9229/easynet/easynet2"
)

func main() {
	initLog(false)
	tcpAddr := "localhost:54321"
	eServer := easynet2.NewEasyServerImpl()
	eServer.RegEasyConnected(EgOnConnected)
	eServer.RegEasyDisConnected(EgOnDisconnected)
	eServer.RegEasyMessage(EgOnMessage)
	go eServer.Run(tcpAddr)
	jiaoHu()
}
