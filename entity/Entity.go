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
}

type IEntity interface {
	OnCreated()
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *Entity) OnCreated() {
	gwlog.Debug("%s.OnCreated", e)
}
