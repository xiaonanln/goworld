package main

import (
	"fmt"

	"reflect"

	"math/rand"
	"time"

	"os"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/typeconv"
)

const (
	_AVERAGE_DO_SOMETHING_INTERVAL = time.Second * 10
)

type clientAttrs map[string]interface{}

func (attrs clientAttrs) HasKey(key string) bool {
	_, ok := attrs[key]
	return ok
}

type clientEntity struct {
	owner    *ClientBot
	TypeName string
	ID       common.EntityID

	Attrs     clientAttrs
	IsPlayer  bool
	destroyed bool
	timers    map[*timer.Timer]bool

	pos entity.Vector3
	yaw entity.Yaw

	currentThing          string
	currentThingStartTime time.Time
	currentTimeoutTimer   *timer.Timer
}

func newClientEntity(owner *ClientBot, typeName string, entityid common.EntityID, isPlayer bool, clientData map[string]interface{},
	x, y, z entity.Coord, yaw entity.Yaw) *clientEntity {
	e := &clientEntity{
		owner:    owner,
		TypeName: typeName,
		ID:       entityid,
		Attrs:    clientData,
		IsPlayer: isPlayer,
		timers:   map[*timer.Timer]bool{},
		pos:      entity.Vector3{x, y, z},
		yaw:      yaw,
	}

	e.OnCreated()
	return e
}

func (e *clientEntity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *clientEntity) Destroy() {
	if e.destroyed {
		return
	}
	e.destroyed = true
}

func (e *clientEntity) OnCreated() {
	if !quiet {
		gwlog.Debugf("%s.OnCreated, IsPlayer=%v", e, e.IsPlayer)
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

func (e *clientEntity) onAvatarCreated() {
	gwlog.Infof("Avatar created on pos %v yaw %v", e.pos, e.yaw)
}

func (e *clientEntity) doSomethingLater() {
	randomDelay := time.Duration(rand.Int63n(int64(_AVERAGE_DO_SOMETHING_INTERVAL * 2)))
	e.AddCallback(randomDelay, func() {
		e.doSomething()
	})
}

func (e *clientEntity) AddCallback(d time.Duration, callback timer.CallbackFunc) *timer.Timer {
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
	if e == nil {
		gwlog.TraceError("entity is nil: %e", e)
		os.Exit(2)
	}
	e.timers[t] = true
	return t
}

func (e *clientEntity) AddTimer(d time.Duration, callback timer.CallbackFunc) *timer.Timer {
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

func (e *clientEntity) CancelTimer(t *timer.Timer) {
	t.Cancel()
	delete(e.timers, t)
}

type _Something struct {
	Method  string
	Weight  int
	Timeout time.Duration
}

var (
	_DO_THINGS = []*_Something{
		{"DoEnterRandomSpace", 20, time.Minute},
		{"DoEnterRandomNilSpace", 10, time.Minute},
		//{"DoSendMail", 5, time.Minute},
		//{"DoGetMails", 10, time.Minute},
		{"DoSayInWorldChannel", 5, time.Minute},
		{"DoSayInProfChannel", 5, time.Minute},
		{"DoTestListField", 10, time.Minute},
		//{"DoTestPublish", 1, time.Minute},
	}
)

func (e *clientEntity) doSomething() {
	if e.currentThing != "" {
		gwlog.Panicf("%s can not do something while doing %s", e, e.currentThing)
	}
	var thing *_Something
	if e.currentSpaceKind() == 0 {
		thing = _DO_THINGS[0]
	} else {
		thing = e.chooseThingByWeight()
	}

	e.currentThing = thing.Method
	e.currentThingStartTime = time.Now()
	e.currentTimeoutTimer = e.AddCallback(thing.Timeout, func() {
		gwlog.Warnf("[%s] %s %s TIMEOUT !!!", time.Now(), e, thing)

		e.currentThing = ""
		e.currentThingStartTime = time.Time{}
		e.currentTimeoutTimer = nil

		e.doSomethingLater()
	})

	gwlog.Debugf("[%s] %s STARTS %s", e.currentThingStartTime, e, e.currentThing)
	reflect.ValueOf(e).MethodByName(thing.Method).Call(nil)
}

func (e *clientEntity) notifyThingDone(thing string) {
	if e.currentThing == thing {
		now := time.Now()
		//gwlog.Infof("[%s] %s FINISHES %s, TAKES %s", now, e, thing, now.Sub(e.currentThingStartTime))
		recordThingTime(thing, now.Sub(e.currentThingStartTime))

		e.currentThing = ""
		e.currentThingStartTime = time.Time{}
		e.currentTimeoutTimer.Cancel()
		e.currentTimeoutTimer = nil

		e.doSomethingLater()
	}
}

func (e *clientEntity) chooseThingByWeight() *_Something {
	totalWeight := 0
	for _, t := range _DO_THINGS {
		totalWeight += t.Weight
	}
	randWeight := rand.Intn(totalWeight)
	for _, t := range _DO_THINGS {
		if randWeight < t.Weight {
			return t
		}
		randWeight -= t.Weight
	}
	gwlog.Panicf("never goes here")
	return nil
}

func (e *clientEntity) currentSpaceKind() int {
	curSpaceKind := 0
	if e.owner.currentSpace != nil {
		curSpaceKind = e.owner.currentSpace.Kind
	}
	return curSpaceKind
}

func (e *clientEntity) DoEnterRandomSpace() {
	curSpaceKind := e.currentSpaceKind()
	spaceKindMax := numClients / 400
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

func (e *clientEntity) DoEnterRandomNilSpace() {
	e.CallServer("EnterRandomNilSpace")
	//gameIDs := goworld.ListGameIDs()
	//idx := rand.Intn(len(gameIDs))
	//targetGameID := gameIDs[idx]
	//nilSpaceID := goworld.GetNilSpaceID(targetGameID)
}

func (e *clientEntity) OnEnterRandomNilSpace() {
	e.notifyThingDone("DoEnterRandomNilSpace")
}

func (e *clientEntity) DoSendMail() {
	neighbors := e.Neighbors()
	//gwlog.Infof("Neighbors: %v", neighbors)

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

func (e *clientEntity) DoGetMails() {
	e.CallServer("GetMails")
}

func (e *clientEntity) OnGetMails(ok bool) {
	e.notifyThingDone("DoGetMails")
}

func (e *clientEntity) DoSayInWorldChannel() {
	channel := "world"
	e.CallServer("Say", channel, fmt.Sprintf("this is a message in %s channel", channel))
}

func (e *clientEntity) DoSayInProfChannel() {
	channel := "prof"
	e.CallServer("Say", channel, fmt.Sprintf("this is a message in %s channel", channel))
}

func (e *clientEntity) OnSay(senderID common.EntityID, senderName string, channel string, content string) {
	if channel == "world" && senderID == e.ID {
		//gwlog.Infof("%s %s @%s: %s", senderID, senderName, channel, content)
		e.notifyThingDone("DoSayInWorldChannel")
	} else if channel == "prof" && senderID == e.ID {
		e.notifyThingDone("DoSayInProfChannel")
	}
}

func (e *clientEntity) DoTestListField() {
	e.CallServer("TestListField")
}

func (e *clientEntity) OnTestListField(serverList []interface{}) {
	clientList := e.Attrs["testListField"].([]interface{})
	gwlog.Debugf("OnTestListField: server=%v, client=%v", serverList, clientList)
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

func (e *clientEntity) DoTestPublish() {
	e.CallServer("TestPublish")
}

func (e *clientEntity) OnTestPublish(publisher common.EntityID, subject string, content string) {
	gwlog.Debugf("OnTestPublish: publisher=%v, subject=%v, content=%v", publisher, subject, content)
	if publisher == e.ID {
		e.notifyThingDone("DoTestPublish")
	}
}

func (e *clientEntity) onAccountCreated() {
	post.Post(func() {
		username := e.owner.username()
		password := e.owner.password()
		e.CallServer("Login", username, password)
	})
}

func (e *clientEntity) CallServer(method string, args ...interface{}) {
	e.owner.CallServer(e.ID, method, args)
}

func (e *clientEntity) applyMapAttrChange(path []interface{}, key string, val interface{}) {
	_attr, _, _ := e.findAttrByPath(path)
	attr := _attr.(map[string]interface{})
	//if _, ok := val.(map[interface{}]interface{}); ok {
	//	val = typeconv.MapStringAnything(val)
	//}
	attr[key] = val
	e.onAttrChange(path, key)
}

func (e *clientEntity) applyMapAttrDel(path []interface{}, key string) {
	_attr, _, _ := e.findAttrByPath(path)
	attr := _attr.(map[string]interface{})
	delete(attr, key)
	e.onAttrChange(path, key)
}

func (e *clientEntity) applyListAttrChange(path []interface{}, index int, val interface{}) {
	gwlog.Debugf("applyListAttrChange: path=%v, index=%v, val=%v", path, index, val)
	_attr, _, _ := e.findAttrByPath(path)
	attr := _attr.([]interface{})
	attr[index] = val
	e.onAttrChange(path, "")
}

func (e *clientEntity) applyListAttrAppend(path []interface{}, val interface{}) {
	gwlog.Debugf("applyListAttrAppend: path=%v, val=%v, attrs=%v", path, val, e.Attrs)
	_attr, parent, pkey := e.findAttrByPath(path)
	attr := _attr.([]interface{})

	if parentmap, ok := parent.(map[string]interface{}); ok {
		parentmap[pkey.(string)] = append(attr, val)
	} else if parentlist, ok := parent.([]interface{}); ok {
		parentlist[pkey.(int64)] = append(attr, val)
	}

	e.onAttrChange(path, "")
}
func (e *clientEntity) applyListAttrPop(path []interface{}) {
	gwlog.Debugf("applyListAttrPop: path=%v", path)
	_attr, parent, pkey := e.findAttrByPath(path)
	attr := _attr.([]interface{})

	if parentmap, ok := parent.(map[string]interface{}); ok {
		parentmap[pkey.(string)] = attr[:len(attr)-1]
	} else if parentlist, ok := parent.([]interface{}); ok {
		parentlist[pkey.(int64)] = attr[:len(attr)-1]
	}

	e.onAttrChange(path, "")
}

func (e *clientEntity) onAttrChange(path []interface{}, key string) {
	var rootkey string
	if len(path) > 0 {
		rootkey = path[len(path)-1].(string)
	} else {
		rootkey = key
	}

	callbackFuncName := "OnAttrChange_" + rootkey
	callbackMethod := reflect.ValueOf(e).MethodByName(callbackFuncName)
	if !callbackMethod.IsValid() {
		gwlog.Debugf("Attribute change callback of %s is not defined (%s)", rootkey, callbackFuncName)
		return
	}
	callbackMethod.Call([]reflect.Value{}) // call the attr change callback func
}

func (e *clientEntity) findAttrByPath(path []interface{}) (attr interface{}, parent interface{}, pkey interface{}) {
	// note that path is reversed
	parent, pkey = nil, nil
	attr = map[string]interface{}(e.Attrs) // root attr

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

func (attrs clientAttrs) GetInt(key string) int {
	return int(typeconv.Int(attrs[key]))
}

func (e *clientEntity) OnAttrChange_exp() {
	if !quiet {
		gwlog.Debugf("%s: attr exp change to %d", e, e.Attrs.GetInt("exp"))
	}
}

func (e *clientEntity) OnAttrChange_testpop() {
	var v int
	if e.Attrs.HasKey("testpop") {
		v = e.Attrs.GetInt("testpop")
	} else {
		v = -1
	}
	if !quiet {
		gwlog.Debugf("%s: attr testpop change to %d", e, v)
	}
}

func (e *clientEntity) OnAttrChange_subattr() {
	var v interface{}
	if e.Attrs.HasKey("subattr") {
		v = e.Attrs["subattr"]
	} else {
		v = nil
	}
	if !quiet {
		gwlog.Debugf("%s: attr subattr change to %v", e, v)
	}
}

func (e *clientEntity) OnLogin(ok bool) {
	gwlog.Debugf("%s OnLogin %v", e, ok)
}

func (e *clientEntity) OnSendMail(ok bool) {
	gwlog.Debugf("%s OnSendMail %v", e, ok)
	e.notifyThingDone("DoSendMail")
}

func (e *clientEntity) Neighbors() []*clientEntity {
	var neighbors []*clientEntity
	for _, other := range e.owner.entities {
		if other.TypeName == "Avatar" {
			neighbors = append(neighbors, other)
		}
	}
	return neighbors
}
