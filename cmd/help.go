package cmd

import (
	"fmt"
)

func PrintHelp() {
	fmt.Printf("Usage: \n")
	fmt.Printf("  go-rustdesk-server [command]\n\n")
	fmt.Printf("Available Commands:\n")
	fmt.Printf("  no any:Run Rustdesk Server and Relay\n")
	fmt.Printf("  version:Print version of go-rustdesk-server\n")
	fmt.Printf("  doctor [IP/Domain]:Test rustdesk-server\n")
	fmt.Printf("  any:Print help information\n")
}
