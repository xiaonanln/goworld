package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
)

// serverDelegate 定义一些游戏服务器的回调函数
type serverDelegate struct {
	game.GameDelegate
}

func main() {
	goworld.RegisterSpace(&MySpace{}) // 注册自定义的Space类型

	// 注册Account类型
	goworld.RegisterEntity("Account", &Account{}, false, false)
	// 注册Avatar类型，并定义属性
	goworld.RegisterEntity("Avatar", &Avatar{}, true, true).DefineAttrs(map[string][]string{
		"name":     {"Client", "Persistent"},
		"chatroom": {"Client"},
	})

	// 运行游戏服务器
	goworld.Run(&serverDelegate{})
}
