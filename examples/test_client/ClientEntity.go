package main

import (
	"fmt"

	"github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
)

type ClientEntity struct {
	owner    *ClientBot
	TypeName string
	ID       EntityID

	Attrs map[string]interface{}
}

func newClientEntity(owner *ClientBot, typeName string, entityid EntityID) *ClientEntity {
	e := &ClientEntity{
		owner:    owner,
		TypeName: typeName,
		ID:       entityid,
		Attrs:    make(map[string]interface{}),
	}

	e.OnCreated()
	return e
}

func (e *ClientEntity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}
func (e *ClientEntity) OnCreated() {
	gwlog.Info("%s.OnCreated ", e)
	if e.TypeName == "Account" {
		timer.AddCallback(0, func() {
			username := e.owner.username()
			password := e.owner.password()
			e.CallServer("Login", username, password)
		})
	}
}

func (e *ClientEntity) CallServer(method string, args ...interface{}) {
	e.owner.CallServer(e.ID, method, args)
}

func (e *ClientEntity) applyAttrChange(path []string, key string, val interface{}) {
	attr := e.findAttrByPath(path)
	attr[key] = val
}

func (entity *ClientEntity) findAttrByPath(path []string) map[string]interface{} {
	// note that path is reversed
	attr := entity.Attrs // root attr
	plen := len(path)
	for i := plen - 1; i >= 0; i-- {
		name := path[i]
		attr = attr[name].(map[string]interface{})
	}
	return attr
}
