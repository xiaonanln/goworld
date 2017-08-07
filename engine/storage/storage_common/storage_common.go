package storagecommon

import "github.com/xiaonanln/goworld/engine/common"

// EntityStorage defines the interface of entity storage backends
type EntityStorage interface {
	List(typeName string) ([]common.EntityID, error)
	Write(typeName string, entityID common.EntityID, data interface{}) error
	Read(typeName string, entityID common.EntityID) (interface{}, error)
	Exists(typeName string, entityID common.EntityID) (bool, error)
	Close()
	IsEOF(err error) bool
}
