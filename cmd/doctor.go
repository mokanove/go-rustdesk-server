package cmd

import (
	"fmt"
	"net"
	"os"
	"time"
)

const dialTimeout = 3 * time.Second

type portCheck struct {
	port    string
	network string
	desc    string
}

var rustdeskPorts = []portCheck{
	{"21114", "tcp", "WebUI and API_Server"},
	{"21115", "tcp", "NAT Checker"},
	{"21116", "tcp", "Tunnel and Connect"},
	{"21116", "udp", "ID sign_in and heartbeat"},
	{"21117", "tcp", "Relay and file transfer"},
	{"21118", "tcp", "WebConnect"},
	{"21119", "tcp", "WebConnect"},
}

func Doctor() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go-rustdesk-server doctor <IP or domain>")
		os.Exit(1)
	}

	host := os.Args[2]
	fmt.Printf("[Doctor] Checking %s\n\n", host)
	allOK := true
	for _, c := range rustdeskPorts {
		addr := net.JoinHostPort(host, c.port)
		if isReachable(c.network, addr) {
			fmt.Printf("  OK   %-30s %s\n", c.desc, addr)
		} else {
			fmt.Printf("  FAIL %-30s %s\n", c.desc, addr)
			allOK = false
		}
	}

	fmt.Println()
	if allOK {
		fmt.Println("[Doctor] All ports OK")
		os.Exit(0)
	} else {
		fmt.Println("[Doctor] Some ports are unreachable. Check your firewall and that all services are running.")
		os.Exit(1)
	}
}

func isReachable(network, addr string) bool {
	conn, err := net.DialTimeout(network, addr, dialTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
