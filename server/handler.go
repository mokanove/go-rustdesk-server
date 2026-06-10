package server

import (
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model/model_proto"
	"google.golang.org/protobuf/proto"
	"reflect"
	"time"
)

type ringMsg struct {
	ID      string
	Type    string
	TimeOut uint32
	InsTime time.Time
	Val     interface{}
	Writer  *common.Writer
}

func getMsgForm(id string, types []string, timeOut uint) (*common.Writer, interface{}) {
	if timeOut == 0 {
		timeOut = 3
	}
	end := time.Now().Add(time.Second * time.Duration(timeOut))
	for time.Now().Before(end) {
		for _, t := range types {
			if val, ok := ringMsgMap.Load(id + t); ok {
				v := val.(*ringMsg)
				ringMsgMap.Delete(id + t)
				return v.Writer, v.Val
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, nil
}

func handlerMsg(msg []byte, writer *common.Writer) {
	logs.Debug("RX", writer.Type(), writer.GetAddrStr(), "len", len(msg))
	message := model_proto.RendezvousMessage{}
	if err := proto.Unmarshal(msg, &message); err != nil {
		logs.Err("unmarshal", err)
		return
	}
	if message.Union == nil {
		logs.Debug("empty Union from", writer.GetAddrStr())
		return
	}
	if blacklistDetection("", writer.GetAddr()) {
		return
	}
	msgType := reflect.TypeOf(message.Union).String()
	logs.Debug("RX type", msgType, "from", writer.Type(), writer.GetAddrStr())
	var response proto.Message
	switch msgType {
	case model_proto.TypeRendezvousMessagePunchHoleRequest:
		if req := message.GetPunchHoleRequest(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessagePunchHoleRequest(req, writer))
		}
	case model_proto.TypeRendezvousMessageRegisterPk:
		if req := message.GetRegisterPk(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessageRegisterPk(req, writer))
		}
	case model_proto.TypeRendezvousMessageRegisterPeer:
		if req := message.GetRegisterPeer(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessageRegisterPeer(req, writer))
			ConfigureUpdate(writer)
		}
	case model_proto.TypeRendezvousMessageSoftwareUpdate:
		if req := message.GetSoftwareUpdate(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessageSoftwareUpdate(req))
		}
	case model_proto.TypeRendezvousMessageTestNatRequest:
		if req := message.GetTestNatRequest(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessageTestNatRequest(req, writer))
		}
	case model_proto.TypeRendezvousMessageLocalAddr:
		if req := message.GetLocalAddr(); req != nil {
			RendezvousMessageLocalAddr(req, writer)
		}
	case model_proto.TypeRendezvousMessageRequestRelay:
		if req := message.GetRequestRelay(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessageRequestRelay(req))
		}
	case model_proto.TypeRendezvousMessageRelayResponse:
		if req := message.GetRelayResponse(); req != nil {
			RendezvousMessageRelayResponse(writer, req)
		}
	case model_proto.TypeRendezvousMessagePunchHoleSent:
		if req := message.GetPunchHoleSent(); req != nil {
			RendezvousMessagePunchHoleSent(req, writer)
		}
	case model_proto.TypeRendezvousMessageConfigureUpdate:
		if req := message.GetConfigureUpdate(); req != nil {
			RendezvousMessageConfigureUpdate(req)
		}
	case model_proto.TypeRendezvousMessageOnlineRequest:
		if req := message.GetOnlineRequest(); req != nil {
			response = model_proto.NewRendezvousMessage(RendezvousMessageOnlineRequest(req))
		}
	default:
		logs.Debug("RX unknown type", msgType)
	}
	if response != nil {
		logs.Debug("TX response to", writer.Type(), writer.GetAddrStr())
		writer.SendMsg(response)
	}
}
