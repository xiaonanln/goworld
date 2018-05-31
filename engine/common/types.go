package common

import (
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/uuid"
)

// ENTITYID_LENGTH is the length of Entity IDs
const ENTITYID_LENGTH = uuid.UUID_LENGTH

// EntityID type
type EntityID string

// IsNil returns if EntityID is nil
func (id EntityID) IsNil() bool {
	return id == ""
}

// GenEntityID generates a new EntityID
func GenEntityID() EntityID {
	return EntityID(uuid.GenUUID())
}

// MustEntityID assures a string to be EntityID
func MustEntityID(id string) EntityID {
	if len(id) != ENTITYID_LENGTH {
		gwlog.Panicf("%s of len %d is not a valid entity ID (len=%d)", id, len(id), ENTITYID_LENGTH)
	}
	return EntityID(id)
}

// ClientID type
type ClientID string

// GenClientID generates a new Client ID
func GenClientID() ClientID {
	return ClientID(uuid.GenUUID())
}

// IsNil returns if ClientID is nil
func (id ClientID) IsNil() bool {
	return id == ""
}

// CLIENTID_LENGTH is the length of Client IDs
const CLIENTID_LENGTH = uuid.UUID_LENGTH
