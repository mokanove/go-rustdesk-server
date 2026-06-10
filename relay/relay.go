package relay

import (
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model/model_proto"
	"go-rustdesk-server/my_bytes"
	"google.golang.org/protobuf/proto"
	"reflect"
)

func Start() {
	logs.Info("Relay TCP listening", common.PortRelay)
	common.NewMonitor(true, "tcp", "0.0.0.0"+common.PortRelay, handlerMsg).Start()
}

func handlerMsg(msg []byte, writer *common.Writer) {
	logs.Debug("RELAY-RX", writer.GetAddrStr(), "len", len(msg))
	if blacklistDetection(writer.GetAddr()) {
		writer.Close()
		return
	}
	message := model_proto.RendezvousMessage{}
	if err := proto.Unmarshal(msg, &message); err != nil || message.Union == nil {
		if err != nil {
			logs.Err("RELAY unmarshal", err)
		}
		return
	}
	msgType := reflect.TypeOf(message.Union).String()
	logs.Debug("RELAY-RX type", msgType, "from", writer.GetAddrStr())
	switch msgType {
	case model_proto.TypeRendezvousMessageRequestRelay:
		rr := message.GetRequestRelay()
		if rr == nil {
			return
		}
		if common.MustKey && rr.LicenceKey != common.GetPkStr() {
			logs.Debug("RELAY key mismatch from", writer.GetAddrStr())
			return
		}
		uuid := rr.GetUuid()
		logs.Debug("RELAY uuid", uuid, "from", writer.GetAddrStr(), my_bytes.DecodeAddr(rr.SocketAddr))
		if uuid == "" {
			return
		}
		w, err := common.GetWriter(uuid, common.TCP)
		if err != nil {
			logs.Debug("RELAY waiting for peer", uuid)
			writer.SetKey(uuid)
		} else {
			logs.Debug("RELAY pairing", writer.GetAddrStr(), "<->", w.GetAddrStr())
			common.RemoveWriter(writer)
			common.RemoveWriter(w)
			go writer.Copy(w)
		}
	default:
		logs.Debug("RELAY unknown type", msgType)
	}
}

func blacklistDetection(addr *common.Addr) bool {
	return common.InList(addr.GetIP())
}
