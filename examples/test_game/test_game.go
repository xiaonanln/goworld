package main

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/ext/pubsub"
)

var (
	_SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
		"MailService",
		pubsub.ServiceName,
	}
)

func init() {

}

func main() {
	goworld.RegisterSpace(&MySpace{}) // Register the space type

	// Register each entity types
	goworld.RegisterEntity("Account", &Account{})
	goworld.RegisterEntity("AOITester", &AOITester{})
	goworld.RegisterService("OnlineService", &OnlineService{}, 3)
	goworld.RegisterService("SpaceService", &SpaceService{}, 1)
	goworld.RegisterService("MailService", &MailService{}, 1)

	pubsub.RegisterService(1)

	// Register Monster type and define attributes
	goworld.RegisterEntity("Monster", &Monster{})
	// Register Avatar type and define attributes
	goworld.RegisterEntity("Avatar", &Avatar{})

	// Run the game server
	goworld.Run()
}

func checkServerStarted() {
	ok := isAllServicesReady()
	gwlog.Infof("checkServerStarted: %v", ok)
	if ok {
		onAllServicesReady()
	} else {
		timer.AddCallback(time.Millisecond*1000, checkServerStarted)
	}
}

func isAllServicesReady() bool {
	for _, serviceName := range _SERVICE_NAMES {
		if !goworld.CheckServiceEntitiesReady(serviceName) {
			gwlog.Infof("%s entities are not ready ...", serviceName)
			return false
		}
	}
	return true
}

func onAllServicesReady() {
	gwlog.Infof("ALL SERVICES ARE READY!!!")
	goworld.CallNilSpaces("TestCallNilSpaces", 1, "abc", true, 2.3)
}
