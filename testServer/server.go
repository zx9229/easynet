package main

import "github.com/zx9229/EasyTest/easynet"

func main() {
	tcpAddr := "localhost:54321"
	eServer := easynet.NewEasyServer()
	eServer.RegCbConnected(easynet.EgOnConnected)
	eServer.RegCbDisConnected(easynet.EgOnDisconnected)
	eServer.RegCbMessage(easynet.EgOnMessage)
	eServer.Run(tcpAddr)
}
