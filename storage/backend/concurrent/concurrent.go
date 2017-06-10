package concurrent

import (
	"sync"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/storage"
)

type ConcurrentEntityStorage struct {
	subStorages []storage.EntityStorage
	locks       []sync.RWMutex
}

func NewConcurrentEntityStorage(subStorages []storage.EntityStorage) storage.EntityStorage {
	locks := make([]sync.RWMutex, len(subStorages), len(subStorages))
	return &ConcurrentEntityStorage{
		subStorages: subStorages,
		locks:       locks,
	}
}

func hash(entityID common.EntityID) uint {
	var h uint = 1
	for _, c := range []byte(entityID) {
		h = h * uint(c)
	}
	return h
}

func (ss *ConcurrentEntityStorage) subStorageOf(name string, entityID common.EntityID) uint {
	return hash(entityID) % uint(len(ss.subStorages))
}

func (ss *ConcurrentEntityStorage) Write(name string, entityID common.EntityID, data interface{}) error {
	idx := ss.subStorageOf(name, entityID)
	lock := ss.locks[idx]
	lock.Lock()
	defer lock.Unlock()

	return ss.subStorages[idx].Write(name, entityID, data)
}

func (ss *ConcurrentEntityStorage) Read(name string, entityID common.EntityID) (interface{}, error) {
	idx := ss.subStorageOf(name, entityID)
	lock := ss.locks[idx]
	lock.RLock()
	defer lock.RUnlock()

	return ss.subStorages[idx].Read(name, entityID)
}
