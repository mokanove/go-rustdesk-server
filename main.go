package main

import (
	"fmt"
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/common"
	"go-rustdesk-server/http_server"
	"go-rustdesk-server/relay"
	"go-rustdesk-server/server"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			cmd.PrintVersion()
			os.Exit(0)
		case "doctor":
			cmd.Doctor()
			os.Exit(0)
		case "help":
			cmd.PrintHelp()
			os.Exit(0)
		default:
			fmt.Printf("Unknown Command\n")
			fmt.Printf("Using: ./go-rustdesk-server help for usage.\n")
			os.Exit(0)
		}
	}

	cmd.Log()
	go http_server.Always200Server()
	common.LoadKey()
	go server.Start()
	go relay.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
