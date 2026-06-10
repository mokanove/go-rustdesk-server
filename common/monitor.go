package common

import (
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
	call    func([]byte, *Writer)
	relay   bool
}

func NewMonitor(relay bool, network, addr string, call func([]byte, *Writer)) *monitor {
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
			logs.Err("UDP listen", m.addr, err)
			return
		}
		logs.Info("UDP listening", m.addr)
		m.readUdp()
	} else {
		m.listen, err = net.Listen(m.network, m.addr)
		if err != nil {
			logs.Err("TCP listen", m.addr, err)
			return
		}
		logs.Info("TCP listening", m.addr)
		for {
			conn, err2 := m.listen.Accept()
			if err2 != nil {
				logs.Err(err2)
			} else {
				logs.Debug("CONNECT TCP", conn.RemoteAddr())
				go m.accept(conn)
			}
		}
	}
}

func (m *monitor) accept(conn net.Conn) {
	writer := &Writer{_type: tcp, tConn: conn, loop: true}
	addWriter(conn.RemoteAddr().String(), tcp, writer)
	defer func() {
		logs.Debug("DISCONNECT TCP", conn.RemoteAddr())
		if writer != nil && writer.loop {
			writer.Close()
		}
	}()
	buf := buffer.NewPool().Get()
	var realLength uint
	temp := make([]byte, 4096)
	for writer.loop {
		n, err := conn.Read(temp)
		if err != nil {
			logs.Debug("TCP read", conn.RemoteAddr(), err)
			return
		}
		if n == 0 {
			continue
		}
		_, _ = buf.Write(temp[:n])
		if realLength == 0 {
			_, realLength, err = my_bytes.DecodeHead(buf.Bytes())
			if err != nil {
				logs.Err("DecodeHead", conn.RemoteAddr(), err)
				return
			}
			logs.Debug("TCP-RX", conn.RemoteAddr(), "bodyLen", int(realLength)-int((buf.Bytes()[0]&0x3)+1))
		}
		if buf.Len() >= int(realLength) {
			cp := make([]byte, realLength)
			copy(cp, buf.Bytes())
			if buf.Len() != int(realLength) {
				rem := make([]byte, buf.Len()-int(realLength))
				copy(rem, buf.Bytes()[realLength:])
				buf.Reset()
				_, _ = buf.Write(rem)
			} else {
				buf.Reset()
			}
			realLength = 0
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
		n, addr, err := m.conn.ReadFromUDP(temp)
		if err != nil {
			logs.Err("UDP read", err)
			return
		}
		if n == 0 {
			continue
		}
		logs.Debug("UDP-RX", addr, "len", n)
		payload := make([]byte, n)
		copy(payload, temp[:n])
		writer, err := GetWriter(addr.String(), udp)
		if err != nil {
			writer = &Writer{_type: "udp", uConn: m.conn, addr: addr}
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
		logs.Err("Decode", conn.RemoteAddr(), err)
		return
	}
	writer, err := GetWriter(conn.RemoteAddr().String(), tcp)
	if err != nil {
		writer = &Writer{_type: "tcp", tConn: conn}
		addWriter(conn.RemoteAddr().String(), tcp, writer)
	}
	m.call(payload, writer)
}
