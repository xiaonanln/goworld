package main

import (
	. "github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Avatar entity which is the player itself
type Avatar struct {
	entity.Entity // Entity type should always inherit entity.Entity
}

func (a *Avatar) OnInit() {
}

func (a *Avatar) OnCreated() {
	a.Entity.OnCreated()

	a.setDefaultAttrs()

	gwlog.Info("Avatar %s on created: client=%s", a, a.GetClient())
	//gwlog.Debug("Found OnlineService: %s", onlineServiceEid)
	//a.CallService("OnlineService", "CheckIn", a.ID, a.Attrs.GetStr("name"), a.Attrs.GetInt("level"))
}

func (a *Avatar) setDefaultAttrs() {
	a.Attrs.SetDefault("name", "noname")

	a.SetFilterProp("chatroom", "1")
}

func (a *Avatar) GetSpaceID(callerID EntityID) {
	a.Call(callerID, "OnGetAvatarSpaceID", a.ID, a.Space.ID)
}

func (a *Avatar) SendChat_Client(text string) {
	a.CallFitleredClients("chatroom", "1", "OnRecvChat", a.GetStr("name"), text)
}
