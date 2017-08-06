package main

import (
	"strings"

	"regexp"

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
	a.Attrs.Set("chatroom", "1")
}

func (a *Avatar) GetSpaceID(callerID EntityID) {
	a.Call(callerID, "OnGetAvatarSpaceID", a.ID, a.Space.ID)
}

var spaceSep *regexp.Regexp = regexp.MustCompile("\\s")

func (a *Avatar) SendChat_Client(text string) {
	text = strings.TrimSpace(text)
	if text[0] == '/' {
		// this is a command
		cmd := spaceSep.Split(text[1:], -1)
		if cmd[0] == "join" {
			a.enterRoom(cmd[1])
		} else {
			a.CallClient("ShowError", "无法识别的命令："+cmd[0])
		}
	} else {
		a.CallFitleredClients("chatroom", a.GetStr("chatroom"), "OnRecvChat", a.GetStr("name"), text)
	}
}
func (a *Avatar) enterRoom(name string) {
	gwlog.Debug("%s enter room %s", a, name)
	a.SetFilterProp("chatroom", name)
	a.Attrs.Set("chatroom", name)
}
