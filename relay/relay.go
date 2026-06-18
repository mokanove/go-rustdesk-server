package relay

import (
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model/model_proto"
	"go-rustdesk-server/my_bytes"
	"google.golang.org/protobuf/proto"
	"reflect"
)

func Start() {
	cmd.Info("Relay TCP listening %s", common.PortRelay)
	common.NewMonitor(true, "tcp", "[::]"+common.PortRelay, handlerMsg).Start()
}

func handlerMsg(msg []byte, writer *common.Writer) {
	cmd.Info("RELAY-RX %s len %d", writer.GetAddrStr(), len(msg))
	message := model_proto.RendezvousMessage{}
	if err := proto.Unmarshal(msg, &message); err != nil || message.Union == nil {
		if err != nil {
			cmd.Fatal("RELAY unmarshal", err)
		}
		return
	}
	msgType := reflect.TypeOf(message.Union).String()
	cmd.Info("RELAY-RX type %s from %s", msgType, writer.GetAddrStr())
	switch msgType {
	case model_proto.TypeRendezvousMessageRequestRelay:
		rr := message.GetRequestRelay()
		if rr == nil {
			return
		}
		uuid := rr.GetUuid()
		cmd.Info("RELAY uuid %s from %s %s", uuid, writer.GetAddrStr(), my_bytes.DecodeAddr(rr.SocketAddr))
		if uuid == "" {
			return
		}
		w, err := common.GetWriter(uuid, common.TCP)
		if err != nil {
			cmd.Info("RELAY waiting for peer %s", uuid)
			writer.SetKey(uuid)
		} else {
			cmd.Info("RELAY pairing %s <-> %s", writer.GetAddrStr(), w.GetAddrStr())
			common.RemoveWriter(writer)
			common.RemoveWriter(w)
			go writer.Copy(w)
		}
	default:
		cmd.Info("RELAY unknown type %s", msgType)
	}
}