package common

import (
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/uuid"
)

const ENTITYID_LENGTH = uuid.UUID_LENGTH

type EntityID string

func (id EntityID) IsNil() bool {
	return id == ""
}

func GenEntityID() EntityID {
	return EntityID(uuid.GenUUID())
}

func MustEntityID(id string) EntityID {
	if len(id) != ENTITYID_LENGTH {
		gwlog.Panicf("%s of len %d is not a valid entity ID (len=%d)", id, len(id), ENTITYID_LENGTH)
	}
	return EntityID(id)
}

type ClientID string

func GenClientID() ClientID {
	return ClientID(uuid.GenUUID())
}

func (id ClientID) IsNil() bool {
	return id == ""
}

const CLIENTID_LENGTH = uuid.UUID_LENGTH

type MapData map[string]interface{}
type ListData []interface{}
