package main

import (
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"go-rustdesk-server/http_server"
	"go-rustdesk-server/relay"
	"go-rustdesk-server/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logs.SetLevel(logs.DEBUG)
	logs.SetWriteLogs(logs.DEBUG | logs.INFO | logs.ERR)
	common.LoadKey()
	go http_server.Always200Server()
	go server.Start()
	go relay.Start()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
