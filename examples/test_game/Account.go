package main

import (
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
}
