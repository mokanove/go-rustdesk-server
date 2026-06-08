package main

import (
	"flag"
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"go-rustdesk-server/http_server"
	"go-rustdesk-server/data_server"
	"go-rustdesk-server/relay"
	"go-rustdesk-server/server"
	"os"
	"os/signal"
	"syscall"
)
func main(){
    go http_server.Always200Server()
    go relay.Start()
    go server.Start()
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
}

func rdks() {
	_relay := flag.Bool("relay", common.Conf.RelayName != "", "run relay")
	_server := flag.Bool("server", true, "run server")
	flag.Parse()
	if *_relay {

	}
	if *_server {

	}
	if common.Conf.Debug {
		logs.SetLevel(logs.DEBUG)
		logs.SetWriteLogs(logs.INFO | logs.ERR | logs.DEBUG)
	} else {
		logs.SetLevel(logs.INFO)
		logs.SetWriteLogs(logs.INFO | logs.ERR)
	}
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		sever, err := data_server.GetDataSever()
		if err == nil {
			err = sever.Close()
			if err != nil {
				logs.Err(err)
			}
		}
		done <- true
	}()
	<-done
}
