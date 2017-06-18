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
}

func (a *Avatar) setDefaultAttrs() {
	gwlog.Info("%s set default attrs: %v", a, a.Attrs.ToMap())
	a.Attrs.SetDefault("name", "无名")
	a.Attrs.SetDefault("level", 1)
	a.Attrs.SetDefault("exp", 0)
	a.Attrs.SetDefault("spaceno", DEFAULT_SPACE_ID)
}

func (a *Avatar) OnEnterSpace() {
	a.Entity.OnEnterSpace()
}

func (a *Avatar) IsPersistent() bool {
	return true
}

func (a *Avatar) enterSpace(spaceno int) {
	if a.Space.Kind == spaceno {
		return
	}
	gwlog.Info("%s enter space from %d => %d", a, a.Space.Kind, spaceno)
	a.Call(goworld.GetServiceProviders("SpaceService")[0], "EnterSpace", a.ID, spaceno)
}

func (a *Avatar) OnClientConnected() {
	a.Attrs.Set("exp", a.Attrs.GetInt("exp")+1)
	a.Attrs.Set("testpop", 1)
	v := a.Attrs.Pop("testpop")
	gwlog.Info("Avatar pop testpop => %v", v)

	a.Attrs.Set("subattr", goworld.MapAttr())
	subattr := a.Attrs.GetMapAttr("subattr")
	subattr.Set("a", 1)
	subattr.Set("b", 1)
	subattr = a.Attrs.PopMapAttr("subattr")
	a.Attrs.Set("subattr", subattr)
}

func (a *Avatar) OnClientDisconnected() {
	gwlog.Info("%s client disconnected", a)
}
