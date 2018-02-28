package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Account 是账号对象类型，用于处理注册、登录逻辑
type Account struct {
	entity.Entity // 自定义对象类型必须继承entity.Entity
	logining      bool
}

func (a *Account) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

// Register_Client 是处理玩家注册请求的RPC函数
func (a *Account) Register_Client(username string, password string) {
	gwlog.Debugf("Register %s %s", username, password)
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
			avatar.Attrs.SetStr("name", username)
			avatar.Destroy()
			goworld.PutKVDB("avatarID$"+username, string(avatarID), func(err error) {
				a.CallClient("ShowInfo", "注册成功，请点击登录")
			})
		})
	})
}

// Login_Client 是处理玩家登录请求的RPC函数
func (a *Account) Login_Client(username string, password string) {
	if a.logining {
		// logining
		gwlog.Errorf("%s is already logining", a)
		return
	}

	gwlog.Infof("%s logining with username %s password %s ...", a, username, password)
	a.logining = true
	goworld.GetKVDB("password$"+username, func(correctPassword string, err error) {
		if err != nil {
			a.logining = false
			a.CallClient("ShowError", "服务器错误："+err.Error())
			return
		}

		if correctPassword == "" {
			a.logining = false
			a.CallClient("ShowError", "账号不存在")
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

// OnGetAvatarSpaceID 是用于接收Avatar场景编号的回调函数
func (a *Account) OnGetAvatarSpaceID(avatarID common.EntityID, spaceID common.EntityID) {
	// avatar may be in the same space with account, check again
	avatar := goworld.GetEntity(avatarID)
	if avatar != nil {
		a.onAvatarEntityFound(avatar)
		return
	}

	a.Attrs.SetStr("loginAvatarID", string(avatarID))
	a.EnterSpace(spaceID, entity.Vector3{})
}

func (a *Account) onAvatarEntityFound(avatar *entity.Entity) {
	a.logining = false
	a.GiveClientTo(avatar) // 将Account的客户端移交给Avatar
}

// OnClientDisconnected 在客户端掉线或者给了Avatar后触发
func (a *Account) OnClientDisconnected() {
	a.Destroy()
}

// OnMigrateIn 在账号迁移到目标服务器的时候调用
func (a *Account) OnMigrateIn() {
	loginAvatarID := common.EntityID(a.Attrs.GetStr("loginAvatarID"))
	avatar := goworld.GetEntity(loginAvatarID)
	gwlog.Debugf("%s migrating in, attrs=%v, loginAvatarID=%s, avatar=%v, client=%s", a, a.Attrs.ToMap(), loginAvatarID, avatar, a.GetClient())

	if avatar != nil {
		a.onAvatarEntityFound(avatar)
	} else {
		// failed
		a.CallClient("ShowError", "登录失败，请重试")
		a.logining = false
	}
}
