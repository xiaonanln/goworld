package main

import (
	"fmt"

	"reflect"

	"github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

type ClientAttrs map[string]interface{}

func (attrs ClientAttrs) HasKey(key string) bool {
	_, ok := attrs[key]
	return ok
}

type ClientEntity struct {
	owner    *ClientBot
	TypeName string
	ID       EntityID

	Attrs ClientAttrs
}

func newClientEntity(owner *ClientBot, typeName string, entityid EntityID) *ClientEntity {
	e := &ClientEntity{
		owner:    owner,
		TypeName: typeName,
		ID:       entityid,
		Attrs:    make(ClientAttrs),
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
	var rootkey string
	if len(path) > 0 {
		rootkey = path[len(path)-1]
	} else {
		rootkey = key
	}
	attr[key] = val

	callbackFuncName := "OnAttrChange_" + rootkey
	reflect.ValueOf(e).MethodByName(callbackFuncName).Call([]reflect.Value{}) // call the attr change callback func
}

func (e *ClientEntity) applyAttrDel(path []string, key string) {
	attr := e.findAttrByPath(path)
	var rootkey string
	if len(path) > 0 {
		rootkey = path[len(path)-1]
	} else {
		rootkey = key
	}

	delete(attr, key)

	callbackFuncName := "OnAttrChange_" + rootkey
	reflect.ValueOf(e).MethodByName(callbackFuncName).Call([]reflect.Value{}) // call the attr change callback func
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

func (attrs ClientAttrs) GetInt(key string) int {
	return int(typeconv.Int(attrs[key]))
}

func (entity *ClientEntity) OnAttrChange_exp() {
	gwlog.Info("%s: attr exp change to %d", entity, entity.Attrs.GetInt("exp"))
}

func (entity *ClientEntity) OnAttrChange_testpop() {
	var v int
	if entity.Attrs.HasKey("testpop") {
		v = entity.Attrs.GetInt("testpop")
	} else {
		v = -1
	}
	gwlog.Info("%s: attr testpop change to %d", entity, v)
}
