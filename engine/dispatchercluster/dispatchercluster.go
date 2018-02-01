package dispatchercluster

import (
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/common"
)

var (
	dispatcherConns []*dispatcherclient.DispatcherConnMgr
)

func Initialize() {
	dispIds := config.GetDispatcherIDs()
	dispatcherConns = make([]*dispatcherclient.DispatcherConnMgr, len(dispIds))
	for _, dispid := range dispIds {
		dispatcherConns[dispid-1] = dispatcherclient.NewDispatcherConnMgr(dispid)
	}
}

func SendNotifyDestroyEntity(id common.EntityID) error()  {
	return nil
}