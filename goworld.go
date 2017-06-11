package goworld

import (
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/storage"
)

func Run(delegate game.IGameDelegate) {
	game.Run(delegate)
}

func RegisterEntity(typeName string, entityPtr entity.IEntity) {
	entity.RegisterEntity(typeName, entityPtr)
}

//func createEntity(typeName string) {
//	game.createEntity(typeName)
//}

func CreateSpaceAnywhere() {
	entity.CreateSpaceAnywhere()
}

func CreateSpaceLocally() {
	entity.CreateSpaceLocally()
}

func CreateEntityLocally(typeName string) EntityID {
	return entity.CreateEntityLocally(typeName)
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

func SetSpaceDelegate(delegate entity.ISpaceDelegate) {
	entity.SetSpaceDelegate(delegate)
}

func GetServiceProviders(serviceName string) []EntityID {
	return game.GetServiceProviders(serviceName)
}

// Get all saved entity ids in storage, may take long time and block the main routine
func ListEntityIDs(typeName string) []EntityID {
	return storage.ListEntityIDs(typeName)
}
