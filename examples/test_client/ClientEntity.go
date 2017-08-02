package main

import (
	"fmt"

	"reflect"

	"math/rand"
	"time"

	"github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/post"
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
	owner    *ClientBot
	TypeName string
	ID       EntityID

	Attrs     ClientAttrs
	IsPlayer  bool
	destroyed bool
	timers    map[*timer.Timer]bool

	pos entity.Position
	yaw entity.Yaw

	currentThing          string
	currentThingStartTime time.Time
	currentTimeoutTimer   *timer.Timer
}

func newClientEntity(owner *ClientBot, typeName string, entityid EntityID, isPlayer bool, clientData map[string]interface{},
	x, y, z entity.Coord, yaw entity.Yaw) *ClientEntity {
	e := &ClientEntity{
		owner:    owner,
		TypeName: typeName,
		ID:       entityid,
		Attrs:    clientData,
		IsPlayer: isPlayer,
		timers:   map[*timer.Timer]bool{},
		pos:      entity.Position{x, y, z},
		yaw:      yaw,
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
		gwlog.Debug("%s.OnCreated, IsPlayer=%v", e, e.IsPlayer)
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
	gwlog.Info("Avatar created on pos %v yaw %v", e.pos, e.yaw)
}

func (e *ClientEntity) doSomethingLater() {
	randomDelay := time.Duration(rand.Int63n(int64(AVERAGE_DO_SOMETHING_INTERVAL * 2)))
	e.AddCallback(randomDelay, func() {
		e.doSomething()
	})
}

func (e *ClientEntity) AddCallback(d time.Duration, callback timer.CallbackFunc) *timer.Timer {
	var t *timer.Timer
	t = timer.AddCallback(d, func() {
		e.owner.Lock()
		defer e.owner.Unlock()

		if !e.timers[t] {
			// timer is cancelled
			return
		}

		delete(e.timers, t)
		if e.destroyed {
			return
		}

		callback()
	})
	e.timers[t] = true
	return t
}

func (e *ClientEntity) AddTimer(d time.Duration, callback timer.CallbackFunc) *timer.Timer {
	var t *timer.Timer
	t = timer.AddTimer(d, func() {
		e.owner.Lock()
		defer e.owner.Unlock()

		if !e.timers[t] {
			// timer is cancelled
			return
		}

		if e.destroyed {
			t.Cancel()
			delete(e.timers, t)
			return
		}

		callback()
	})
	e.timers[t] = true
	return t
}

func (e *ClientEntity) CancelTimer(t *timer.Timer) {
	t.Cancel()
	delete(e.timers, t)
}

type _Something struct {
	Method  string
	Weight  int
	Timeout time.Duration
}

var (
	DO_THINGS = []*_Something{
		{"DoEnterRandomSpace", 10, time.Minute},
		{"DoSendMail", 5, time.Minute},
		{"DoGetMails", 10, time.Minute},
		{"DoSayInWorldChannel", 5, time.Minute},
		{"DoSayInProfChannel", 5, time.Minute},
		{"DoTestListField", 10, time.Minute},
	}
)

func (e *ClientEntity) doSomething() {
	if e.currentThing != "" {
		gwlog.Panicf("%s can not do something while doing %s", e, e.currentThing)
	}

	thing := e.chooseThingByWeight()
	e.currentThing = thing.Method
	e.currentThingStartTime = time.Now()
	e.currentTimeoutTimer = e.AddCallback(thing.Timeout, func() {
		gwlog.Warn("[%s] %s %s TIMEOUT !!!", time.Now(), e, thing)

		e.currentThing = ""
		e.currentThingStartTime = time.Time{}
		e.currentTimeoutTimer = nil

		e.doSomethingLater()
	})

	gwlog.Debug("[%s] %s STARTS %s", e.currentThingStartTime, e, e.currentThing)
	reflect.ValueOf(e).MethodByName(thing.Method).Call(nil)
}

func (e *ClientEntity) notifyThingDone(thing string) {
	if e.currentThing == thing {
		now := time.Now()
		//gwlog.Info("[%s] %s FINISHES %s, TAKES %s", now, e, thing, now.Sub(e.currentThingStartTime))
		recordThingTime(thing, now.Sub(e.currentThingStartTime))

		e.currentThing = ""
		e.currentThingStartTime = time.Time{}
		e.currentTimeoutTimer.Cancel()
		e.currentTimeoutTimer = nil

		e.doSomethingLater()
	}
}

func (e *ClientEntity) chooseThingByWeight() *_Something {
	totalWeight := 0
	for _, t := range DO_THINGS {
		totalWeight += t.Weight
	}
	randWeight := rand.Intn(totalWeight)
	for _, t := range DO_THINGS {
		if randWeight < t.Weight {
			return t
		}
		randWeight -= t.Weight
	}
	gwlog.Panicf("never goes here")
	return nil
}

func (e *ClientEntity) DoEnterRandomSpace() {
	curSpaceKind := 0
	if e.owner.currentSpace != nil {
		curSpaceKind = e.owner.currentSpace.Kind
	}

	spaceKindMax := N / 200
	if spaceKindMax < 2 {
		spaceKindMax = 2 // use at least 2 space
	}
	spaceKind := 1 + rand.Intn(spaceKindMax)
	for spaceKind == curSpaceKind {
		if spaceKind == spaceKindMax {
			spaceKind = 1
		} else {
			spaceKind += 1
		}
	}

	e.CallServer("EnterSpace", spaceKind)
}

func (e *ClientEntity) DoSendMail() {
	neighbors := e.Neighbors()
	//gwlog.Info("Neighbors: %v", neighbors)

	receiver := e
	if len(neighbors) > 0 {
		receiver = neighbors[rand.Intn(len(neighbors))]
	}

	mail := map[string]interface{}{
		"a": 1,
		"b": "b",
		"c": false,
		"d": 1231.111,
	}

	for i := 0; i < 1000; i++ {
		mail[fmt.Sprintf("field%d", i+1)] = rand.Intn(100)
	}

	e.CallServer("SendMail", receiver.ID, mail)
}

func (e *ClientEntity) DoGetMails() {
	e.CallServer("GetMails")
}

func (e *ClientEntity) OnGetMails(ok bool) {
	e.notifyThingDone("DoGetMails")
}

func (e *ClientEntity) DoSayInWorldChannel() {
	channel := "world"
	e.CallServer("Say", channel, fmt.Sprintf("this is a message in %s channel", channel))
}

func (e *ClientEntity) DoSayInProfChannel() {
	channel := "prof"
	e.CallServer("Say", channel, fmt.Sprintf("this is a message in %s channel", channel))
}

func (e *ClientEntity) OnSay(senderID EntityID, senderName string, channel string, content string) {
	if channel == "world" && senderID == e.ID {
		//gwlog.Info("%s %s @%s: %s", senderID, senderName, channel, content)
		e.notifyThingDone("DoSayInWorldChannel")
	} else if channel == "prof" && senderID == e.ID {
		e.notifyThingDone("DoSayInProfChannel")
	}
}

func (e *ClientEntity) DoTestListField() {
	e.CallServer("TestListField")
}

func (e *ClientEntity) OnTestListField(serverList []interface{}) {
	clientList := e.Attrs["testListField"].([]interface{})
	gwlog.Debug("OnTestListField: server=%v, client=%v", serverList, clientList)
	if len(serverList) != len(clientList) {
		gwlog.Panicf("Server list size is %d, but client list size is %d", len(serverList), len(clientList))
	}

	for i, cv := range clientList {
		cv := typeconv.Int(cv)
		sv := typeconv.Int(serverList[i])
		if cv != sv {
			gwlog.Panicf("Server item is %T %v, but client item is %T %v", sv, sv, cv, cv)
		}
	}

	e.notifyThingDone("DoTestListField")
}

func (e *ClientEntity) onAccountCreated() {
	post.Post(func() {
		username := e.owner.username()
		password := e.owner.password()
		e.CallServer("Login", username, password)
	})
}

func (e *ClientEntity) CallServer(method string, args ...interface{}) {
	e.owner.CallServer(e.ID, method, args)
}

func (e *ClientEntity) applyMapAttrChange(path []interface{}, key string, val interface{}) {
	_attr, _, _ := e.findAttrByPath(path)
	attr := _attr.(map[string]interface{})
	//if _, ok := val.(map[interface{}]interface{}); ok {
	//	val = typeconv.MapStringAnything(val)
	//}
	attr[key] = val
	e.onAttrChange(path, key)
}

func (e *ClientEntity) applyMapAttrDel(path []interface{}, key string) {
	_attr, _, _ := e.findAttrByPath(path)
	attr := _attr.(map[string]interface{})
	delete(attr, key)
	e.onAttrChange(path, key)
}

func (e *ClientEntity) applyListAttrChange(path []interface{}, index int, val interface{}) {
	gwlog.Debug("applyListAttrChange: path=%v, index=%v, val=%v", path, index, val)
	_attr, _, _ := e.findAttrByPath(path)
	attr := _attr.([]interface{})
	attr[index] = val
	e.onAttrChange(path, "")
}

func (e *ClientEntity) applyListAttrAppend(path []interface{}, val interface{}) {
	gwlog.Debug("applyListAttrAppend: path=%v, val=%v, attrs=%v", path, val, e.Attrs)
	_attr, parent, pkey := e.findAttrByPath(path)
	attr := _attr.([]interface{})

	if parentmap, ok := parent.(map[string]interface{}); ok {
		parentmap[pkey.(string)] = append(attr, val)
	} else if parentlist, ok := parent.([]interface{}); ok {
		parentlist[pkey.(int64)] = append(attr, val)
	}

	e.onAttrChange(path, "")
}
func (e *ClientEntity) applyListAttrPop(path []interface{}) {
	gwlog.Debug("applyListAttrPop: path=%v", path)
	_attr, parent, pkey := e.findAttrByPath(path)
	attr := _attr.([]interface{})

	if parentmap, ok := parent.(map[string]interface{}); ok {
		parentmap[pkey.(string)] = attr[:len(attr)-1]
	} else if parentlist, ok := parent.([]interface{}); ok {
		parentlist[pkey.(int64)] = attr[:len(attr)-1]
	}

	e.onAttrChange(path, "")
}

func (e *ClientEntity) onAttrChange(path []interface{}, key string) {
	var rootkey string
	if len(path) > 0 {
		rootkey = path[len(path)-1].(string)
	} else {
		rootkey = key
	}

	callbackFuncName := "OnAttrChange_" + rootkey
	callbackMethod := reflect.ValueOf(e).MethodByName(callbackFuncName)
	if !callbackMethod.IsValid() {
		gwlog.Debug("Attribute change callback of %s is not defined (%s)", rootkey, callbackFuncName)
		return
	}
	callbackMethod.Call([]reflect.Value{}) // call the attr change callback func
}

func (entity *ClientEntity) findAttrByPath(path []interface{}) (attr interface{}, parent interface{}, pkey interface{}) {
	// note that path is reversed
	parent, pkey = nil, nil
	attr = map[string]interface{}(entity.Attrs) // root attr

	plen := len(path)
	for i := plen - 1; i >= 0; i-- {
		parent = attr
		pkey = path[i]

		if mapattr, ok := attr.(map[string]interface{}); ok {
			key := path[i].(string)
			attr = mapattr[key]
		} else if listattr, ok := attr.([]interface{}); ok {
			index := path[i].(int)
			attr = listattr[index]
		} else {
			gwlog.Panicf("Attr is neither map nor list: %T", attr)
		}
	}
	return
}

func (attrs ClientAttrs) GetInt(key string) int {
	return int(typeconv.Int(attrs[key]))
}

func (entity *ClientEntity) OnAttrChange_exp() {
	if !quiet {
		gwlog.Debug("%s: attr exp change to %d", entity, entity.Attrs.GetInt("exp"))
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
		gwlog.Debug("%s: attr testpop change to %d", entity, v)
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
		gwlog.Debug("%s: attr subattr change to %v", entity, v)
	}
}

func (entity *ClientEntity) OnLogin(ok bool) {
	gwlog.Debug("%s OnLogin %v", entity, ok)
}

func (entity *ClientEntity) OnSendMail(ok bool) {
	gwlog.Debug("%s OnSendMail %v", entity, ok)
	entity.notifyThingDone("DoSendMail")
}

func (entity *ClientEntity) Neighbors() []*ClientEntity {
	var neighbors []*ClientEntity
	for _, other := range entity.owner.entities {
		if other.TypeName == "Avatar" {
			neighbors = append(neighbors, other)
		}
	}
	return neighbors
}
