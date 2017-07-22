package main

import (
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type AvatarInfo struct {
	name  string
	level int
}

type OnlineService struct {
	entity.Entity

	avatars  map[EntityID]*AvatarInfo
	maxlevel int
}

func (s *OnlineService) OnInit() {
	s.avatars = map[EntityID]*AvatarInfo{}
}

func (s *OnlineService) OnCreated() {
	gwlog.Info("Registering OnlineService ...")
	s.DeclareService("OnlineService")
}

func (s *OnlineService) CheckIn(avatarID EntityID, name string, level int) {
	s.avatars[avatarID] = &AvatarInfo{
		name:  name,
		level: level,
	}
	if level > s.maxlevel {
		s.maxlevel = level
	}
	gwlog.Info("%s CHECK IN: %s %s %d, total online %d", s, avatarID, name, level, len(s.avatars))
}

func (s *OnlineService) CheckOut(avatarID EntityID) {
	delete(s.avatars, avatarID)
	gwlog.Info("%s CHECK OUT: %s, total online %d", s, avatarID, len(s.avatars))
}
