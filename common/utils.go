package common

import (
	"net"
	"os"
	"strconv"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func InSameSubnet(ip1Str, ip2Str, maskStr string) bool {
	ip1 := net.ParseIP(ip1Str)
	ip2 := net.ParseIP(ip2Str)
	mask := net.IPMask(net.ParseIP(maskStr).To4())
	if ip1 == nil || ip2 == nil {
		return false
	}
	return ip1.Mask(mask).Equal(ip2.Mask(mask))
}

func InSubnet(ip string) bool {
	if !(strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.")) {
		return false
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				localIP := ipNet.IP.String()
				if strings.HasPrefix(localIP, "192.168.") || strings.HasPrefix(localIP, "10.") || strings.HasPrefix(localIP, "172.") {
					if InSameSubnet(localIP, ip, net.IP(ipNet.Mask).String()) {
						return true
					}
				}
			}
		}
	}
	return false
}

func GetHostPort(addr string) (string, uint64) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0
	}
	p, _ := strconv.ParseUint(portStr, 10, 32)
	return host, p
}

func OutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	host, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	return host
}
