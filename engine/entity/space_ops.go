package entity

import "github.com/xiaonanln/goworld/engine/common"

// CreateSpaceLocally creates a space in the local game server
func CreateSpaceLocally(kind int) common.EntityID {
	return createEntity(_SPACE_ENTITY_TYPE, nil, Vector3{}, "", map[string]interface{}{
		_SPACE_KIND_ATTR_KEY: kind,
	}, nil, nil, ccCreate)
}

// CreateSpaceAnywhere creates a space in any game server
func CreateSpaceAnywhere(kind int) common.EntityID {
	return createEntityAnywhere(_SPACE_ENTITY_TYPE, map[string]interface{}{
		_SPACE_KIND_ATTR_KEY: kind,
	})
}
