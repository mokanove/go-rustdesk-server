package common

import (
	"fmt"
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/my_bytes"
	"go.uber.org/zap/buffer"
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

	// 使用 buffer 池积累数据，处理 TCP 粘包/半包
	buf := buffer.NewPool().Get()
	realLength := uint(0)
	temp := make([]byte, 4096)

	for writer.loop {
		readLen, err := conn.Read(temp)
		if err != nil {
			fmt.Printf("[TCP] %s body read error: %v\n", conn.RemoteAddr(), err)
			return
		}
		if readLen == 0 {
			continue
		}
		_, _ = buf.Write(temp[:readLen])

		// 尝试解析帧头（只在还不知道帧长度时）
		if realLength == 0 {
			_, realLength, err = my_bytes.DecodeHead(buf.Bytes())
			if err != nil {
				fmt.Printf("[TCP] %s DecodeHead error: %v  header=%x\n",
					conn.RemoteAddr(), err, buf.Bytes())
				return
			}
			bodyLen := int(realLength) - int((buf.Bytes()[0]&0x3)+1)
			fmt.Printf("[TCP-RX] %s  bodyLen=%d\n", conn.RemoteAddr(), bodyLen)
		}

		// 凑够一个完整帧才处理
		if buf.Len() >= int(realLength) {
			cp := make([]byte, realLength)
			copy(cp, buf.Bytes())
			// 如果缓冲区里还有下一帧的数据，保留它
			if buf.Len() != int(realLength) {
				remaining := make([]byte, buf.Len()-int(realLength))
				copy(remaining, buf.Bytes()[realLength:])
				buf.Reset()
				_, _ = buf.Write(remaining)
			} else {
				buf.Reset()
			}
			realLength = 0 // 重置，等待下一帧头

			if m.relay && writer != nil {
				writer.loop = false
			}
			go m.processMessageData(cp, conn)
		}
	}
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
		fmt.Printf("[FRAMING] Decode error from %s: %v\n", conn.RemoteAddr(), err)
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
