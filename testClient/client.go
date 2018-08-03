package main

import (
	"fmt"

	"github.com/zx9229/EasyTest/easynet"
)

func main() {
	tcpAddr := "localhost:54321"
	eClient := easynet.NewEasyClient()
	eClient.RegCbConnected(easynet.EgOnConnected)
	eClient.RegCbDisConnected(easynet.EgOnDisconnected)
	eClient.RegCbMessage(easynet.EgOnMessage)
	eClient.Connect(tcpAddr, true)
	for {
		var line string
		fmt.Scanln(&line)
		eClient.Send([]byte(line))
	}
}
