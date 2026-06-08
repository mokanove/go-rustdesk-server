package server

import (
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"go-rustdesk-server/model"
	"go-rustdesk-server/model/model_proto"
	"go-rustdesk-server/my_bytes"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync"
	"time"
)

// ================= 纯内存数据库与消息队列平替方案 =================
var (
	memoryPeerMap     sync.Map // key: id (string), value: *model.Peer
	rendezvousServers []string 
	serial            int32    = 1
	ringMsgMap        sync.Map // key: string (id+type), value: *ringMsg
)

// 黑名单检测默认全部放行
func blacklistDetection(id string, addr *common.Addr) bool {
	return false 
}

// 焊死 21117 中继端口
func getRelay() string {
	if common.Conf.RelayName != "" {
		return common.Conf.RelayName
	}
	return ":21117"
}

func RendezvousMessageRegisterPeer(message *model_proto.RegisterPeer, writer *common.Writer) *model_proto.RegisterPeerResponse {
	res := &model_proto.RegisterPeerResponse{}

	var peer *model.Peer
	if val, ok := memoryPeerMap.Load(message.GetId()); ok {
		peer = val.(*model.Peer)
	}

	if peer == nil {
		res.RequestPk = true
	} else {
		ipChange := false
		res.RequestPk = false
		w, err1 := common.GetWriter(message.GetId(), common.UDP)
		if err1 != nil {
			ipChange = true
		} else if w.GetAddrStr() != writer.GetAddrStr() {
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
	}
	return res
}

func RendezvousMessageRegisterPk(message *model_proto.RegisterPk, writer *common.Writer) *model_proto.RegisterPkResponse {
	res := &model_proto.RegisterPkResponse{Result: model_proto.RegisterPkResponse_SERVER_ERROR}

	if len(message.GetId()) < common.MinKeyLen {
		res.Result = model_proto.RegisterPkResponse_UUID_MISMATCH
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
		memoryPeerMap.Range(func(key, value interface{}) bool {
			p := value.(*model.Peer)
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
	ipChange := false
	if err == nil {
		if getWriter.GetAddrStr() != writer.GetAddrStr() {
			ipChange = true
		}
	}
	if ipChange && idPeer != nil {
		idPeer.IP = writer.GetAddr().GetIP()
		writer.SetKey(idPeer.ID)
	}
	change = ipChange || change || idPeer == nil
	uid := ""
	if idPeer != nil && idPeer.Uid != "" {
		uid = idPeer.Uid
	}
	if change {
		now := time.Now()
		peer := &model.Peer{
			Uid:         uid,
			ID:          message.Id,
			UUID:        string(message.Uuid),
			PK:          message.Pk,
			LastRegTime: &now,
			IP:          writer.GetAddr().GetIP(),
		}
		memoryPeerMap.Store(peer.ID, peer)
		res.Result = model_proto.RegisterPkResponse_OK
		writer.SetKey(message.GetId())
	}

	return res
}

func RendezvousMessageSoftwareUpdate(message *model_proto.SoftwareUpdate) *model_proto.SoftwareUpdate {
	res := &model_proto.SoftwareUpdate{}
	return res
}

func RendezvousMessagePunchHoleRequest(message *model_proto.PunchHoleRequest, writer *common.Writer) protoreflect.ProtoMessage {
	res := &model_proto.PunchHoleResponse{}
	if common.Conf.MustKey {
		if message.LicenceKey != common.GetPkStr() {
			res.Failure = model_proto.PunchHoleResponse_LICENSE_MISMATCH
			return res
		}
	}

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

	writer.GetAddr().GetIP()
	peerIsLan := common.InSubnet(peer.IP)
	isLan := common.InSubnet(writer.GetAddr().GetIP())
	natType := message.GetNatType()
	relayServer := getRelay()
	logs.Debug("peerIsLan", peerIsLan, "isLan", isLan)
	if peerIsLan != isLan {
		if peerIsLan {
			relayServer = writer.SelfAddr()
		}
		natType = model_proto.NatType_SYMMETRIC
	}
	sameIntranet := writer.GetAddr().GetIP() == peer.IP
	logs.Debug("sameIntranet", sameIntranet, natType)
	if sameIntranet {
		getPeer.SendMsg(model_proto.NewRendezvousMessage(&model_proto.FetchLocalAddr{
			SocketAddr:  my_bytes.EncodeAddr(writer.GetAddrStr()),
			RelayServer: relayServer,
		}))
		_, lMsg := getMsgForm(message.GetId(), []string{model_proto.TypeRendezvousMessageLocalAddr}, 3)
		if lMsg == nil {
			res.OtherFailure = "NoReturnMessage"
			return res
		}
		if m, ok1 := lMsg.(*model_proto.LocalAddr); ok1 {
			logs.Debug("LocalAddr", m.GetLocalAddr())
			res.SocketAddr = m.GetLocalAddr()
			res.RelayServer = m.GetRelayServer()
			res.Pk = common.GetSignPK(m.GetVersion(), peer.ID, peer.PK)
			res.Union = &model_proto.PunchHoleResponse_IsLocal{IsLocal: true}
		}
	} else {
		logs.Debug("PunchHole", writer.GetAddrStr())
		getPeer.SendMsg(model_proto.NewRendezvousMessage(&model_proto.PunchHole{
			SocketAddr:  my_bytes.EncodeAddr(writer.GetAddrStr()),
			RelayServer: relayServer,
			NatType:     natType,
		}))
		w, lMsg := getMsgForm(message.GetId(), []string{model_proto.TypeRendezvousMessagePunchHoleSent, model_proto.TypeRendezvousMessageRelayResponse}, 3)
		if lMsg == nil {
			res.OtherFailure = "NoReturnMessage"
			return res
		}

		if m, ok1 := lMsg.(*model_proto.PunchHoleSent); ok1 {
			logs.Debug("PunchHoleSent", w.GetAddrStr(), my_bytes.DecodeAddr(m.GetSocketAddr()))
			res.SocketAddr = my_bytes.EncodeAddr(w.GetAddrStr())
			res.RelayServer = m.GetRelayServer()
			res.Pk = common.GetSignPK(m.GetVersion(), peer.ID, peer.PK)
			res.Union = &model_proto.PunchHoleResponse_NatType{
				NatType: m.GetNatType(),
			}
		}
		if m, ok1 := lMsg.(*model_proto.RelayResponse); ok1 {
			m.SocketAddr = my_bytes.EncodeAddr(w.GetAddrStr())
			m.RelayServer = m.GetRelayServer()
			m.Union = &model_proto.RelayResponse_Pk{
				Pk: common.GetSignPK(m.GetVersion(), peer.ID, peer.PK),
			}
			return m
		}
	}
	return res
}

func RendezvousMessageTestNatRequest(message *model_proto.TestNatRequest, writer *common.Writer) *model_proto.TestNatResponse {
	res := &model_proto.TestNatResponse{
		Port: int32(writer.GetAddr().GetPort()),
	}
	res.Cu = &model_proto.ConfigUpdate{
		Serial:            message.Serial,
		RendezvousServers: rendezvousServers,
	}
	return res
}

func RendezvousMessageLocalAddr(message *model_proto.LocalAddr, writer *common.Writer) {
	ringMsgMap.Store(message.GetId()+model_proto.TypeRendezvousMessageLocalAddr, &ringMsg{
		ID:      message.GetId(),
		Type:    model_proto.TypeRendezvousMessageLocalAddr,
		TimeOut: 3,
		InsTime: time.Now(),
		Val:     message,
		Writer:  writer,
	})
}

func RendezvousMessageRequestRelay(message *model_proto.RequestRelay) *model_proto.RelayResponse {
	res := &model_proto.RelayResponse{}
	w, err := common.GetWriter(message.GetId(), common.UDP)
	if err != nil {
		return nil
	} else {
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
		} else if m, ok1 := lMsg.(*model_proto.RelayResponse); ok1 {
			m.Union = &model_proto.RelayResponse_Pk{
				Pk: common.GetSignPK(m.GetVersion(), m.GetId(), peer.PK),
			}
			res = m
		}
	}
	return res
}

func RendezvousMessageRelayResponse(w *common.Writer, message *model_proto.RelayResponse) {
	ringMsgMap.Store(message.GetId()+model_proto.TypeRendezvousMessageRelayResponse, &ringMsg{
		ID:      message.GetId(),
		Type:    model_proto.TypeRendezvousMessageRelayResponse,
		TimeOut: 3,
		InsTime: time.Now(),
		Val:     message,
		Writer:  w,
	})
}

func ConfigureUpdate(writer *common.Writer) {
	// 焊死下发给客户端的配置，让它们去撞 21116
	relaySrv := common.Conf.RelayName
	if relaySrv == "" {
		relaySrv = ":21116" 
	}

	writer.SendMsg(model_proto.NewRendezvousMessage(&model_proto.ConfigUpdate{
		Serial:            serial,
		RendezvousServers: []string{relaySrv}, 
	}))
}

func RendezvousMessagePunchHoleSent(message *model_proto.PunchHoleSent, writer *common.Writer) {
	ringMsgMap.Store(message.GetId()+model_proto.TypeRendezvousMessagePunchHoleSent, &ringMsg{
		ID:      message.GetId(),
		Type:    model_proto.TypeRendezvousMessagePunchHoleSent,
		TimeOut: 3,
		InsTime: time.Now(),
		Val:     message,
		Writer:  writer,
	})
}

func RendezvousMessageConfigureUpdate(message *model_proto.ConfigUpdate) {
	logs.Debug(message.Serial, message.RendezvousServers)
}

func RendezvousMessageOnlineRequest(message *model_proto.OnlineRequest) *model_proto.OnlineResponse {
	states := make([]byte, (len(message.GetPeers())+7)/8)
	for i, peerID := range message.GetPeers() {
		var peer *model.Peer
		if val, ok := memoryPeerMap.Load(peerID); ok {
			peer = val.(*model.Peer)
		}
		if peer == nil || peer.LastRegTime == nil {
			continue
		}
		statesIdx := i / 8
		bitIdx := 7 - i%8
		if time.Since(*peer.LastRegTime) < time.Millisecond*30000 {
			states[statesIdx] |= 0x01 << bitIdx
		}
	}
	return &model_proto.OnlineResponse{
		States: states,
	}
}