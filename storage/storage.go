package storage

import "github.com/xiaonanln/goworld/common"

type EntityStorage interface {
	Write(typeName string, entityID common.EntityID, data interface{}) error
	Read(typeName string, entityID common.EntityID) (interface{}, error)
}
