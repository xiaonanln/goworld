package dispatchercluster

import (
	"time"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
)

var (
	dispatcherConns []*dispatcherclient.DispatcherConnMgr
	dispatcherNum   int
	gid             uint16
)

func Initialize(_gid uint16, dctype dispatcherclient.DispatcherClientType, isRestoreGame bool, delegate dispatcherclient.IDispatcherClientDelegate) {
	gid = _gid
	dispIds := config.GetDispatcherIDs()
	dispatcherNum = len(dispIds)
	dispatcherConns = make([]*dispatcherclient.DispatcherConnMgr, dispatcherNum)
	for _, dispid := range dispIds {
		dispatcherConns[dispid-1] = dispatcherclient.NewDispatcherConnMgr(gid, dctype, dispid, isRestoreGame, delegate)
	}
	for _, dispConn := range dispatcherConns {
		dispConn.Connect()
	}

	go gwutils.RepeatUntilPanicless(autoFlushRoutine)
}

func SendNotifyDestroyEntity(id common.EntityID) error {
	return SelectByEntityID(id).SendNotifyDestroyEntity(id)
}

func SendDeclareService(id common.EntityID, serviceName string) error {
	return SelectByEntityID(id).SendDeclareService(id, serviceName)
}

func SendClearClientFilterProp(gateid uint16, clientid common.ClientID) (err error) {
	return SelectByGateID(gateid).SendClearClientFilterProp(gateid, clientid)
}

func SendSetClientFilterProp(gateid uint16, clientid common.ClientID, key, val string) (err error) {
	return SelectByGateID(gateid).SendSetClientFilterProp(gateid, clientid, key, val)
}

func SendMigrateRequest(spaceID common.EntityID, entityID common.EntityID) error {
	return SelectByEntityID(entityID).SendMigrateRequest(spaceID, entityID)
}

func SendRealMigrate(eid common.EntityID, targetGame uint16, targetSpace common.EntityID, x, y, z float32,
	typeName string, migrateData map[string]interface{}, timerData []byte, clientid common.ClientID, clientsrv uint16) error {
	return SelectByEntityID(eid).SendRealMigrate(eid, targetGame, targetSpace, x, y, z, typeName, migrateData, timerData, clientid, clientsrv)
}
func SendCallFilterClientProxies(key string, val string, method string, args []interface{}) (anyerror error) {
	for _, dcm := range dispatcherConns {
		err := dcm.GetDispatcherClientForSend().SendCallFilterClientProxies(key, val, method, args)
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
	for _, dispconn := range dispatcherConns {
		err := dispconn.GetDispatcherClientForSend().SendStartFreezeGame(gameid)
		if err != nil {
			anyerror = err
		}
	}
	return
}

func SelectByEntityID(id common.EntityID) *dispatcherclient.DispatcherClient {
	idx := hashEntityID(id) % dispatcherNum
	return dispatcherConns[idx].GetDispatcherClientForSend()
}

func SelectByGateID(gateid uint16) *dispatcherclient.DispatcherClient {
	idx := hashGateID(gateid) % dispatcherNum
	return dispatcherConns[idx].GetDispatcherClientForSend()
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
