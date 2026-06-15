package main

import (
	"fmt"
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/common"
	"go-rustdesk-server/server"
	"go-rustdesk-server/relay"
	"go-rustdesk-server/http_server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			cmd.PrintVersion()
		case "doctor":
			cmd.Doctor()
		case "help":
			cmd.PrintHelp()
		default:
			fmt.Printf("Unknown Command\n")
			fmt.Printf("Using: ./go-rustdesk-server help for usage.\n")
			os.Exit(0)
		}
	}
	cmd.Log()
	http_server.Always200Server()
	common.LoadKey()
	go server.Start()
	relay.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
