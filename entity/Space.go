package entity

import "github.com/xiaonanln/goworld/gwlog"

const (
	SPACE_ENTITY_TYPE = "__space__"
)

type Space struct {
	Entity
}

func init() {
	RegisterEntity(SPACE_ENTITY_TYPE, &Space{})
}

func (space *Space) OnCreated() {
	gwlog.Debug("%s.OnCreated", space)
	spaceDelegate.OnSpaceCreated(space)
}

func (space *Space) CreateEntity(typeName string) {
	createEntity(typeName, space)
}

func (space *Space) enter(entity *Entity) {
	gwlog.Info("%s.enter <<< %s", space, entity)
	entity.space = space
}

func (space *Space) leave(entity *Entity) {
	entity.space = nil
}
