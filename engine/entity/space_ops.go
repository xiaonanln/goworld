package entity

import (
	"strconv"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/uuid"
)

// CreateSpaceLocally creates a space in the local game server
func CreateSpaceLocally(kind int) *Space {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	e := createEntity(_SPACE_ENTITY_TYPE, nil, Vector3{}, "", map[string]interface{}{
		_SPACE_KIND_ATTR_KEY: kind,
	})
	return e.AsSpace()
}

// CreateSpaceSomewhere creates a space in any game server
func CreateSpaceSomewhere(gameid uint16, kind int) common.EntityID {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return createEntitySomewhere(gameid, _SPACE_ENTITY_TYPE, map[string]interface{}{
		_SPACE_KIND_ATTR_KEY: kind,
	})
}

// CreateNilSpace creates the nil space
func CreateNilSpace(gameid uint16) *Space {
	spaceID := GetNilSpaceID(gameid)
	e := createEntity(_SPACE_ENTITY_TYPE, nil, Vector3{}, spaceID, map[string]interface{}{
		_SPACE_KIND_ATTR_KEY: 0,
	})
	return e.AsSpace()
}

// GetNilSpaceEntityID returns the EntityID for Nil Space on the specified game
// GoWorld uses fixed EntityID for nil spaces on each game
func GetNilSpaceID(gameid uint16) common.EntityID {
	gameidStr := strconv.Itoa(int(gameid))
	return common.EntityID(uuid.GenFixedUUID([]byte(gameidStr)))
}

func GetNilSpace() *Space {
	return nilSpace
}
