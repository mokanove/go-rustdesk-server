package server

import (
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model/model_proto"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync"
	"time"
)

// notifyResult is handed to every goroutine waiting on a given key when
// the matching rendezvous reply (PunchHoleSent / RelayResponse / LocalAddr)
// comes in.
type notifyResult struct {
	Writer *common.Writer
	Val    interface{}
}

// notifier is a small pub/sub used to wake up goroutines blocked on
// getMsgForm. It replaces the old "store in a map, poll every 50ms, first
// reader deletes it" scheme:
//   - wakeup is immediate (channel signal) instead of poll-interval latency
//   - if several goroutines are waiting on the same key at once (this
//     happens whenever a client retries/duplicates a request before
//     getting an answer), ALL of them get the result instead of a single
//     winner stealing it from the others.
type notifier struct {
	mu      sync.Mutex
	waiters map[string][]chan notifyResult
}

var msgNotifier = &notifier{waiters: make(map[string][]chan notifyResult)}

func (n *notifier) subscribe(keys []string) (chan notifyResult, func()) {
	ch := make(chan notifyResult, 1)
	n.mu.Lock()
	for _, k := range keys {
		n.waiters[k] = append(n.waiters[k], ch)
	}
	n.mu.Unlock()

	cancel := func() {
		n.mu.Lock()
		for _, k := range keys {
			arr := n.waiters[k]
			for i, c := range arr {
				if c == ch {
					n.waiters[k] = append(arr[:i], arr[i+1:]...)
					break
				}
			}
			if len(n.waiters[k]) == 0 {
				delete(n.waiters, k)
			}
		}
		n.mu.Unlock()
	}
	return ch, cancel
}

// publish fans val out to every goroutine currently subscribed to key.
func (n *notifier) publish(key string, w *common.Writer, val interface{}) {
	n.mu.Lock()
	subs := n.waiters[key]
	delete(n.waiters, key)
	n.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- notifyResult{Writer: w, Val: val}:
		default:
		}
	}
}

// getMsgForm blocks until a reply tagged with id+(one of types) is
// published, or timeOut seconds elapse. Same signature/semantics as
// before from the caller's point of view, just no busy-polling and no
// "single consumer wins" race.
func getMsgForm(id string, types []string, timeOut uint) (*common.Writer, interface{}) {
	if timeOut == 0 {
		timeOut = 3
	}
	keys := make([]string, len(types))
	for i, t := range types {
		keys[i] = id + t
	}
	ch, cancel := msgNotifier.subscribe(keys)
	defer cancel()

	select {
	case res := <-ch:
		return res.Writer, res.Val
	case <-time.After(time.Duration(timeOut) * time.Second):
		return nil, nil
	}
}

func handlerMsg(msg []byte, writer *common.Writer) {
	cmd.Info("RX %s %s len %d", writer.Type(), writer.GetAddrStr(), len(msg))
	message := model_proto.RendezvousMessage{}
	if err := proto.Unmarshal(msg, &message); err != nil {
		cmd.Fatal("unmarshal", err)
		return
	}
	if message.Union == nil {
		cmd.Info("empty Union from %s", writer.GetAddrStr())
		return
	}
	if blacklistDetection("", writer.GetAddr()) {
		return
	}
	msgType := reflect.TypeOf(message.Union).String()
	cmd.Info("RX type %s from %s %s", msgType, writer.Type(), writer.GetAddrStr())
	var response proto.Message
	switch msgType {
	case model_proto.TypeRendezvousMessagePunchHoleRequest:
		if req := message.GetPunchHoleRequest(); req != nil {
			go func(r *model_proto.PunchHoleRequest, w *common.Writer) {
				resp := RendezvousMessagePunchHoleRequest(r, w)
				if resp != nil {
					cmd.Info("TX response to %s %s", w.Type(), w.GetAddrStr())
					w.SendMsg(model_proto.NewRendezvousMessage(resp))
				}
			}(req, writer)
		}
		return
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
			go func(r *model_proto.RequestRelay, w *common.Writer) {
				resp := RendezvousMessageRequestRelay(r)
				if resp != nil {
					cmd.Info("TX response to %s %s", w.Type(), w.GetAddrStr())
					w.SendMsg(model_proto.NewRendezvousMessage(resp))
				}
			}(req, writer)
		}
		return
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
		cmd.Info("RX unknown type %s", msgType)
	}
	if response != nil {
		cmd.Info("TX response to %s %s", writer.Type(), writer.GetAddrStr())
		writer.SendMsg(response)
	}
}
