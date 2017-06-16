package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type Account struct {
	entity.Entity
	username string
}

func (a *Account) OnInit() {

}

func (a *Account) OnCreated() {
	gwlog.Info("%s created: client=%v", a, a.GetClient())
}

func (a *Account) Login_Client(username string, password string) {
	gwlog.Info("%s logining with username %s password %s ...", a, username, password)
	if password != "123456" {
		a.GetClient().Call("OnLogin", false)
		return
	}

	a.CallClient("OnLogin", true)

	avatarID := goworld.CreateEntityLocally("Avatar")

	a.Post(func() {
		avatar := goworld.GetEntity(avatarID)
		if avatar == nil {
			// login fail
			gwlog.Panicf("avatar %s not found", avatarID)
		}
		a.GiveClientTo(avatar)
	})
}
