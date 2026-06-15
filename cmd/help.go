package cmd

import (
	"fmt"
	"os"
)

func PrintHelp() {
	fmt.Println("Usage:")
	fmt.Println("  go-rustdesk-server [command]\n")
	fmt.Println("Commands:")
	fmt.Println("  (none)              Start the RustDesk relay server")
	fmt.Println("  version             Print version and build info")
	fmt.Println("  doctor <IP/domain>  Check all RustDesk port connectivity on a host")
	fmt.Println("  help/any            Show this help message")
	os.Exit(0)
}
