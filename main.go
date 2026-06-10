package main

import (
	"os"
	"fmt"
	"go-rustdesk-server/cmd"
)

func main() {
    if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			cmd.PrintVersion()
			return
		case "help":
			cmd.PrintVersion()
			return
		default:
			fmt.Printf("Unknow Command.\n")
			fmt.Printf("Using: ./go-rustdesk-server help for usage.\n")
			os.Exit(1)
		}
	}
}