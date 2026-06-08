package relay

import (
	"fmt"
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model/model_proto"
	"go-rustdesk-server/my_bytes"
	"google.golang.org/protobuf/proto"
	"reflect"
)

func Start() {
	fmt.Printf("[Relay] Listening TCP %s\n", common.PortRelay)
	common.NewMonitor(true, "tcp", "0.0.0.0"+common.PortRelay, handlerMsg).Start()
}

func handlerMsg(msg []byte, writer *common.Writer) {
	fmt.Printf("[RELAY-RX] %s  len=%d\n", writer.GetAddrStr(), len(msg))
	if blacklistDetection(writer.GetAddr()) {
		writer.Close()
		return
	}
	message := model_proto.RendezvousMessage{}
	err := proto.Unmarshal(msg, &message)
	if err != nil || message.Union == nil {
		if err != nil {
			fmt.Printf("[RELAY-RX] unmarshal error: %v\n", err)
			logs.Err(err)
		}
		return
	}
	msgType := reflect.TypeOf(message.Union).String()
	fmt.Printf("[RELAY-RX] type=%s from=%s\n", msgType, writer.GetAddrStr())
	logs.Debug(writer.Type(), writer.GetAddrStr(), msgType)

	switch msgType {
	case model_proto.TypeRendezvousMessageRequestRelay:
		rr := message.GetRequestRelay()
		if rr == nil {
			return
		}
		if common.MustKey && rr.LicenceKey != common.GetPkStr() {
			fmt.Printf("[RELAY] key mismatch from %s\n", writer.GetAddrStr())
			return
		}
		uuid := rr.GetUuid()
		fmt.Printf("[RELAY] uuid=%s from=%s\n", uuid, writer.GetAddrStr())
		logs.Debug(rr.Id, uuid, rr.RelayServer, rr.Token, rr.Secure, my_bytes.DecodeAddr(rr.SocketAddr))
		if uuid == "" {
			return
		}
		w, err1 := common.GetWriter(uuid, common.TCP)
		if err1 != nil {
			fmt.Printf("[RELAY] waiting for peer %s\n", uuid)
			writer.SetKey(uuid)
		} else {
			fmt.Printf("[RELAY] pairing %s ↔ %s\n", writer.GetAddrStr(), w.GetAddrStr())
			common.RemoveWriter(writer)
			common.RemoveWriter(w)
			go writer.Copy(w)
		}
	default:
		fmt.Printf("[RELAY] unknown type: %s\n", msgType)
		logs.Debug(writer.GetAddrStr(), msgType)
	}
}

func blacklistDetection(addr *common.Addr) bool {
	return common.InList(addr.GetIP())
}
