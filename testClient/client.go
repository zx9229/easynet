package main

import (
	"fmt"

	"github.com/zx9229/easynet/easynet"
)

func main() {
	tcpAddr := "localhost:54321"
	eClient := easynet.NewEasyClientImpl()
	eClient.RegEasyConnected(easynet.EgOnConnected)
	eClient.RegEasyDisConnected(easynet.EgOnDisconnected)
	eClient.RegEasyMessage(easynet.EgOnMessage)
	eClient.Connect(tcpAddr, true)
	for {
		var line string
		fmt.Scanln(&line)
		eClient.Send([]byte(line))
	}
}
