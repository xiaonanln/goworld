package main

import (
	"github.com/xiaonanln/goworld/engine/common"
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

	avatars  map[common.EntityID]*avatarInfo
	maxlevel int
}

func (s *OnlineService) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

// OnInit initialize OnlineService fields
func (s *OnlineService) OnInit() {
	s.avatars = map[common.EntityID]*avatarInfo{}
}

// OnCreated is called when OnlineService is created
func (s *OnlineService) OnCreated() {
	gwlog.Infof("Registering OnlineService ...")
	s.DeclareService("OnlineService")
}

// CheckIn is called when Avatars login
func (s *OnlineService) CheckIn(avatarID common.EntityID, name string, level int) {
	s.avatars[avatarID] = &avatarInfo{
		name:  name,
		level: level,
	}
	if level > s.maxlevel {
		s.maxlevel = level
	}
	gwlog.Infof("%s CHECK IN: %s %s %d, total online %d", s, avatarID, name, level, len(s.avatars))
}

// CheckOut is called when Avatars logout
func (s *OnlineService) CheckOut(avatarID common.EntityID) {
	delete(s.avatars, avatarID)
	gwlog.Infof("%s CHECK OUT: %s, total online %d", s, avatarID, len(s.avatars))
}
