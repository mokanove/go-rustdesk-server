package server

import (
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/common"
)

func Start() {
	go func() {
		cmd.Info("UDP listening %s", common.PortSignal)
		common.NewMonitor(false, "udp", "0.0.0.0"+common.PortSignal, handlerMsg).Start()
	}()
	for _, addr := range []string{common.PortNAT, common.PortSignal, common.PortWS} {
		addr := addr
		go func() {
			cmd.Info("TCP listening %s", addr)
			common.NewMonitor(false, "tcp", "0.0.0.0"+addr, handlerMsg).Start()
		}()
	}
}
