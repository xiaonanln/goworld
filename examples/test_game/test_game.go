package main

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/server"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	SERVICE_NAMES = []string{
		"OnlineService",
	}
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

	if goworld.GetServerID() == 1 { // Create services on just 1 server
		for _, serviceName := range SERVICE_NAMES {
			eids := goworld.ListEntityIDs(serviceName)
			gwlog.Info("Found saved %s ids: %v", serviceName, eids)

			if len(eids) == 0 {
				goworld.CreateEntityAnywhere(serviceName)
			} else {
				// already exists
				serviceID := eids[0]
				goworld.LoadEntityAnywhere(serviceName, serviceID)
			}
		}
	}

	timer.AddCallback(time.Millisecond*1000, server.checkServerStarted)
}

func (server serverDelegate) checkServerStarted() {
	ok := server.isAllServicesReady()
	gwlog.Info("checkServerStarted: %v", ok)
	if ok {
		server.onAllServicesReady()
	} else {
		timer.AddCallback(time.Millisecond*1000, server.checkServerStarted)
	}
}

func (server serverDelegate) isAllServicesReady() bool {
	for _, serviceName := range SERVICE_NAMES {
		if len(goworld.GetServiceProviders(serviceName)) == 0 {
			return false
		}
	}
	return true
}

func (server serverDelegate) onAllServicesReady() {
	gwlog.Info("All services are ready!")
	if goworld.GetServerID() == 1 {
		goworld.CreateSpaceAnywhere()
	}
}
