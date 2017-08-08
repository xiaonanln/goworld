package main

import "github.com/xiaonanln/goworld/engine/entity"

// MySpace 是一个自定义的场景类型
//
// 由于聊天服务器没有任何场景逻辑，因此这个类型也没有任何具体的代码实现
type MySpace struct {
	entity.Space // 自定义的场景类型必须继承一个引擎所提供的entity.Space类型
}
