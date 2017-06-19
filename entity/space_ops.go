package entity

import . "github.com/xiaonanln/goworld/common"

func CreateSpaceLocally(kind int) EntityID {
	return createEntity(SPACE_ENTITY_TYPE, nil, "", map[string]interface{}{
		SPACE_KIND_ATTR_KEY: kind,
	}, nil, false)
}

func CreateSpaceAnywhere(kind int) {
	createEntityAnywhere(SPACE_ENTITY_TYPE, map[string]interface{}{
		SPACE_KIND_ATTR_KEY: kind,
	})
}
