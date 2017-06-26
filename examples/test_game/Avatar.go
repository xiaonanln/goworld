package main

import (
	"math/rand"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type Avatar struct {
	entity.Entity
}

func (a *Avatar) OnInit() {
}

func (a *Avatar) OnCreated() {
	a.Entity.OnCreated()

	a.setDefaultAttrs()

	//gwlog.Debug("Found OnlineService: %s", onlineServiceEid)
	a.CallService("OnlineService", "CheckIn", a.ID, a.Attrs.GetStr("name"), a.Attrs.GetInt("level"))
}

func (a *Avatar) setDefaultAttrs() {
	a.Attrs.SetDefault("name", "无名")
	a.Attrs.SetDefault("level", 1)
	a.Attrs.SetDefault("exp", 0)
	a.Attrs.SetDefault("spaceKind", 1+rand.Intn(100))
}

func (a *Avatar) IsPersistent() bool {
	return true
}

func (a *Avatar) enterSpace(spaceKind int) {
	if a.Space.Kind == spaceKind {
		return
	}
	if consts.DEBUG_SPACES {
		gwlog.Info("%s enter space from %d => %d", a, a.Space.Kind, spaceKind)
	}
	a.CallService("SpaceService", "EnterSpace", a.ID, spaceKind)
}

func (a *Avatar) OnClientConnected() {
	//gwlog.Info("%s.OnClientConnected: current space = %s", a, a.Space)
	//a.Attrs.Set("exp", a.Attrs.GetInt("exp")+1)
	//a.Attrs.Set("testpop", 1)
	//v := a.Attrs.Pop("testpop")
	//gwlog.Info("Avatar pop testpop => %v", v)
	//
	//a.Attrs.Set("subattr", goworld.MapAttr())
	//subattr := a.Attrs.GetMapAttr("subattr")
	//subattr.Set("a", 1)
	//subattr.Set("b", 1)
	//subattr = a.Attrs.PopMapAttr("subattr")
	//a.Attrs.Set("subattr", subattr)

	a.enterSpace(a.GetInt("spaceKind"))
}

func (a *Avatar) OnClientDisconnected() {
	gwlog.Info("%s client disconnected", a)
	a.Destroy()
}

func (a *Avatar) EnterSpace_Client(kind int) {
	a.enterSpace(kind)
}

func (a *Avatar) DoEnterSpace_Server(kind int, spaceID common.EntityID) {
	// let the avatar enter space with spaceID
	a.EnterSpace(spaceID)
}

func (a *Avatar) OnEnterSpace() {
	if consts.DEBUG_SPACES {
		gwlog.Info("%s ENTER SPACE %s", a, a.Space)
	}
}
