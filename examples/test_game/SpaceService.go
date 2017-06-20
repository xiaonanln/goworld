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

	spaceKindToId   map[int]common.EntityID
	pendingRequests []enterSpaceReq
}

func (s *SpaceService) OnInit() {
	s.spaceKindToId = map[int]common.EntityID{}
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

	spaceId := s.spaceKindToId[kind]
	if !spaceId.IsNil() {
		// space already exists, tell the avatar
		s.Call(avatarId, "DoEnterSpace", kind, spaceId)
	} else {
		s.pendingRequests = append(s.pendingRequests, enterSpaceReq{
			avatarId, kind,
		})
		// create the space
		goworld.CreateSpaceAnywhere(kind)
	}
}

func (s *SpaceService) NotifySpaceLoaded_Server(loadKind int, loadSpaceID common.EntityID) {
	gwlog.Info("%s: space is loaded: kind=%d, loadSpaceID=%s", s, loadKind, loadSpaceID)
	spaceID := s.spaceKindToId[loadKind]
	if !spaceID.IsNil() {
		// duplicate space created ... can happen, solve it later ...
		gwlog.Panicf("duplicate space created: kind=%d, spaceID=%s", loadKind, loadSpaceID)
	}

	s.spaceKindToId[loadKind] = spaceID
	// notify all pending requests
	leftPendingReqs := []enterSpaceReq{}
	satisfyingReqs := []enterSpaceReq{}
	for _, req := range s.pendingRequests {
		if req.kind == loadKind {
			// this req can be satisfied
			satisfyingReqs = append(satisfyingReqs, req)
		} else {
			// this req can not be satisfied
			leftPendingReqs = append(leftPendingReqs, req)
		}
	}

	if len(satisfyingReqs) > 0 {
		// if some req is satisfied
		s.pendingRequests = leftPendingReqs
		for _, req := range satisfyingReqs {
			s.Call(req.avatarId, "DoEnterSpace", loadKind, loadSpaceID)
		}
	}
}
