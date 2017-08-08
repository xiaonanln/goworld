package entity

import (
	"reflect"

	. "github.com/xiaonanln/goworld/engine/common"
)

var (
	spaceManager = newSpaceManager()
	spaceType    reflect.Type
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

func RegisterSpace(spacePtr ISpace) {
	//if spaceType == nil {
	//	gwlog.Panicf("RegisterSpace: Space already registered")
	//}
	spaceVal := reflect.Indirect(reflect.ValueOf(spacePtr))
	spaceType = spaceVal.Type()

	RegisterEntity(_SPACE_ENTITY_TYPE, spacePtr.(IEntity), false, false).DefineAttrs(map[string][]string{
		_SPACE_KIND_ATTR_KEY: {"AllClients"}, // set to AllClients so that entities in space can visit space kind
	})
}
