package main

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/server"
	"github.com/xiaonanln/goworld/gwlog"
)

func init() {

}

type serverDelegate struct {
	server.ServerDelegate
}

func main() {
	goworld.RegisterEntity("Account", &Account{})
	goworld.RegisterEntity("OnlineService", &OnlineService{})
	goworld.RegisterEntity("Monster", &Monster{})
	goworld.RegisterEntity("Avatar", &Avatar{})

	goworld.SetSpaceDelegate(&SpaceDelegate{})

	goworld.Run(&serverDelegate{})
}

func (server serverDelegate) OnServerReady() {
	server.ServerDelegate.OnServerReady()

	eids := goworld.ListEntityIDs("OnlineService")
	gwlog.Info("Found saved OnlineService ids: %v", eids)

	if len(eids) == 0 {
		goworld.CreateEntityAnywhere("OnlineService")
	} else {
		// already exists
		onlineServiceID := eids[0]
		goworld.LoadEntityAnywhere("OnlineService", onlineServiceID)
	}

	timer.AddCallback(time.Millisecond*1000, server.checkServerStarted)
}

func (server serverDelegate) checkServerStarted() {
	ok := server.isServerStarted()
	gwlog.Info("checkServerStarted: %v", ok)
	if ok {
		server.onServerStarted()
	} else {
		timer.AddCallback(time.Millisecond*1000, server.checkServerStarted)
	}
}

func (server serverDelegate) isServerStarted() bool {
	if len(goworld.GetServiceProviders("OnlineService")) == 0 {
		return false
	}
	return true
}

func (server serverDelegate) onServerStarted() {
	goworld.CreateSpaceAnywhere()
}
