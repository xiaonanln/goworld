package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type Avatar struct {
	entity.Entity
}

func (a *Avatar) OnCreated() {
	a.Entity.OnCreated()

	onlineServiceEid := goworld.GetServiceProviders("OnlineService")[0]
	gwlog.Debug("Found OnlineService: %s", onlineServiceEid)
	a.Call(onlineServiceEid, "CheckIn", a.ID)
}

func (a *Avatar) OnEnterSpace() {
	a.Entity.OnEnterSpace()

}
