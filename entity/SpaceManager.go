package entity

import . "github.com/xiaonanln/goworld/common"

var (
	spaceManager = newSpaceManager()
)

type SpaceManager struct {
	spaces map[EntityID]*Space
}

func newSpaceManager() *SpaceManager {
	return &SpaceManager{
		spaces: map[EntityID]*Space{},
	}
}

func (spmgr *SpaceManager) putSpace(space *Space) {
	spmgr.spaces[space.ID] = space
}

func (spmgr *SpaceManager) delSpace(id EntityID) {
	delete(spmgr.spaces, id)
}

func (spmgr *SpaceManager) getSpace(id EntityID) *Space {
	return spmgr.spaces[id]
}
