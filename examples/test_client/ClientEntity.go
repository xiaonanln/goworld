package main

import (
	"fmt"

	"reflect"

	"math/rand"
	"time"

	"sync"

	"github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

const (
	AVERAGE_DO_SOMETHING_INTERVAL = time.Second * 15
)

type ClientAttrs map[string]interface{}

func (attrs ClientAttrs) HasKey(key string) bool {
	_, ok := attrs[key]
	return ok
}

type ClientEntity struct {
	sync.Mutex

	owner    *ClientBot
	TypeName string
	ID       EntityID

	Attrs     ClientAttrs
	IsPlayer  bool
	destroyed bool
}

func newClientEntity(owner *ClientBot, typeName string, entityid EntityID, isPlayer bool) *ClientEntity {
	e := &ClientEntity{
		owner:    owner,
		TypeName: typeName,
		ID:       entityid,
		Attrs:    make(ClientAttrs),
		IsPlayer: isPlayer,
	}

	e.OnCreated()
	return e
}

func (e *ClientEntity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *ClientEntity) Destroy() {
	if e.destroyed {
		return
	}
	e.destroyed = true
}

func (e *ClientEntity) OnCreated() {
	if !quiet {
		gwlog.Info("%s.OnCreated, IsPlayer=%v", e, e.IsPlayer)
	}
	if !e.IsPlayer {
		return
	}

	if e.TypeName == "Avatar" {
		e.onAvatarCreated()
	} else if e.TypeName == "Account" {
		e.onAccountCreated()
	}

}

func (e *ClientEntity) onAvatarCreated() {
	e.doSomethingLater()
}

func (e *ClientEntity) doSomethingLater() {
	randomDelay := time.Duration(rand.Int63n(int64(AVERAGE_DO_SOMETHING_INTERVAL * 2)))
	timer.AddCallback(randomDelay, func() {
		e.Lock()
		defer e.Unlock()

		if e.destroyed {
			return
		}
		e.doSomething()
		e.doSomethingLater()
	})
}

type _Something struct {
	Method string
	Weight int
}

var (
	DO_THINGS = []_Something{
		{"DoEnterRandomSpace", 100},
	}
)

func (e *ClientEntity) doSomething() {
	thing := e.chooseThingByWeight()
	reflect.ValueOf(e).MethodByName(thing).Call(nil)
}

func (e *ClientEntity) chooseThingByWeight() string {
	return "DoEnterRandomSpace"
}

func (e *ClientEntity) DoEnterRandomSpace() {
	spaceKind := SPACE_KIND_MIN + rand.Intn(SPACE_KIND_MAX-SPACE_KIND_MIN+1)
	e.CallServer("EnterSpace", spaceKind)
}

func (e *ClientEntity) onAccountCreated() {
	timer.AddCallback(0, func() {

		e.Lock()
		defer e.Unlock()

		username := e.owner.username()
		password := e.owner.password()
		e.CallServer("Login", username, password)
	})
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

	if _, ok := val.(map[interface{}]interface{}); ok {
		val = typeconv.MapStringAnything(val)
	}
	attr[key] = val

	callbackFuncName := "OnAttrChange_" + rootkey
	callbackMethod := reflect.ValueOf(e).MethodByName(callbackFuncName)
	if !callbackMethod.IsValid() {
		gwlog.Warn("Attribute change callback of %s is not defined (%s)", rootkey, callbackFuncName)
		return
	}
	callbackMethod.Call([]reflect.Value{}) // call the attr change callback func
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
	callbackMethod := reflect.ValueOf(e).MethodByName(callbackFuncName)
	if !callbackMethod.IsValid() {
		gwlog.Warn("Attribute change callback of %s is not defined (%s)", rootkey, callbackFuncName)
		return
	}
	callbackMethod.Call([]reflect.Value{}) // call the attr change callback func
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
	if !quiet {
		gwlog.Info("%s: attr exp change to %d", entity, entity.Attrs.GetInt("exp"))
	}
}

func (entity *ClientEntity) OnAttrChange_testpop() {
	var v int
	if entity.Attrs.HasKey("testpop") {
		v = entity.Attrs.GetInt("testpop")
	} else {
		v = -1
	}
	if !quiet {
		gwlog.Info("%s: attr testpop change to %d", entity, v)
	}
}

func (entity *ClientEntity) OnAttrChange_subattr() {
	var v interface{}
	if entity.Attrs.HasKey("subattr") {
		v = entity.Attrs["subattr"]
	} else {
		v = nil
	}
	if !quiet {
		gwlog.Info("%s: attr subattr change to %v", entity, v)
	}
}
