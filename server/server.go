package server

import (
	"fmt"
	"go-rustdesk-server/common"
)

func Start() {
	// UDP 21116：信令/打洞核心端口
	go func() {
		udpMonitor := common.NewMonitor(false, "udp", "0.0.0.0"+common.PortSignal, handlerMsg)
		fmt.Printf("[Server] Listening UDP %s\n", common.PortSignal)
		udpMonitor.Start()
	}()

	// TCP 21115 / 21116 / 21118
	for _, addr := range []string{common.PortNAT, common.PortSignal, common.PortWS} {
		addr := addr
		go func() {
			fmt.Printf("[Server] Listening TCP %s\n", addr)
			common.NewMonitor(false, "tcp", "0.0.0.0"+addr, handlerMsg).Start()
		}()
	}
}
