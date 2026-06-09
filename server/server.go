package server

import (
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
)

func Start() {
	go func() {
		logs.Info("UDP listening", common.PortSignal)
		common.NewMonitor(false, "udp", "0.0.0.0"+common.PortSignal, handlerMsg).Start()
	}()
	for _, addr := range []string{common.PortNAT, common.PortSignal, common.PortWS} {
		addr := addr
		go func() {
			logs.Info("TCP listening", addr)
			common.NewMonitor(false, "tcp", "0.0.0.0"+addr, handlerMsg).Start()
		}()
	}
}
