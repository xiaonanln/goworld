package goworld

import (
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/server"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/storage"
)

func Run(delegate server.IServerDelegate) {
	server.Run(delegate)
}

func RegisterEntity(typeName string, entityPtr entity.IEntity) {
	entity.RegisterEntity(typeName, entityPtr)
}

//func createEntity(typeName string) {
//	server.createEntity(typeName)
//}

func CreateSpaceAnywhere(kind int) {
	entity.CreateSpaceAnywhere(kind)
}

func CreateSpaceLocally(kind int) {
	entity.CreateSpaceLocally(kind)
}

func CreateEntityLocally(typeName string) EntityID {
	return entity.CreateEntityLocally(typeName, nil, nil)
}

func CreateEntityAnywhere(typeName string) {
	entity.CreateEntityAnywhere(typeName)
}

func LoadEntityLocally(typeName string, entityID EntityID) {
	entity.LoadEntityLocally(typeName, entityID)
}

func LoadEntityAnywhere(typeName string, entityID EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

func GetServiceProviders(serviceName string) entity.EntityIDSet {
	return entity.GetServiceProviders(serviceName)
}

// Get all saved entity ids in storage, may take long time and block the main routine
func ListEntityIDs(typeName string, callback storage.ListCallbackFunc) {
	storage.ListEntityIDs(typeName, callback)
}

func Exists(typeName string, entityID EntityID, callback storage.ExistsCallbackFunc) {
	storage.Exists(typeName, entityID, callback)
}

func GetEntity(id EntityID) *entity.Entity {
	return entity.GetEntity(id)
}

func GetServerID() uint16 {
	return server.GetServerID()
}

func MapAttr() *entity.MapAttr {
	return entity.NewMapAttr()
}

func RegisterSpace(spacePtr entity.ISpace) {
	entity.RegisterSpace(spacePtr)
}
