package entity

import (
	"reflect"

	"github.com/xiaonanln/goworld/engine/common"
)

var (
	spaceManager = newSpaceManager()
	spaceType    reflect.Type
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
func RegisterSpace(spacePtr interface{}) {
	spaceVal := reflect.Indirect(reflect.ValueOf(spacePtr))
	spaceType = spaceVal.Type()

	RegisterEntity(_SPACE_ENTITY_TYPE, spacePtr, false, false).DefineAttrs(map[string][]string{
		_SPACE_KIND_ATTR_KEY: {"AllClients"}, // set to AllClients so that entities in space can visit space kind
	})
}
