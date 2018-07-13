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
	// 注册Account类型
	goworld.RegisterEntity("Account", &Account{})
	// 运行游戏服务器
	goworld.Run()
}
