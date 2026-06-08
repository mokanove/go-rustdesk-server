package common

import (
	"fmt"
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/my_bytes"
	"net"
)

type monitor struct {
	network string
	addr    string
	listen  net.Listener
	conn    *net.UDPConn
	call    func(msg []byte, writer *Writer)
	relay   bool
}

func NewMonitor(relay bool, network, addr string, call func(msg []byte, writer *Writer)) *monitor {
	return &monitor{network: network, addr: addr, call: call, relay: relay}
}

func (m *monitor) Start() {
	defer func() {
		if m.listen != nil {
			_ = m.listen.Close()
		}
		if m.conn != nil {
			_ = m.conn.Close()
		}
	}()
	var err error
	if m.network == udp {
		addr, err1 := net.ResolveUDPAddr(m.network, m.addr)
		if err1 != nil {
			logs.Err(err1)
			return
		}
		m.conn, err = net.ListenUDP(m.network, addr)
		if err != nil {
			fmt.Printf("[MONITOR] UDP listen %s failed: %v\n", m.addr, err)
			logs.Err(err)
			return
		}
		fmt.Printf("[MONITOR] UDP listening on %s\n", m.addr)
		m.readUdp()
	} else {
		m.listen, err = net.Listen(m.network, m.addr)
		if err != nil {
			fmt.Printf("[MONITOR] TCP listen %s failed: %v\n", m.addr, err)
			logs.Err(err)
			return
		}
		fmt.Printf("[MONITOR] TCP listening on %s\n", m.addr)
		for {
			conn, err2 := m.listen.Accept()
			if err2 != nil {
				logs.Err(err2)
			} else {
				fmt.Printf("[CONNECT] TCP %s → %s\n", conn.RemoteAddr(), m.addr)
				go m.accept(conn)
			}
		}
	}
}

func (m *monitor) accept(conn net.Conn) {
	writer := &Writer{
		_type: tcp,
		tConn: conn,
		loop:  true,
	}
	addWriter(conn.RemoteAddr().String(), tcp, writer)
	defer func() {
		fmt.Printf("[DISCONNECT] TCP %s\n", conn.RemoteAddr())
		if writer != nil && writer.loop {
			writer.Close()
		}
	}()

	// 整块读取，直到凑够一个完整帧
	header := make([]byte, 4)
	for writer.loop {
		// 1. 读取 4 字节长度头
		if _, err := readFull(conn, header); err != nil {
			fmt.Printf("[TCP] %s header read error: %v\n", conn.RemoteAddr(), err)
			return
		}
		_, totalLen, err := my_bytes.DecodeHead(header)
		if err != nil {
			fmt.Printf("[TCP] %s DecodeHead error: %v  header=%x\n", conn.RemoteAddr(), err, header)
			return
		}
		bodyLen := totalLen - 4
		fmt.Printf("[TCP-RX] %s  bodyLen=%d\n", conn.RemoteAddr(), bodyLen)

		// 2. 读取 body
		body := make([]byte, bodyLen)
		if _, err := readFull(conn, body); err != nil {
			fmt.Printf("[TCP] %s body read error: %v\n", conn.RemoteAddr(), err)
			return
		}

		// 3. 拼成完整帧交给 processMessageData（它会再 Decode 去掉头）
		frame := make([]byte, 4+bodyLen)
		copy(frame[:4], header)
		copy(frame[4:], body)

		if m.relay && writer != nil {
			writer.loop = false
		}
		go m.processMessageData(frame, conn)
	}
}

// readFull 从 conn 中精确读取 len(buf) 个字节，处理短读。
func readFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (m *monitor) readUdp() {
	temp := make([]byte, 65535)
	for {
		readLen, addr, err := m.conn.ReadFromUDP(temp)
		if err != nil {
			fmt.Printf("[MONITOR] UDP read error: %v\n", err)
			return
		}
		if readLen == 0 {
			continue
		}
		fmt.Printf("[UDP-RX] %s  len=%d\n", addr, readLen)
		payload := make([]byte, readLen)
		copy(payload, temp[:readLen])
		writer, err := GetWriter(addr.String(), udp)
		if err != nil {
			writer = &Writer{
				_type: "udp",
				uConn: m.conn,
				addr:  addr,
			}
			addWriter(addr.String(), udp, writer)
		}
		m.call(payload, writer)
	}
}

func (m *monitor) processMessageData(data []byte, conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			logs.Err(r)
		}
	}()
	payload, err := my_bytes.Decode(data)
	if err != nil {
		fmt.Printf("[FRAMING] Decode error from %s: %v  raw=%x\n", conn.RemoteAddr(), err, data[:min(16, len(data))])
		logs.Err(err)
		return
	}
	writer, err := GetWriter(conn.RemoteAddr().String(), tcp)
	if err != nil {
		writer = &Writer{
			_type: "tcp",
			tConn: conn,
		}
		addWriter(conn.RemoteAddr().String(), tcp, writer)
	}
	m.call(payload, writer)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
