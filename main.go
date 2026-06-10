package main

import (
	"os"
	"fmt"
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/http_server"
)

func main() {
    if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			cmd.PrintVersion()
			os.Exit(0)
		case "help":
			cmd.PrintHelp()
			os.Exit(0)
		default:
			fmt.Printf("Unknow Command\n")
			fmt.Printf("Using: ./go-rustdesk-server help for usage.\n")
			os.Exit(0)
		}
	}
    http_server.Always200Server()
}