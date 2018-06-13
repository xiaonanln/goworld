package dispatchercluster

import (
	"time"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
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
	dispIds := config.GetDispatcherIDs()
	dispatcherNum = len(dispIds)
	dispatcherConns = make([]*dispatcherclient.DispatcherConnMgr, dispatcherNum)
	for _, dispid := range dispIds {
		dispatcherConns[dispid-1] = dispatcherclient.NewDispatcherConnMgr(gid, dctype, dispid, isRestoreGame, isBanBootEntity, delegate)
	}
	for _, dispConn := range dispatcherConns {
		dispConn.Connect()
	}

	go gwutils.RepeatUntilPanicless(autoFlushRoutine)
}

func SendNotifyDestroyEntity(id common.EntityID) error {
	return SelectByEntityID(id).SendNotifyDestroyEntity(id)
}

func SendClearClientFilterProp(gateid uint16, clientid common.ClientID) (err error) {
	return SelectByGateID(gateid).SendClearClientFilterProp(gateid, clientid)
}

func SendSetClientFilterProp(gateid uint16, clientid common.ClientID, key, val string) (err error) {
	return SelectByGateID(gateid).SendSetClientFilterProp(gateid, clientid, key, val)
}

func SendMigrateRequest(entityID common.EntityID, spaceID common.EntityID, spaceGameID uint16) error {
	return SelectByEntityID(entityID).SendMigrateRequest(entityID, spaceID, spaceGameID)
}

func SendRealMigrate(eid common.EntityID, targetGame uint16, targetSpace common.EntityID, x, y, z float32,
	typeName string, migrateData map[string]interface{}, timerData []byte, clientid common.ClientID, clientsrv uint16) error {
	return SelectByEntityID(eid).SendRealMigrate(eid, targetGame, targetSpace, x, y, z, typeName, migrateData, timerData, clientid, clientsrv)
}
func SendCallFilterClientProxies(op proto.FilterClientsOpType, key, val string, method string, args []interface{}) (anyerror error) {
	for _, dcm := range dispatcherConns {
		err := dcm.GetDispatcherClientForSend().SendCallFilterClientProxies(op, key, val, method, args)
		if err != nil && anyerror == nil {
			anyerror = err
		}
	}
	return
}

func SendNotifyCreateEntity(id common.EntityID) error {
	return SelectByEntityID(id).SendNotifyCreateEntity(id)
}

func SendLoadEntityAnywhere(typeName string, entityID common.EntityID) error {
	return SelectByEntityID(entityID).SendLoadEntityAnywhere(typeName, entityID)
}

func SendCreateEntityAnywhere(entityid common.EntityID, typeName string, data map[string]interface{}) error {
	return SelectByEntityID(entityid).SendCreateEntityAnywhere(entityid, typeName, data)
}

func SendStartFreezeGame(gameid uint16) (anyerror error) {
	for _, dcm := range dispatcherConns {
		err := dcm.GetDispatcherClientForSend().SendStartFreezeGame(gameid)
		if err != nil {
			anyerror = err
		}
	}
	return
}

func SendSrvdisRegister(srvid string, info string, force bool) {
	SelectBySrvID(srvid).SendSrvdisRegister(srvid, info, force)
}

func SendCallNilSpaces(exceptGameID uint16, method string, args []interface{}) (anyerror error) {
	// construct one packet for multiple sending
	packet := netutil.NewPacket()
	packet.AppendUint16(proto.MT_CALL_NIL_SPACES)
	packet.AppendUint16(exceptGameID)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)

	for _, dcm := range dispatcherConns {
		err := dcm.GetDispatcherClientForSend().SendPacket(packet)
		if err != nil {
			anyerror = err
		}
	}

	packet.Release()
	return
}

func EntityIDToDispatcherID(entityid common.EntityID) uint16 {
	return uint16((hashEntityID(entityid) % dispatcherNum) + 1)
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

//func Flush(reason string) (anyerror error) {
//	for _, dispconn := range dispatcherConns {
//		err := dispconn.GetDispatcherClientForSend().Flush(reason)
//		if err != nil {
//			anyerror = err
//		}
//	}
//	return
//}

func autoFlushRoutine() {
	// TODO: each dipsatcher client flush by itself
	for {
		time.Sleep(10 * time.Millisecond)
		for _, dispconn := range dispatcherConns {
			err := dispconn.GetDispatcherClientForSend().Flush("dispatchercluster")
			if err != nil {
				gwlog.Errorf("dispatchercluster: %s flush failed: %v", dispconn, err)
			}
		}
	}
}
