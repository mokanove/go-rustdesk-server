package cmd

import (
	"fmt"
	"net"
	"os"
	"time"
)

type portCheck struct {
	port    string
	network string
	desc    string
}

var checks = []portCheck{
	{"21114", "tcp", "Always 200OK Server"},
	{"21115", "tcp", "Server TCP"},
	{"21116", "tcp", "NAT Checker"},
	{"21116", "udp", "Server UDP"},
	{"21117", "tcp", "Relay TCP"},
	{"21118", "tcp", "Server WebSocket"},
}

func Doctor() {
	if len(os.Args) < 3 {
		fmt.Printf("Can't get IP/Domain, Try using go-rustdesk-server help\n")
		os.Exit(0)
	}
	host := os.Args[2]
	fmt.Printf("[Doctor]: Checking %s\n", host)
	allOK := true
	for _, c := range checks {
		addr := host + ":" + c.port
		ok := checkConn(c.network, addr)
		if ok {
			fmt.Printf("[Doctor]: OK: %s %s\n", c.desc, addr)
		} else {
			fmt.Printf("[Doctor]: FAIL: %s %s\n", c.desc, addr)
			allOK = false
		}
	}
	if allOK {
		fmt.Println("[Doctor]: All OK \n")
	} else {
		fmt.Println("[Doctor]: Have problem \n")
		os.Exit(0)
	}
}

func checkConn(network, addr string) bool {
	conn, err := net.DialTimeout(network, addr, time.Second*3)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
