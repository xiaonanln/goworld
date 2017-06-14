package entity

import . "github.com/xiaonanln/goworld/common"

func CreateSpaceLocally() EntityID {
	return createEntity(SPACE_ENTITY_TYPE, nil, "", nil, nil)
}

func CreateSpaceAnywhere() {
	createEntityAnywhere(SPACE_ENTITY_TYPE)
}
