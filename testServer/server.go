package main

import "github.com/zx9229/easynet/easynet"

func main() {
	tcpAddr := "localhost:54321"
	eServer := easynet.NewEasyServer3()
	eServer.RegEasyConnected(easynet.EgOnConnected)
	eServer.RegEasyDisConnected(easynet.EgOnDisconnected)
	eServer.RegEasyMessage(easynet.EgOnMessage)
	eServer.Run(tcpAddr)
}
