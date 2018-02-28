package main

import (
	"github.com/xiaonanln/goworld"
)

var (
	_SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
	}
)

func main() {
	goworld.RegisterSpace(&MySpace{}) // 注册自定义的Space类型

	goworld.RegisterEntity("OnlineService", &OnlineService{})
	goworld.RegisterEntity("SpaceService", &SpaceService{})
	// 注册Account类型
	goworld.RegisterEntity("Account", &Account{})
	// 注册Monster类型
	goworld.RegisterEntity("Monster", &Monster{})
	// 注册Avatar类型，并定义属性
	goworld.RegisterEntity("Player", &Player{})
	// 运行游戏服务器
	goworld.Run()
}
