package my_bytes

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// ---- TCP 帧格式 ----
// RustDesk hbbs 使用固定 4 字节大端长度前缀
// | 4 bytes big-endian length | payload... |

// Encoder 将 payload 打包成带 4 字节长度头的帧。
func Encoder(data []byte) ([]byte, error) {
	totalLen := 4 + len(data)
	buf := make([]byte, totalLen)
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	return buf, nil
}

// DecodeHead 解析长度头，返回 (头部字节数, 完整帧总长度, error)。
// 调用方用 totalLen 判断缓冲区是否已收齐完整帧。
func DecodeHead(buf []byte) (uint, uint, error) {
	if len(buf) < 4 {
		return 0, 0, errors.New("incomplete header")
	}
	bodyLen := binary.BigEndian.Uint32(buf[:4])
	return 4, uint(4) + uint(bodyLen), nil
}

// Decode 去掉 4 字节长度头，返回纯 payload。
func Decode(frame []byte) ([]byte, error) {
	if len(frame) < 4 {
		return nil, errors.New("frame too short")
	}
	bodyLen := binary.BigEndian.Uint32(frame[:4])
	if uint32(len(frame)) < 4+bodyLen {
		return nil, errors.New("frame truncated")
	}
	payload := make([]byte, bodyLen)
	copy(payload, frame[4:4+bodyLen])
	return payload, nil
}

// ---- SocketAddr 编码 ----
// IPv4: [0x01, b0,b1,b2,b3, portHi, portLo]          7 字节
// IPv6: [0x02, b0..b15,     portHi, portLo]           19 字节

// EncodeAddr 将 "host:port" 字符串编码为 RustDesk SocketAddr 字节序列。
func EncodeAddr(addr string) []byte {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}
	var port uint16
	fmt.Sscanf(portStr, "%d", &port)

	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}

	if ip4 := ip.To4(); ip4 != nil {
		buf := make([]byte, 7)
		buf[0] = 0x01
		copy(buf[1:5], ip4)
		binary.BigEndian.PutUint16(buf[5:7], port)
		return buf
	}

	// IPv6
	ip6 := ip.To16()
	buf := make([]byte, 19)
	buf[0] = 0x02
	copy(buf[1:17], ip6)
	binary.BigEndian.PutUint16(buf[17:19], port)
	return buf
}

// DecodeAddr 将 RustDesk SocketAddr 字节序列解码为 "host:port" 字符串。
func DecodeAddr(data []byte) string {
	if len(data) == 7 && data[0] == 0x01 {
		ip := net.IP(data[1:5])
		port := binary.BigEndian.Uint16(data[5:7])
		return fmt.Sprintf("%s:%d", ip.String(), port)
	}
	if len(data) == 19 && data[0] == 0x02 {
		ip := net.IP(data[1:17])
		port := binary.BigEndian.Uint16(data[17:19])
		return fmt.Sprintf("[%s]:%d", ip.String(), port)
	}
	return ""
}
