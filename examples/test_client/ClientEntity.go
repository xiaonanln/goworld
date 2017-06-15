package main

import (
	"github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
)

type ClientEntity struct {
	owner    *ClientBot
	TypeName string
	ID       EntityID
}

func newClientEntity(owner *ClientBot, typeName string, entityid EntityID) *ClientEntity {
	e := &ClientEntity{
		owner:    owner,
		TypeName: typeName,
		ID:       entityid,
	}

	e.OnCreated()
	return e
}

func (e *ClientEntity) OnCreated() {
	gwlog.Info("%s.OnCreated ")
	timer.AddCallback(0, func() {
		username := e.owner.username()
		password := e.owner.password()
		e.CallServer("Login", username, password)
	})
}

func (e *ClientEntity) CallServer(method string, args ...interface{}) {
	e.owner.CallServer(e.ID, method, args)
}
