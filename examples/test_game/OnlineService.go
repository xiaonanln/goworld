package main

import (
	. "github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type avatarInfo struct {
	name  string
	level int
}

// OnlineService is the service entity for maintain total online avatar infos
type OnlineService struct {
	entity.Entity

	avatars  map[EntityID]*avatarInfo
	maxlevel int
}

func (s *OnlineService) OnInit() {
	s.avatars = map[EntityID]*avatarInfo{}
}

func (s *OnlineService) OnCreated() {
	gwlog.Info("Registering OnlineService ...")
	s.DeclareService("OnlineService")
}

// CheckIn is called when Avatars login
func (s *OnlineService) CheckIn(avatarID EntityID, name string, level int) {
	s.avatars[avatarID] = &avatarInfo{
		name:  name,
		level: level,
	}
	if level > s.maxlevel {
		s.maxlevel = level
	}
	gwlog.Info("%s CHECK IN: %s %s %d, total online %d", s, avatarID, name, level, len(s.avatars))
}

// CheckOut is called when Avatars logout
func (s *OnlineService) CheckOut(avatarID EntityID) {
	delete(s.avatars, avatarID)
	gwlog.Info("%s CHECK OUT: %s, total online %d", s, avatarID, len(s.avatars))
}
