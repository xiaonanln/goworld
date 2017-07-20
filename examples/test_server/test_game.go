package main

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
		"MailService",
	}
)

func init() {

}

type serverDelegate struct {
	game.GameDelegate
}

func main() {
	goworld.RegisterSpace(&MySpace{}) // Register the space type

	// Register each entity types
	goworld.RegisterEntity("Account", &Account{})
	goworld.RegisterEntity("OnlineService", &OnlineService{})
	goworld.RegisterEntity("SpaceService", &SpaceService{})
	goworld.RegisterEntity("MailService", &MailService{})

	// Register Monster type and define attributes
	goworld.RegisterEntity("Monster", &Monster{}).DefineAttrs(map[string][]string{
		"name": {"all_client"},
	})
	// Register Avatar type and define attributes
	goworld.RegisterEntity("Avatar", &Avatar{}).DefineAttrs(map[string][]string{
		"name":       {"all_client", "persistent"},
		"level":      {"all_client", "persistent"},
		"prof":       {"all_client", "persistent"},
		"exp":        {"client", "persistent"},
		"spaceKind":  {"persistent"},
		"lastMailID": {"persistent"},
		"mails":      {"client", "persistent"},
	})

	// Run the game server
	goworld.Run(&serverDelegate{})
}

// Called when the game server is ready
func (server serverDelegate) OnServerReady() {
	server.GameDelegate.OnServerReady()

	if goworld.GetServerID() == 1 { // Create services on just 1 server
		for _, serviceName := range SERVICE_NAMES {
			serviceName := serviceName
			goworld.ListEntityIDs(serviceName, func(eids []common.EntityID, err error) {
				gwlog.Info("Found saved %s ids: %v", serviceName, eids)

				if len(eids) == 0 {
					goworld.CreateEntityAnywhere(serviceName)
				} else {
					// already exists
					serviceID := eids[0]
					goworld.LoadEntityAnywhere(serviceName, serviceID)
				}
			})
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
			gwlog.Info("%s is not ready ...", serviceName)
			return false
		}
	}
	return true
}

func (server serverDelegate) onAllServicesReady() {
	gwlog.Info("All services are ready!")
	//if goworld.GetServerID() == 1 {
	//	goworld.CreateSpaceAnywhere(1)
	//}
}
