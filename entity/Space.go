package entity

import (
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	SPACE_ENTITY_TYPE = "__space__"
)

type Space struct {
	Entity

	entities EntitySet
}

func init() {
	RegisterEntity(SPACE_ENTITY_TYPE, &Space{})
}

func (space *Space) OnInit() {
	space.entities = EntitySet{}
}

func (space *Space) OnCreated() {
	gwlog.Debug("%s.OnCreated", space)
	space.Post(func() {
		spaceDelegate.OnSpaceCreated(space)
	})
}

func (space *Space) CreateEntity(typeName string) {
	createEntity(typeName, space, "", nil)
}

func (space *Space) LoadEntity(typeName string, entityID common.EntityID) {
	loadEntityLocally(typeName, entityID, space)
}

func (space *Space) enter(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.enter <<< %s", space, entity)
	}
	entity.space = space
	for other := range space.entities {
		entity.interest(other)
		other.interest(entity)
	}
	space.entities.Add(entity)

	entity.I.OnEnterSpace()
}

func (space *Space) leave(entity *Entity) {
	entity.space = nil
	// remove from space entities
	space.entities.Del(entity)
	for other := range space.entities {
		entity.uninterest(other)
		other.uninterest(entity)
	}
}
