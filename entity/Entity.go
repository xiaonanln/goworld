package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/uuid"
)

const ENTITYID_LENGTH = uuid.UUID_LENGTH

type EntityID string

func GenEntityID() EntityID {
	return EntityID(uuid.GenUUID())
}

type Entity struct {
	ID       EntityID
	TypeName string
	I        IEntity
	space    *Space
}

type IEntity interface {
	OnCreated()
	OnDestroy()
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *Entity) Destroy() {
	gwlog.Info("%s.Destroy.", e)
	if e.space != nil {
		e.space.leave(e)
	}
	e.I.OnDestroy()
	entityManager.del(e.ID)
}

// Default Handlers
func (e *Entity) OnCreated() {
	gwlog.Debug("%s.OnCreated", e)
}

func (e *Entity) OnDestroy() {
}
