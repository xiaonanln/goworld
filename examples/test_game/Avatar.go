package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	DEFAULT_SPACE_ID = 1
)

type Avatar struct {
	entity.Entity
}

func (a *Avatar) OnInit() {
}

func (a *Avatar) OnCreated() {
	a.Entity.OnCreated()

	a.setDefaultAttrs()

	onlineServiceEid := goworld.GetServiceProviders("OnlineService")[0]
	gwlog.Debug("Found OnlineService: %s", onlineServiceEid)
	a.Call(onlineServiceEid, "CheckIn", a.ID, a.Attrs.GetStr("name"), a.Attrs.GetInt("level"))

	a.enterSpace(a.Attrs.GetInt("spaceId"))
}
func (a *Avatar) setDefaultAttrs() {
	gwlog.Info("%s set default attrs: %s", a, a.Attrs.ToMap())
	a.Attrs.SetDefault("name", "无名")
	a.Attrs.SetDefault("level", 1)
	a.Attrs.SetDefault("exp", 0)
	a.Attrs.SetDefault("spaceId", DEFAULT_SPACE_ID)
}

func (a *Avatar) OnEnterSpace() {
	a.Entity.OnEnterSpace()
}

func (a *Avatar) IsPersistent() bool {
	return true
}

func (a *Avatar) enterSpace(spaceId int) {
	//curspace := a.GetSpace()
	//if curspace.Attrs.GetInt("spaceId") == spaceId {
	//
	//}
}
