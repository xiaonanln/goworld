package main

import (
	"math/rand"

	"fmt"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type Avatar struct {
	entity.Entity
	name  string
	level int
	money int
}

func (a *Avatar) OnInit() {
	a.name = fmt.Sprintf("Avatar%d", rand.Intn(100))
	a.level = 1 + rand.Intn(100)
	a.money = rand.Intn(10000)
}

func (a *Avatar) OnCreated() {
	a.Entity.OnCreated()

	onlineServiceEid := goworld.GetServiceProviders("OnlineService")[0]
	gwlog.Debug("Found OnlineService: %s", onlineServiceEid)
	a.Call(onlineServiceEid, "CheckIn", a.ID, a.name, a.level)
}

func (a *Avatar) OnEnterSpace() {
	a.Entity.OnEnterSpace()

}
