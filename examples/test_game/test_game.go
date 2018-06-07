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
	goworld.RegisterEntity("OnlineService", &OnlineService{})
	goworld.RegisterEntity("SpaceService", &SpaceService{})
	goworld.RegisterEntity("MailService", &MailService{})
	pubsub.RegisterService()

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
		if len(goworld.GetServiceProviders(serviceName)) == 0 {
			gwlog.Infof("%s is not ready ...", serviceName)
			return false
		}
	}
	return true
}

func onAllServicesReady() {
	gwlog.Infof("ALL SERVICES ARE READY!!!")
	goworld.CallNilSpaces("TestCallNilSpaces", 1, "abc", true, 2.3)
}
