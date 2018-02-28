package main

import (
	"strings"

	"regexp"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Avatar 对象代表一名玩家
type Avatar struct {
	entity.Entity
}

func (a *Avatar) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetPersistent(true).SetUseAOI(true)
	desc.DefineAttr("name", "Client", "Persistent")
	desc.DefineAttr("chatroom", "Client")
}

// OnCreated 在Avatar对象创建后被调用
func (a *Avatar) OnCreated() {
	a.Entity.OnCreated()
	a.setDefaultAttrs()
}

// setDefaultAttrs 设置玩家的一些默认属性
func (a *Avatar) setDefaultAttrs() {
	a.Attrs.SetDefaultStr("name", "noname")
	a.SetFilterProp("chatroom", "1")
	a.Attrs.SetStr("chatroom", "1")
}

// GetSpaceID 获得玩家的场景ID并发给调用者
func (a *Avatar) GetSpaceID(callerID common.EntityID) {
	a.Call(callerID, "OnGetAvatarSpaceID", a.ID, a.Space.ID)
}

var spaceSep *regexp.Regexp = regexp.MustCompile("\\s")

// SendChat_Client 是用来发送聊天信息的客户端RPC
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
		a.CallFilteredClients("chatroom", "=", a.GetStr("chatroom"), "OnRecvChat", a.GetStr("name"), text)
	}
}

// enterRoom 进入一个聊天室，本质上就是设置Filter属性
func (a *Avatar) enterRoom(name string) {
	gwlog.Debugf("%s enter room %s", a, name)
	a.SetFilterProp("chatroom", name)
	a.Attrs.SetStr("chatroom", name)
}
