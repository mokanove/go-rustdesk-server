package server

import (
	"fmt"
	"go-rustdesk-server/common"
)

func Start() {
	// 1. 启动核心的 21116 UDP 信号打洞端口，直接对接原作者正牌的 handlerMsg 函数
	udpMonitor := common.NewMonitor(false, "udp", "0.0.0.0:21116", handlerMsg)
	go udpMonitor.Start()
	fmt.Printf("[Server] Listen UDP 21116 Successfully!\n")

	// 2. 循环启动 21115, 21116, 21118 TCP 核心服务端口，同样对接 handlerMsg
	tcpPorts := []string{":21115", ":21116", ":21118"}
	for _, addr := range tcpPorts {
		go func(listenAddr string) {
			// 这里原作者底层在 monitor 里可能对网络类型做了区分，直接传入正牌的 handlerMsg 即可
			tcpMonitor := common.NewMonitor(false, "tcp", "0.0.0.0"+listenAddr, handlerMsg)
			tcpMonitor.Start()
		}(addr)
		fmt.Printf("[Server] Listen " + addr + " Successfully\n")
	}
}