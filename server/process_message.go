package server

import (
	"go-rustdesk-server/cmd"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model"
	"go-rustdesk-server/model/model_proto"
	"go-rustdesk-server/my_bytes"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync"
	"time"
)

var (
	memoryPeerMap sync.Map
	serial        int32 = 1
	// punchHoleGroup coalesces concurrent/duplicate PunchHoleRequests for
	// the same target id so only one notify-and-wait round trip to the
	// peer actually happens, see coalesce.go.
	punchHoleGroup = newCallGroup()
)

var relayAddr = func() string {
	return common.OutboundIP() + ":21117"
}()

func blacklistDetection(_ string, _ *common.Addr) bool { return false }
func getRelay() string                                 { return relayAddr }

func RendezvousMessageRegisterPeer(message *model_proto.RegisterPeer, writer *common.Writer) *model_proto.RegisterPeerResponse {
	res := &model_proto.RegisterPeerResponse{}
	var peer *model.Peer
	if val, ok := memoryPeerMap.Load(message.GetId()); ok {
		peer = val.(*model.Peer)
	}
	if peer == nil {
		res.RequestPk = true
		return res
	}
	ipChange := false
	w, err := common.GetWriter(message.GetId(), common.UDP)
	if err != nil || w.GetAddrStr() != writer.GetAddrStr() {
		ipChange = true
	}
	now := time.Now()
	res.RequestPk = len(peer.PK) == 0 || ipChange
	if ipChange {
		peer.IP = writer.GetAddr().GetIP()
		peer.LastRegTime = &now
		memoryPeerMap.Store(peer.ID, peer)
		writer.SetKey(message.GetId())
	}
	return res
}

func RendezvousMessageRegisterPk(message *model_proto.RegisterPk, writer *common.Writer) *model_proto.RegisterPkResponse {
	res := &model_proto.RegisterPkResponse{Result: model_proto.RegisterPkResponse_SERVER_ERROR}
	if len(message.GetId()) < 6 {
		res.Result = model_proto.RegisterPkResponse_UUID_MISMATCH
		cmd.Info("RegisterPk id=%s too short len=%d", message.GetId(), len(message.GetId()))
		return res
	}
	change := false
	id := message.GetId()
	if message.GetOldId() != "" {
		change = true
		id = message.GetOldId()
	}
	var idPeer *model.Peer
	if val, ok := memoryPeerMap.Load(id); ok {
		idPeer = val.(*model.Peer)
	}
	if idPeer == nil {
		memoryPeerMap.Range(func(k, v interface{}) bool {
			p := v.(*model.Peer)
			if p.UUID == string(message.GetUuid()) {
				idPeer = p
				return false
			}
			return true
		})
	}
	if idPeer != nil && !change {
		if idPeer.UUID == string(message.GetUuid()) {
			if string(idPeer.PK) != string(message.GetPk()) {
				if idPeer.IP != writer.GetAddr().GetIP() {
					res.Result = model_proto.RegisterPkResponse_UUID_MISMATCH
					return res
				}
				change = true
			}
		} else {
			res.Result = model_proto.RegisterPkResponse_ID_EXISTS
			return res
		}
		res.Result = model_proto.RegisterPkResponse_OK
	}
	getWriter, err := common.GetWriter(message.GetId(), common.UDP)
	ipChange := err == nil && getWriter.GetAddrStr() != writer.GetAddrStr()
	if ipChange && idPeer != nil {
		idPeer.IP = writer.GetAddr().GetIP()
		writer.SetKey(idPeer.ID)
	}
	change = ipChange || change || idPeer == nil
	uid := ""
	if idPeer != nil {
		uid = idPeer.Uid
	}
	if change {
		now := time.Now()
		peer := &model.Peer{
			Uid: uid, ID: message.Id, UUID: string(message.Uuid),
			PK: message.Pk, LastRegTime: &now, IP: writer.GetAddr().GetIP(),
		}
		memoryPeerMap.Store(peer.ID, peer)
		res.Result = model_proto.RegisterPkResponse_OK
		writer.SetKey(message.GetId())
	}
	cmd.Info("RegisterPk id=%s result=%v change=%v", message.GetId(), res.Result, change)
	return res
}

func RendezvousMessageSoftwareUpdate(_ *model_proto.SoftwareUpdate) *model_proto.SoftwareUpdate {
	return &model_proto.SoftwareUpdate{}
}

func RendezvousMessagePunchHoleRequest(message *model_proto.PunchHoleRequest, writer *common.Writer) protoreflect.ProtoMessage {
	res := &model_proto.PunchHoleResponse{}
	var peer *model.Peer
	if val, ok := memoryPeerMap.Load(message.Id); ok {
		peer = val.(*model.Peer)
	}
	if peer == nil {
		res.Failure = model_proto.PunchHoleResponse_ID_NOT_EXIST
		return res
	}
	getPeer, err := common.GetWriter(peer.ID, common.UDP)
	if err != nil {
		res.Failure = model_proto.PunchHoleResponse_OFFLINE
		return res
	}
	peerIsLan := common.InSubnet(peer.IP)
	isLan := common.InSubnet(writer.GetAddr().GetIP())
	natType := message.GetNatType()
	relay := getRelay()
	cmd.Info("peerIsLan %v isLan %v", peerIsLan, isLan)
	if peerIsLan != isLan {
		natType = model_proto.NatType_SYMMETRIC
	}

	sameIntranet := writer.GetAddr().GetIP() == peer.IP
	cmd.Info("sameIntranet %v natType %v", sameIntranet, natType)
	cmd.Info("relay addr: %s", relay)

	dedupKey := peer.ID
	if sameIntranet {
		dedupKey += ":local"
	}

	// Whoever gets here first for this target id actually notifies the
	// peer and waits; anyone else asking for the same id while that's
	// still in flight (typically the same caller retrying) shares the
	// result instead of starting another redundant round trip.
	w, lMsg := punchHoleGroup.Do(dedupKey, func() (*common.Writer, interface{}) {
		if sameIntranet {
			getPeer.SendMsg(model_proto.NewRendezvousMessage(&model_proto.FetchLocalAddr{
				SocketAddr: my_bytes.EncodeAddr(writer.GetAddrStr()), RelayServer: relay,
			}))
			return getMsgForm(message.GetId(), []string{model_proto.TypeRendezvousMessageLocalAddr}, 3)
		}
		getPeer.SendMsg(model_proto.NewRendezvousMessage(&model_proto.PunchHole{
			SocketAddr: my_bytes.EncodeAddr(writer.GetAddrStr()), RelayServer: relay, NatType: natType,
		}))
		return getMsgForm(message.GetId(), []string{
			model_proto.TypeRendezvousMessagePunchHoleSent,
			model_proto.TypeRendezvousMessageRelayResponse,
		}, 3)
	})

	if lMsg == nil {
		res.OtherFailure = "NoReturnMessage"
		return res
	}

	if sameIntranet {
		if m, ok := lMsg.(*model_proto.LocalAddr); ok {
			res.SocketAddr = m.GetLocalAddr()
			res.RelayServer = m.GetRelayServer()
			res.Pk = common.GetSignPK(m.GetVersion(), peer.ID, peer.PK)
			res.Union = &model_proto.PunchHoleResponse_IsLocal{IsLocal: true}
		}
		return res
	}

	if m, ok := lMsg.(*model_proto.PunchHoleSent); ok {
		cmd.Info("Responding with relay=%s", relay)
		res.SocketAddr = my_bytes.EncodeAddr(w.GetAddrStr())
		res.RelayServer = relay
		res.Pk = common.GetSignPK(m.GetVersion(), peer.ID, peer.PK)
		res.Union = &model_proto.PunchHoleResponse_NatType{NatType: m.GetNatType()}
	}
	if m, ok := lMsg.(*model_proto.RelayResponse); ok {
		m.SocketAddr = my_bytes.EncodeAddr(w.GetAddrStr())
		m.RelayServer = m.GetRelayServer()
		m.Union = &model_proto.RelayResponse_Pk{Pk: common.GetSignPK(m.GetVersion(), peer.ID, peer.PK)}
		return m
	}
	return res
}

func RendezvousMessageTestNatRequest(message *model_proto.TestNatRequest, writer *common.Writer) *model_proto.TestNatResponse {
	return &model_proto.TestNatResponse{
		Port: int32(writer.GetAddr().GetPort()),
		Cu: &model_proto.ConfigUpdate{
			Serial: message.Serial, RendezvousServers: []string{relayAddr},
		},
	}
}

func RendezvousMessageLocalAddr(message *model_proto.LocalAddr, writer *common.Writer) {
	msgNotifier.publish(message.GetId()+model_proto.TypeRendezvousMessageLocalAddr, writer, message)
}

func RendezvousMessageRequestRelay(message *model_proto.RequestRelay) *model_proto.RelayResponse {
	res := &model_proto.RelayResponse{}
	w, err := common.GetWriter(message.GetId(), common.UDP)
	if err != nil {
		return nil
	}
	var peer *model.Peer
	if val, ok := memoryPeerMap.Load(message.Id); ok {
		peer = val.(*model.Peer)
	}
	if peer == nil {
		return res
	}
	w.SendMsg(model_proto.NewRendezvousMessage(message))
	_, lMsg := getMsgForm(message.GetId(), []string{model_proto.TypeRendezvousMessageRelayResponse}, 3)
	if lMsg == nil {
		return res
	}
	if m, ok := lMsg.(*model_proto.RelayResponse); ok {
		m.Union = &model_proto.RelayResponse_Pk{Pk: common.GetSignPK(m.GetVersion(), m.GetId(), peer.PK)}
		res = m
	}
	return res
}

func RendezvousMessageRelayResponse(w *common.Writer, message *model_proto.RelayResponse) {
	msgNotifier.publish(message.GetId()+model_proto.TypeRendezvousMessageRelayResponse, w, message)
}

func ConfigureUpdate(writer *common.Writer) {
	writer.SendMsg(model_proto.NewRendezvousMessage(&model_proto.ConfigUpdate{
		Serial: serial, RendezvousServers: []string{relayAddr},
	}))
}

func RendezvousMessagePunchHoleSent(message *model_proto.PunchHoleSent, writer *common.Writer) {
	cmd.Info("PunchHoleSent id=%s relay=%s natType=%v",
		message.GetId(), message.GetRelayServer(), message.GetNatType())
	msgNotifier.publish(message.GetId()+model_proto.TypeRendezvousMessagePunchHoleSent, writer, message)
}

func RendezvousMessageConfigureUpdate(message *model_proto.ConfigUpdate) {
	cmd.Info("ConfigureUpdate serial %v", message.Serial)
}

func RendezvousMessageOnlineRequest(message *model_proto.OnlineRequest) *model_proto.OnlineResponse {
	states := make([]byte, (len(message.GetPeers())+7)/8)
	for i, peerID := range message.GetPeers() {
		var peer *model.Peer
		if val, ok := memoryPeerMap.Load(peerID); ok {
			peer = val.(*model.Peer)
		}
		if peer != nil && peer.LastRegTime != nil && time.Since(*peer.LastRegTime) < 30*time.Second {
			states[i/8] |= 0x01 << (7 - i%8)
		}
	}
	return &model_proto.OnlineResponse{States: states}
}
