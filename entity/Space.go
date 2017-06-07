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
	entityID := createEntity(typeName, space)
	gwlog.Info("%s.createEntity %s: %s", space, typeName, entityID)

}

func (space *Space) enter(entity *Entity) {
	gwlog.Info("%s.enter <<< %s", space, entity)
}
