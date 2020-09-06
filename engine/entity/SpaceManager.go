package entity

import (
	"github.com/xiaonanln/goworld/engine/common"
)

var (
	spaceManager = newSpaceManager()
)

type _SpaceManager struct {
	spaces map[common.EntityID]*Space
}

func newSpaceManager() *_SpaceManager {
	return &_SpaceManager{
		spaces: map[common.EntityID]*Space{},
	}
}

func (spmgr *_SpaceManager) putSpace(space *Space) {
	spmgr.spaces[space.ID] = space
}

func (spmgr *_SpaceManager) delSpace(id common.EntityID) {
	delete(spmgr.spaces, id)
}

func (spmgr *_SpaceManager) getSpace(id common.EntityID) *Space {
	return spmgr.spaces[id]
}

// RegisterSpace registers the user custom space type
func RegisterSpace(spacePtr ISpace) {
	RegisterEntity(_SPACE_ENTITY_TYPE, spacePtr, false)
}

func GetSpace(id common.EntityID) *Space {
	return spaceManager.spaces[id]
}
