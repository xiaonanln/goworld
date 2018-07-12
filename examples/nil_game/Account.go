package main

import (
	"github.com/xiaonanln/goworld/engine/entity"
)

// Account 是账号对象类型，用于处理注册、登录逻辑
type Account struct {
	entity.Entity // 自定义对象类型必须继承entity.Entity
}

func (a *Account) DescribeEntityType(desc *entity.EntityTypeDesc) {
}
