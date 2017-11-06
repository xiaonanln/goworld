package main

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/ext/msgbox"
	"github.com/xiaonanln/goworld/ext/pubsub"
)

var (
	_SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
		"MailService",
		"MsgboxService",
		pubsub.ServiceName,
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
	goworld.RegisterEntity("Account", &Account{}, false, false)
	goworld.RegisterEntity("OnlineService", &OnlineService{}, false, false)
	goworld.RegisterEntity("SpaceService", &SpaceService{}, false, false)
	goworld.RegisterEntity("MailService", &MailService{}, true, false)
	pubsub.RegisterService()
	msgbox.RegisterService()

	// Register Monster type and define attributes
	goworld.RegisterEntity("Monster", &Monster{}, false, true)
	// Register Avatar type and define attributes
	goworld.RegisterEntity("Avatar", &Avatar{}, true, true)

	// Run the game server
	goworld.Run(&serverDelegate{})
}

// OnGameReady is called when the game server is ready
func (server serverDelegate) OnGameReady() {
	server.GameDelegate.OnGameReady()

	if goworld.GetGameID() == 1 { // Create services on just 1 server
		for _, serviceName := range _SERVICE_NAMES {
			serviceName := serviceName
			goworld.ListEntityIDs(serviceName, func(eids []common.EntityID, err error) {
				gwlog.Infof("Found saved %s ids: %v", serviceName, eids)

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
	gwlog.Infof("checkServerStarted: %v", ok)
	if ok {
		server.onAllServicesReady()
	} else {
		timer.AddCallback(time.Millisecond*1000, server.checkServerStarted)
	}
}

func (server serverDelegate) isAllServicesReady() bool {
	for _, serviceName := range _SERVICE_NAMES {
		if len(goworld.GetServiceProviders(serviceName)) == 0 {
			gwlog.Infof("%s is not ready ...", serviceName)
			return false
		}
	}
	return true
}

func (server serverDelegate) onAllServicesReady() {
	gwlog.Infof("All services are ready!")
}
