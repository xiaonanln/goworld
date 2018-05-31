package main

import (
	"time"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	_MAX_AVATAR_COUNT_PER_SPACE = 100
)

type enterSpaceReq struct {
	avatarId common.EntityID
	kind     int
}

type _SpaceKindInfo struct {
	spaceEntities map[common.EntityID]*_SpaceEntityInfo
}

func (ki *_SpaceKindInfo) choose() *_SpaceEntityInfo {
	var best *_SpaceEntityInfo

	for _, ei := range ki.spaceEntities {
		if ei.AvatarNum >= _MAX_AVATAR_COUNT_PER_SPACE { // space is full
			continue
		}

		if best == nil || best.AvatarNum < ei.AvatarNum { // choose the space with more avatars
			best = ei
		}
	}
	return best
}

func (ki *_SpaceKindInfo) remove(spaceID common.EntityID) {
	delete(ki.spaceEntities, spaceID)
}

type _SpaceEntityInfo struct {
	EntityID      common.EntityID
	Kind          int
	LastEnterTime time.Time
	AvatarNum     int
}

// SpaceService is the service entity for space management
type SpaceService struct {
	entity.Entity

	spaceKinds      map[int]*_SpaceKindInfo
	pendingRequests []enterSpaceReq
}

func (s *SpaceService) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

func (s *SpaceService) getSpaceKindInfo(kind int) *_SpaceKindInfo {
	ki := s.spaceKinds[kind]
	if ki == nil {
		ki = &_SpaceKindInfo{
			spaceEntities: map[common.EntityID]*_SpaceEntityInfo{},
		}
		s.spaceKinds[kind] = ki
	}
	return ki
}

func (s *SpaceService) getSpaceEntityInfo(kind int, spaceID common.EntityID) *_SpaceEntityInfo {
	kindinfo := s.getSpaceKindInfo(kind)
	return kindinfo.spaceEntities[spaceID]
}

// OnInit initializes SpaceService
func (s *SpaceService) OnInit() {
	s.spaceKinds = map[int]*_SpaceKindInfo{}
	s.pendingRequests = []enterSpaceReq{}
}

// OnCreated is called when entity is created
func (s *SpaceService) OnCreated() {
	gwlog.Infof("Registering SpaceService ...")
	s.DeclareService("SpaceService")
}

// EnterSpace is called by avatar to enter space by kind
func (s *SpaceService) EnterSpace(avatarId common.EntityID, kind int) {
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s.EnterSpace: avatar=%s, kind=%d", s, avatarId, kind)
	}

	spaceKindInfo := s.getSpaceKindInfo(kind)
	spaceInfo := spaceKindInfo.choose()
	if spaceInfo != nil {
		// space already exists, tell the avatar
		spaceInfo.LastEnterTime = time.Now()
		s.Call(avatarId, "DoEnterSpace", kind, spaceInfo.EntityID)
	} else {
		s.pendingRequests = append(s.pendingRequests, enterSpaceReq{
			avatarId, kind,
		})
		// create the space
		goworld.CreateSpaceAnywhere(kind)
	}
}

// NotifySpaceLoaded is called when space is loaded
func (s *SpaceService) NotifySpaceLoaded(loadKind int, loadSpaceID common.EntityID) {
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s: space is loaded: kind=%d, loadSpaceID=%s", s, loadKind, loadSpaceID)
	}
	spaceKindInfo := s.getSpaceKindInfo(loadKind)

	spaceKindInfo.spaceEntities[loadSpaceID] = &_SpaceEntityInfo{
		Kind:          loadKind,
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

// RequestDestroy is RPC request for Spaces to request for destroying self
func (s *SpaceService) RequestDestroy(kind int, spaceID common.EntityID) {
	if consts.DEBUG_SPACES {
		gwlog.Infof("Space %s kind %d is requesting destroy ...", spaceID, kind)
	}
	//spaceKindInfo := s.getSpaceKindInfo(kind)
	spaceInfo := s.getSpaceEntityInfo(kind, spaceID)

	if spaceInfo == nil { // You don't exists
		s.Call(spaceID, "ConfirmRequestDestroy", true)
		return
	}

	if time.Now().After(spaceInfo.LastEnterTime.Add(time.Second * 60)) {
		s.getSpaceKindInfo(kind).remove(spaceID)
		s.Call(spaceID, "ConfirmRequestDestroy", true)
		return
	}
}
