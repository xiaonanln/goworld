package common

import (
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/uuid"
)

const ENTITYID_LENGTH = uuid.UUID_LENGTH

type EntityID string

func GenEntityID() EntityID {
	return EntityID(uuid.GenUUID())
}

func MustEntityID(id string) EntityID {
	if len(id) != ENTITYID_LENGTH {
		gwlog.Panicf("%s of len %d is not a valid entity ID (len=%d)", id, len(id), ENTITYID_LENGTH)
	}
	return EntityID(id)
}
