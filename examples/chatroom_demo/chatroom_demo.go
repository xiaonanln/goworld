package main

import (
	"github.com/xiaonanln/goworld"
)

func main() {
	goworld.RegisterSpace(&MySpace{}) // 注册自定义的Space类型

	// 注册Account类型
	goworld.RegisterEntity("Account", &Account{})
	// 注册Avatar类型，并定义属性
	goworld.RegisterEntity("Avatar", &Avatar{})

	// 运行游戏服务器
	goworld.Run()
}
