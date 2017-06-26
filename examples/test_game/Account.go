package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb"
)

type Account struct {
	entity.Entity
	username string
}

func (a *Account) OnInit() {

}

func (a *Account) OnCreated() {
	//gwlog.Info("%s created: client=%v", a, a.GetClient())
}

func (a *Account) getAvatarID(username string, callback func(entityID common.EntityID, err error)) {
	kvdb.Get(username, func(val string, err error) {
		callback(common.EntityID(val), err)
	})
}

func (a Account) setAvatarID(username string, avatarID common.EntityID) {
	kvdb.Put(username, string(avatarID), nil)
}

func (a *Account) Login_Client(username string, password string) {
	gwlog.Info("%s logining with username %s password %s ...", a, username, password)
	if password != "123456" {
		a.GetClient().Call("OnLogin", false)
		return
	}

	a.CallClient("OnLogin", true)
	a.getAvatarID(username, func(avatarID common.EntityID, err error) {
		if err != nil {
			gwlog.Panic(err)
		}

		gwlog.Info("Username %s get avatar id = %s", username, avatarID)
		if avatarID.IsNil() {
			// avatar not found, create new avatar
			avatarID = goworld.CreateEntityLocally("Avatar")
			a.setAvatarID(username, avatarID)

			avatar := goworld.GetEntity(avatarID)
			if avatar != nil {
				a.onAvatarEntityFound(avatar)
			}
		} else {
			goworld.LoadEntityAnywhere("Avatar", avatarID)
			// ask the avatar: where are you
			a.Call(avatarID, "GetSpaceID", a.ID)
		}

	})

}

func (a *Account) onAvatarEntityFound(avatar *entity.Entity) {
	a.GiveClientTo(avatar)
}
