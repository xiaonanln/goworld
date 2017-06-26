package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	ACCOUNTS_DB_FILE = "accounts.db"
)

var (
	db *leveldb.DB
)

func init() {
	var err error
	db, err = leveldb.OpenFile(ACCOUNTS_DB_FILE, nil)
	if err != nil {
		panic(err)
	}
	gwlog.Info("Account DB opened: %s", ACCOUNTS_DB_FILE)
}

type Account struct {
	entity.Entity
	username string
}

func (a *Account) OnInit() {

}

func (a *Account) OnCreated() {
	//gwlog.Info("%s created: client=%v", a, a.GetClient())
}

func (a *Account) getAvatarID(username string) common.EntityID {
	data, err := db.Get([]byte(username), nil)

	if err != nil {
		if err == leveldb.ErrNotFound {
			return ""
		} else {
			gwlog.Panic(err)
		}
	}
	return common.EntityID(data)
}

func (a Account) setAvatarID(username string, avatarID common.EntityID) {
	err := db.Put([]byte(username), []byte(avatarID), nil)
	if err != nil {
		gwlog.Panic(err)
	}
}

func (a *Account) Login_Client(username string, password string) {
	gwlog.Info("%s logining with username %s password %s ...", a, username, password)
	if password != "123456" {
		a.GetClient().Call("OnLogin", false)
		return
	}

	a.CallClient("OnLogin", true)

	avatarID := a.getAvatarID(username)
	gwlog.Info("Username %s get avatar id = %s", username, avatarID)
	if avatarID.IsNil() {
		// avatar not found, create new avatar
		avatarID = goworld.CreateEntityLocally("Avatar")
		a.setAvatarID(username, avatarID)
	} else {
		goworld.LoadEntityAnywhere("Avatar", avatarID)
	}

	a.Post(func() {
		avatar := goworld.GetEntity(avatarID)
		if avatar == nil {
			// login fail
			gwlog.Panicf("avatar %s not found", avatarID)
		}
		a.GiveClientTo(avatar)
	})
}
