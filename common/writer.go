package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/model/model_msg"
	"go-rustdesk-server/my_bytes"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"time"
)

var ctx = gctx.New()
var cache = gcache.New()
var cacheTimeOut = time.Second * 60

type Writer struct {
	key   string
	_type string
	tConn net.Conn
	uConn *net.UDPConn
	addr  *net.UDPAddr
	loop  bool
}

type Addr struct {
	ip   string
	port uint32
}

func (a *Addr) GetIP() string   { return a.ip }
func (a *Addr) GetPort() uint32 { return a.port }
func (a *Addr) Parsing(addr string) {
	ip, p := GetHostPort(addr)
	a.ip = ip
	a.port = uint32(p)
}

func (w *Writer) Type() string { return w._type }

func (w *Writer) Write(p []byte) (int, error) {
	switch w._type {
	case udp:
		if w.uConn == nil {
			return 0, errors.New("uConn==nil")
		}
		return w.uConn.WriteToUDP(p, w.addr)
	case tcp:
		if w.tConn == nil {
			return 0, errors.New("tConn==nil")
		}
		enc, err := my_bytes.Encoder(p)
		if err != nil {
			return 0, err
		}
		return w.tConn.Write(enc)
	}
	return 0, errors.New("type Err")
}

func (w *Writer) WriteToAddr(p []byte, addr string) (int, error) {
	if w._type != udp {
		return 0, errors.New("unrealized")
	}
	if w.uConn == nil {
		return 0, errors.New("uConn==nil")
	}
	udpAddr, err := net.ResolveUDPAddr(udp, addr)
	if err != nil {
		return 0, err
	}
	return w.uConn.WriteToUDP(p, udpAddr)
}

func (w *Writer) GetAddrStr() string {
	switch w._type {
	case udp:
		return w.addr.String()
	case tcp:
		return w.tConn.RemoteAddr().String()
	}
	return ""
}

func (w *Writer) GetAddr() *Addr {
	a := &Addr{}
	switch w._type {
	case udp:
		a.Parsing(w.addr.String())
	case tcp:
		a.Parsing(w.tConn.RemoteAddr().String())
	}
	return a
}

func (w *Writer) SetKey(key string) {
	_ = cache.Set(ctx, w._type+key, w, cacheTimeOut)
	w.key = key
}

func (w *Writer) remove() {
	switch w._type {
	case udp:
		_, _ = cache.Remove(ctx, udp+w.addr.String())
	case tcp:
		_, _ = cache.Remove(ctx, tcp+w.tConn.RemoteAddr().String())
	}
	if w.key != "" {
		_, _ = cache.Remove(ctx, udp+w.key)
		_, _ = cache.Remove(ctx, tcp+w.key)
	}
}

func (w *Writer) Copy(dst *Writer) {
	if w._type != tcp || dst == nil || dst.tConn == nil {
		return
	}
	defer func() {
		dst.Close()
		w.Close()
	}()
	go io.Copy(dst.tConn, w.tConn)
	io.Copy(w.tConn, dst.tConn)
}

func (w *Writer) SendMsg(message proto.Message) {
	if message == nil {
		return
	}
	data, err := proto.Marshal(message)
	if err != nil {
		cmd.Info("%v", err)
		return
	}
	if _, err = w.Write(data); err != nil {
		cmd.Info("%v", err)
	}
}

func (w *Writer) SendJsonMsg(message *model_msg.Msg) {
	if message == nil {
		return
	}
	data, err := json.Marshal(message)
	if err != nil {
		cmd.Info("%v", err)
		return
	}
	if _, err = w.Write(data); err != nil {
		cmd.Info("%v", err)
	}
}

func (w *Writer) Close() {
	if w._type == tcp {
		_ = w.tConn.Close()
	}
	w.remove()
}

func (w *Writer) SelfAddr() string {
	switch w._type {
	case udp:
		ip, _ := GetHostPort(w.uConn.LocalAddr().String())
		return ip
	case tcp:
		ip, _ := GetHostPort(w.tConn.RemoteAddr().String())
		return ip
	}
	return ""
}

func GetWriter(key, _type string) (*Writer, error) {
	get, _ := cache.Get(ctx, fmt.Sprint(_type, key))
	if get != nil {
		if v, ok := get.Val().(*Writer); ok {
			return v, nil
		}
	}
	return nil, errors.New("OFFLINE")
}

func addWriter(key, _type string, w *Writer) {
	t := cacheTimeOut
	if _type == tcp {
		t = 0
	}
	if err := cache.Set(ctx, fmt.Sprint(_type, key), w, t); err != nil {
		cmd.Info("%v", err)
	}
}

func RemoveWriter(w *Writer) { w.remove() }
