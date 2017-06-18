package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type enterSpaceReq struct {
	avatarId common.EntityID
	kind     int
}

type SpaceService struct {
	entity.Entity

	spaces          map[int]common.EntityID
	pendingRequests []enterSpaceReq
}

func (s *SpaceService) OnInit() {
	s.spaces = map[int]common.EntityID{}
	s.pendingRequests = []enterSpaceReq{}
}

func (s *SpaceService) OnCreated() {
	gwlog.Info("Registering SpaceService ...")
	s.DeclareService("SpaceService")
}

func (s *SpaceService) IsPersistent() bool {
	return false
}

func (s *SpaceService) EnterSpace_Server(avatarId common.EntityID, kind int) {
	gwlog.Info("%s.EnterSpace: avatar=%s, kind=%d", s, avatarId, kind)

	spaceId := s.spaces[kind]
	if !spaceId.IsNil() {
		// space already exists, tell the avatar
		s.Call(avatarId, "OnEnterSpace", kind, spaceId)
	} else {
		s.pendingRequests = append(s.pendingRequests, enterSpaceReq{
			avatarId, kind,
		})
		// create the space
		goworld.CreateSpaceAnywhere(kind)
	}
}
