package server

import (
	"fmt"
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

func getMsgForm(id string, Type []string, timeOut uint) (*common.Writer, interface{}) {
	if timeOut == 0 {
		timeOut = 3
	}
	endTime := time.Now().Add(time.Second * time.Duration(timeOut))
	for time.Now().Before(endTime) {
		for _, t := range Type {
			if val, ok := ringMsgMap.Load(id + t); ok {
				v := val.(*ringMsg)
				ringMsgMap.Delete(id + t)
				return v.Writer, v.Val
			}
		}
		time.Sleep(time.Millisecond * 50)
	}
	return nil, nil
}

func handlerMsg(msg []byte, writer *common.Writer) {
	// 硬打印：无论 logger 级别，只要收到包就能看到
	fmt.Printf("[RX] %s %s  len=%d\n", writer.Type(), writer.GetAddrStr(), len(msg))

	message := model_proto.RendezvousMessage{}
	err := proto.Unmarshal(msg, &message)
	if err != nil {
		fmt.Printf("[RX] unmarshal error: %v  raw=%x\n", err, msg)
		logs.Err(err)
		return
	}
	if message.Union == nil {
		fmt.Printf("[RX] empty Union from %s\n", writer.GetAddrStr())
		return
	}
	if blacklistDetection("", writer.GetAddr()) {
		return
	}

	msgType := reflect.TypeOf(message.Union).String()
	fmt.Printf("[RX] type=%s  from=%s/%s\n", msgType, writer.Type(), writer.GetAddrStr())
	logs.Debug(writer.Type(), writer.GetAddrStr(), msgType)

	var response proto.Message
	switch msgType {
	case model_proto.TypeRendezvousMessagePunchHoleRequest:
		req := message.GetPunchHoleRequest()
		if req == nil {
			return
		}
		response = model_proto.NewRendezvousMessage(RendezvousMessagePunchHoleRequest(req, writer))
	case model_proto.TypeRendezvousMessageRegisterPk:
		req := message.GetRegisterPk()
		if req == nil {
			return
		}
		response = model_proto.NewRendezvousMessage(RendezvousMessageRegisterPk(req, writer))
	case model_proto.TypeRendezvousMessageRegisterPeer:
		req := message.GetRegisterPeer()
		if req == nil {
			return
		}
		peer := RendezvousMessageRegisterPeer(req, writer)
		response = model_proto.NewRendezvousMessage(peer)
		ConfigureUpdate(writer)
	case model_proto.TypeRendezvousMessageSoftwareUpdate:
		req := message.GetSoftwareUpdate()
		if req == nil {
			return
		}
		response = model_proto.NewRendezvousMessage(RendezvousMessageSoftwareUpdate(req))
	case model_proto.TypeRendezvousMessageTestNatRequest:
		req := message.GetTestNatRequest()
		if req == nil {
			return
		}
		response = model_proto.NewRendezvousMessage(RendezvousMessageTestNatRequest(req, writer))
	case model_proto.TypeRendezvousMessageLocalAddr:
		req := message.GetLocalAddr()
		if req == nil {
			return
		}
		RendezvousMessageLocalAddr(req, writer)
	case model_proto.TypeRendezvousMessageRequestRelay:
		req := message.GetRequestRelay()
		if req == nil {
			return
		}
		response = model_proto.NewRendezvousMessage(RendezvousMessageRequestRelay(req))
	case model_proto.TypeRendezvousMessageRelayResponse:
		req := message.GetRelayResponse()
		if req == nil {
			return
		}
		RendezvousMessageRelayResponse(writer, req)
	case model_proto.TypeRendezvousMessagePunchHoleSent:
		req := message.GetPunchHoleSent()
		if req == nil {
			return
		}
		RendezvousMessagePunchHoleSent(req, writer)
	case model_proto.TypeRendezvousMessageConfigureUpdate:
		req := message.GetConfigureUpdate()
		if req == nil {
			return
		}
		RendezvousMessageConfigureUpdate(req)
	case model_proto.TypeRendezvousMessageOnlineRequest:
		req := message.GetOnlineRequest()
		if req == nil {
			return
		}
		response = model_proto.NewRendezvousMessage(RendezvousMessageOnlineRequest(req))
	default:
		fmt.Printf("[RX] UNKNOWN type: %s\n", msgType)
		logs.Debug("unknown:", msgType)
	}
	if response != nil {
		fmt.Printf("[TX] sending response to %s/%s\n", writer.Type(), writer.GetAddrStr())
		writer.SendMsg(response)
	}
}
