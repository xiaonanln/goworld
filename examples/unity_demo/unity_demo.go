package main

import (
	"time"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	_SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
	}
)

// serverDelegate 定义一些游戏服务器的回调函数
type serverDelegate struct {
	game.GameDelegate
}

func main() {
	goworld.RegisterSpace(&MySpace{}) // 注册自定义的Space类型

	goworld.RegisterEntity("OnlineService", &OnlineService{}, false, false)
	goworld.RegisterEntity("SpaceService", &SpaceService{}, false, false)
	// 注册Account类型
	goworld.RegisterEntity("Account", &Account{}, false, false)
	// 注册Monster类型
	goworld.RegisterEntity("Monster", &Monster{}, false, true).DefineAttrs(map[string][]string{
		"name": {"Client"},
		"lv":   {"Client"},
	})
	// 注册Avatar类型，并定义属性
	goworld.RegisterEntity("Player", &Player{}, true, true).DefineAttrs(map[string][]string{
		"name":      {"AllClients", "Persistent"},
		"lv":        {"AllClients", "Persistent"},
		"action":    {"AllClients"},
		"spaceKind": {"Persistent"},
	})

	// 运行游戏服务器
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
