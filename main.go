package main

import (
	logs "github.com/danbai225/go-logs"
	"os"
	"os/signal"
	"syscall"
	"go-rustdesk-server/http_server"
	"go-rustdesk-server/relay"
	"go-rustdesk-server/server"
	
)
func main(){
	logs.SetLevel(logs.INFO)
	logs.SetWriteLogs(logs.INFO | logs.ERR)
	go http_server.Always200Server()
	go server.Start()
	go relay.Start()
	sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
}
