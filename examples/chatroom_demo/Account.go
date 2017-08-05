package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Account entity for login process
type Account struct {
	entity.Entity // Entity type should always inherit entity.Entity
	username      string
	logining      bool
}

func (a *Account) OnInit() {

}

func (a *Account) OnCreated() {
	//gwlog.Info("%s created: client=%v", a, a.GetClient())
}

func (a *Account) Register_Client(username string, password string) {
	gwlog.Debug("Register %s %s", username, password)
	goworld.GetKVDB("password$"+username, func(val string, err error) {
		if err != nil {
			a.CallClient("ShowError", "服务器错误："+err.Error())
			return
		}

		if val != "" {
			a.CallClient("ShowError", "这个账号已经存在")
			return
		}
		goworld.PutKVDB("password$"+username, password, func(err error) {
			avatarID := goworld.CreateEntityLocally("Avatar") // 创建一个Avatar对象然后立刻销毁，产生一次存盘
			avatar := goworld.GetEntity(avatarID)
			avatar.Attrs.Set("name", username)
			avatar.Destroy()
			goworld.PutKVDB("avatarID$"+username, string(avatarID), func(err error) {
				a.CallClient("ShowInfo", "注册成功，请点击登录")
			})
		})
	})
}

func (a *Account) Login_Client(username string, password string) {
	if a.logining {
		// logining
		gwlog.Error("%s is already logining", a)
		return
	}

	gwlog.Info("%s logining with username %s password %s ...", a, username, password)
	a.logining = true
	goworld.GetKVDB("password$"+username, func(correctPassword string, err error) {
		if err != nil {
			a.logining = false
			a.CallClient("ShowError", "服务器错误："+err.Error())
			return
		}

		if password != correctPassword {
			a.logining = false
			a.CallClient("ShowError", "密码错误")
			return
		}

		goworld.GetKVDB("avatarID$"+username, func(_avatarID string, err error) {
			if err != nil {
				a.logining = false
				a.CallClient("ShowError", "服务器错误："+err.Error())
				return
			}
			avatarID := common.EntityID(_avatarID)
			goworld.LoadEntityAnywhere("Avatar", avatarID)
			a.Call(avatarID, "GetSpaceID", a.ID)
		})
	})
}

func (a *Account) OnGetAvatarSpaceID(avatarID common.EntityID, spaceID common.EntityID) {
	// avatar may be in the same space with account, check again
	avatar := goworld.GetEntity(avatarID)
	if avatar != nil {
		a.onAvatarEntityFound(avatar)
		return
	}

	a.Attrs.Set("loginAvatarID", avatarID)
	a.EnterSpace(spaceID, entity.Position{})
}

func (a *Account) onAvatarEntityFound(avatar *entity.Entity) {
	a.logining = false
	a.GiveClientTo(avatar)
}

func (a *Account) OnClientDisconnected() {
	a.Destroy()
}

func (a *Account) OnMigrateIn() {
	loginAvatarID := common.EntityID(a.Attrs.GetStr("loginAvatarID"))
	avatar := goworld.GetEntity(loginAvatarID)
	gwlog.Debug("%s migrating in, attrs=%v, loginAvatarID=%s, avatar=%v, client=%s", a, a.Attrs.ToMap(), loginAvatarID, avatar, a.GetClient())

	if avatar != nil {
		a.onAvatarEntityFound(avatar)
	} else {
		// failed
		a.CallClient("ShowError", "登录失败，请重试")
		a.logining = false
	}
}

func (a *Account) OnMigrateOut() {
	gwlog.Debug("%s migrating out ...", a)
}
