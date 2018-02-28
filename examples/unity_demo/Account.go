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
	logIn         bool
}

func (a *Account) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

// Register_Client 是处理玩家注册请求的RPC函数
func (a *Account) Register_Client(username string, password string) {
	gwlog.Debugf("Register %s %s", username, password)
	goworld.GetOrPutKVDB("password$"+username, password, func(oldVal string, err error) {
		if err != nil {
			a.CallClient("ShowError", "Server Error： "+err.Error()) // 服务器错误
			return
		}

		if oldVal == "" {

			playerID := goworld.CreateEntityLocally("Player") // 创建一个Player对象然后立刻销毁，产生一次存盘
			player := goworld.GetEntity(playerID)
			player.Attrs.SetStr("name", username)
			player.Destroy()

			goworld.PutKVDB("playerID$"+username, string(playerID), func(err error) {
				a.CallClient("ShowInfo", "Registered Successfully, please click login.") // 注册成功，请点击登录
			})
		} else {
			a.CallClient("ShowError", "Sorry, this account aready exists.") // 抱歉，这个账号已经存在
		}
	})
}

// Login_Client 是处理玩家登录请求的RPC函数
func (a *Account) Login_Client(username string, password string) {
	gwlog.Debugf("%s.Login: username=%s, password=%s", a, username, password)
	if a.logIn {
		// logining
		gwlog.Errorf("%s has already started to log in.", a)
		return
	}

	gwlog.Infof("%s started log in with username %s password %s ...", a, username, password)
	a.logIn = true
	goworld.GetKVDB("password$"+username, func(correctPassword string, err error) {
		if err != nil {
			a.logIn = false
			a.CallClient("ShowError", "Server Error： "+err.Error()) // 服务器错误
			return
		}

		if correctPassword == "" {
			a.logIn = false
			a.CallClient("ShowError", "Account does not exist.") // 账号不存在
			return
		}

		if password != correctPassword {
			a.logIn = false
			a.CallClient("ShowError", "Invalid password or username") // 密码错误
			return
		}

		goworld.GetKVDB("playerID$"+username, func(_playerID string, err error) {
			if err != nil {
				a.logIn = false
				a.CallClient("ShowError", "Server Error："+err.Error()) // 服务器错误
				return
			}
			playerID := common.EntityID(_playerID)
			goworld.LoadEntityAnywhere("Player", playerID)
			a.Call(playerID, "GetSpaceID", a.ID)
		})
	})
}

// OnGetPlayerSpaceID 是用于接收Player场景编号的回调函数
func (a *Account) OnGetPlayerSpaceID(playerID common.EntityID, spaceID common.EntityID) {
	// player may be in the same space with account, check again
	player := goworld.GetEntity(playerID)
	if player != nil {
		a.onPlayerEntityFound(player)
		return
	}

	a.Attrs.SetStr("loginPlayerID", string(playerID))
	a.EnterSpace(spaceID, entity.Vector3{})
}

func (a *Account) onPlayerEntityFound(player *entity.Entity) {
	gwlog.Infof("Player %s is found, giving client to ...", player)
	a.logIn = false
	a.GiveClientTo(player) // 将Account的客户端移交给Player
}

// OnClientDisconnected 在客户端掉线或者给了Player后触发
func (a *Account) OnClientDisconnected() {
	gwlog.Debugf("destroying %s ...", a)
	a.Destroy()
}

// OnMigrateIn 在账号迁移到目标服务器的时候调用
func (a *Account) OnMigrateIn() {
	loginPlayerID := common.EntityID(a.Attrs.GetStr("loginPlayerID"))
	player := goworld.GetEntity(loginPlayerID)
	gwlog.Debugf("%s migrating in, attrs=%v, loginPlayerID=%s, player=%v, client=%s", a, a.Attrs.ToMap(), loginPlayerID, player, a.GetClient())

	if player != nil {
		a.onPlayerEntityFound(player)
	} else {
		// failed
		a.CallClient("ShowError", "登录失败，请重试")
		a.logIn = false
	}
}
