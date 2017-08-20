package main

import (
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Player 对象代表一名玩家
type Player struct {
	entity.Entity
}

// OnCreated 在Player对象创建后被调用
func (a *Player) OnCreated() {
	a.Entity.OnCreated()
	a.setDefaultAttrs()
}

// setDefaultAttrs 设置玩家的一些默认属性
func (a *Player) setDefaultAttrs() {
	a.Attrs.SetDefault("spaceKind", 1)
	a.Attrs.SetDefault("name", "noname")
	a.Attrs.SetDefault("lv", 1)
	a.Attrs.SetDefault("action", "idle")
}

// GetSpaceID 获得玩家的场景ID并发给调用者
func (a *Player) GetSpaceID(callerID common.EntityID) {
	a.Call(callerID, "OnGetPlayerSpaceID", a.ID, a.Space.ID)
}

func (p *Player) enterSpace(spaceKind int) {
	if p.Space.Kind == spaceKind {
		return
	}
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s enter space from %d => %d", p, p.Space.Kind, spaceKind)
	}
	p.CallService("SpaceService", "EnterSpace", p.ID, spaceKind)
}

// OnClientConnected is called when client is connected
func (a *Player) OnClientConnected() {
	gwlog.Infof("%s client connected", a)
	a.enterSpace(a.GetInt("spaceKind"))
}

// OnClientDisconnected is called when client is lost
func (a *Player) OnClientDisconnected() {
	gwlog.Infof("%s client disconnected", a)
	a.Destroy()
}

// EnterSpace_Client is enter space RPC for client
func (a *Player) EnterSpace_Client(kind int) {
	a.enterSpace(kind)
}

// DoEnterSpace is called by SpaceService to notify avatar entering specified space
func (a *Player) DoEnterSpace(kind int, spaceID common.EntityID) {
	// let the avatar enter space with spaceID
	a.EnterSpace(spaceID, entity.Vector3{})
}

//func (a *Player) randomPosition() entity.Vector3 {
//	minCoord, maxCoord := -400, 400
//	return entity.Vector3{
//		X: entity.Coord(minCoord + rand.Intn(maxCoord-minCoord)),
//		Y: 0,
//		Z: entity.Coord(minCoord + rand.Intn(maxCoord-minCoord)),
//	}
//}

// OnEnterSpace is called when avatar enters a space
func (a *Player) OnEnterSpace() {
	gwlog.Infof("%s ENTER SPACE %s", a, a.Space)
}

func (a *Player) SetAction_Client(action string) {
	a.Attrs.Set("action", action)
}
