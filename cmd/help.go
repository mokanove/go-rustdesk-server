package cmd

import (
	"fmt"
)

func PrintHelp() {
	fmt.Printf("Usage: \n")
	fmt.Printf("  go-rustdesk-server [command]\n\n")
	fmt.Printf("Available Commands:\n")
	fmt.Printf("  version:Print version of go-rustdesk-server\n")
	fmt.Printf("  doctor:Test rustdesk-server\n")
	fmt.Printf("  any:Print help information\n")
}