package entity

import (
	"reflect"

	. "github.com/xiaonanln/goworld/common"
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

	RegisterEntity(SPACE_ENTITY_TYPE, spacePtr.(IEntity)).DefineAttrs(map[string][]string{
		SPACE_KIND_ATTR_KEY: {"all_client"},
	})
}
