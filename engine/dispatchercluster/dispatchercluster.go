package dispatchercluster

import (
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

var (
	dispatcherConns []*dispatcherclient.DispatcherConnMgr
	dispatcherNum   int
	gid             uint16
)

func Initialize(_gid uint16, dctype dispatcherclient.DispatcherClientType, isRestoreGame, isBanBootEntity bool, delegate dispatcherclient.IDispatcherClientDelegate) {
	gid = _gid
	if gid == 0 {
		gwlog.Fatalf("gid is 0")
	}

	dispIds := config.GetDispatcherIDs()
	dispatcherNum = len(dispIds)
	if dispatcherNum == 0 {
		gwlog.Fatalf("dispatcher number is 0")
	}

	dispatcherConns = make([]*dispatcherclient.DispatcherConnMgr, dispatcherNum)
	for _, dispid := range dispIds {
		dispatcherConns[dispid-1] = dispatcherclient.NewDispatcherConnMgr(gid, dctype, dispid, isRestoreGame, isBanBootEntity, delegate)
	}
	for _, dispConn := range dispatcherConns {
		dispConn.Connect()
	}
}

func SendNotifyDestroyEntity(id common.EntityID) {
	SelectByEntityID(id).SendNotifyDestroyEntity(id)
}

func SendMigrateRequest(entityID common.EntityID, spaceID common.EntityID, spaceGameID uint16) {
	SelectByEntityID(entityID).SendMigrateRequest(entityID, spaceID, spaceGameID)
}

func SendRealMigrate(eid common.EntityID, targetGame uint16, data []byte) {
	SelectByEntityID(eid).SendRealMigrate(eid, targetGame, data)
}
func SendCallFilterClientProxies(op proto.FilterClientsOpType, key, val string, method string, args []interface{}) {
	pkt := proto.AllocCallFilterClientProxiesPacket(op, key, val, method, args)
	broadcast(pkt)
	pkt.Release()
	return
}

func broadcast(packet *netutil.Packet) {
	for _, dcm := range dispatcherConns {
		dcm.GetDispatcherClientForSend().SendPacket(packet)
	}
}

func SendNotifyCreateEntity(id common.EntityID) {
	if gid != 0 {
		SelectByEntityID(id).SendNotifyCreateEntity(id)
	} else {
		// goes here when creating nil space or restoring freezed entities
	}
}

func SendLoadEntityAnywhere(typeName string, entityID common.EntityID) {
	SelectByEntityID(entityID).SendLoadEntitySomewhere(typeName, entityID, 0)
}

func SendLoadEntityOnGame(typeName string, entityID common.EntityID, gameid uint16) {
	SelectByEntityID(entityID).SendLoadEntitySomewhere(typeName, entityID, gameid)
}

func SendCreateEntitySomewhere(gameid uint16, entityid common.EntityID, typeName string, data map[string]interface{}) {
	SelectByEntityID(entityid).SendCreateEntitySomewhere(gameid, entityid, typeName, data)
}

func SendGameLBCInfo(lbcinfo proto.GameLBCInfo) {
	packet := proto.AllocGameLBCInfoPacket(lbcinfo)
	broadcast(packet)
	packet.Release()
}

func SendStartFreezeGame() {
	pkt := proto.AllocStartFreezeGamePacket()
	broadcast(pkt)
	pkt.Release()
	return
}

func SendKvregRegister(srvid string, info string, force bool) {
	SelectBySrvID(srvid).SendKvregRegister(srvid, info, force)
}

func SendCallNilSpaces(exceptGameID uint16, method string, args []interface{}) {
	// construct one packet for multiple sending
	packet := proto.AllocCallNilSpacesPacket(exceptGameID, method, args)
	broadcast(packet)
	packet.Release()
}

func EntityIDToDispatcherID(entityid common.EntityID) uint16 {
	return uint16((hashEntityID(entityid) % dispatcherNum) + 1)
}

func SrvIDToDispatcherID(srvid string) uint16 {
	return uint16((hashSrvID(srvid) % dispatcherNum) + 1)
}

func SelectByEntityID(entityid common.EntityID) *dispatcherclient.DispatcherClient {
	idx := hashEntityID(entityid) % dispatcherNum
	return dispatcherConns[idx].GetDispatcherClientForSend()
}

func SelectByGateID(gateid uint16) *dispatcherclient.DispatcherClient {
	idx := hashGateID(gateid) % dispatcherNum
	return dispatcherConns[idx].GetDispatcherClientForSend()
}

func SelectByDispatcherID(dispid uint16) *dispatcherclient.DispatcherClient {
	return dispatcherConns[dispid-1].GetDispatcherClientForSend()
}

func SelectBySrvID(srvid string) *dispatcherclient.DispatcherClient {
	idx := hashSrvID(srvid) % dispatcherNum
	return dispatcherConns[idx].GetDispatcherClientForSend()
}

func Select(dispidx int) *dispatcherclient.DispatcherClient {
	return dispatcherConns[dispidx].GetDispatcherClientForSend()
}
