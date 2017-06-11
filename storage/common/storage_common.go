package storage_common

import "github.com/xiaonanln/goworld/common"

type EntityStorage interface {
	List(typeName string) ([]common.EntityID, error)
	Write(typeName string, entityID common.EntityID, data interface{}) error
	Read(typeName string, entityID common.EntityID) (interface{}, error)
}
