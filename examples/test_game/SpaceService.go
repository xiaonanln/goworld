package main

import (
	"time"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type enterSpaceReq struct {
	avatarId common.EntityID
	kind     int
}

type _SpaceKindInfo struct {
	EntityID      common.EntityID
	LastEnterTime time.Time
}

type SpaceService struct {
	entity.Entity

	spaceKinds      map[int]*_SpaceKindInfo
	pendingRequests []enterSpaceReq
}

func (s *SpaceService) OnInit() {
	s.spaceKinds = map[int]*_SpaceKindInfo{}
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

	spaceKindInfo := s.spaceKinds[kind]
	if spaceKindInfo != nil {
		// space already exists, tell the avatar
		s.Call(avatarId, "DoEnterSpace", kind, spaceKindInfo.EntityID)
		spaceKindInfo.LastEnterTime = time.Now()
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
	spaceKindInfo := s.spaceKinds[loadKind]
	if spaceKindInfo != nil {
		// duplicate space created ... can happen, solve it later ...
		gwlog.Panicf("duplicate space created: kind=%d, spaceID=%s", loadKind, loadSpaceID)
	}

	s.spaceKinds[loadKind] = &_SpaceKindInfo{
		EntityID:      loadSpaceID,
		LastEnterTime: time.Now(),
	}
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

func (s *SpaceService) RequestDestroy_Server(kind int, spaceID common.EntityID) {
	gwlog.Info("Space %s kind %d is requesting destroy ...", spaceID, kind)
	spaceKindInfo := s.spaceKinds[kind]
	if spaceKindInfo == nil || spaceKindInfo.EntityID != spaceID {
		s.Call(spaceID, "ConfirmRequestDestroy", true)
		return
	}
	if time.Now().After(spaceKindInfo.LastEnterTime.Add(time.Second * 5)) {
		delete(s.spaceKinds, kind)
		s.Call(spaceID, "ConfirmRequestDestroy", true)
		return
	}
}
