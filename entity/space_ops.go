package entity

import . "github.com/xiaonanln/goworld/common"

func CreateSpace() EntityID {
	return createEntity(SPACE_ENTITY_TYPE, nil)
}
