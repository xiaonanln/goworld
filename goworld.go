package goworld

import (
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/server"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/storage"
)

// Run the server
//
// This is the main routine for the server and all entity logic,
// and this function never quit
func Run(delegate server.IServerDelegate) {
	server.Run(delegate)
}

// Register the entity type
//
// returns the entity type description object which can be used to define more properties
// of entity type
func RegisterEntity(typeName string, entityPtr entity.IEntity) *entity.EntityTypeDesc {
	return entity.RegisterEntity(typeName, entityPtr)
}

//// Register a service entity type
//func RegisterService(typeName string, entityPtr entity.IEntity) *entity.EntityTypeDesc {
//	return entity.RegisterService(typeName, entityPtr)
//}

// Create a space with specified kind in any game server
func CreateSpaceAnywhere(kind int) {
	entity.CreateSpaceAnywhere(kind)
}

// Create a space with specified kind in the local game server
//
// returns the space EntityID
func CreateSpaceLocally(kind int) EntityID {
	return entity.CreateSpaceLocally(kind)
}

// Create a entity on the local server
//
// returns EntityID
func CreateEntityLocally(typeName string) EntityID {
	return entity.CreateEntityLocally(typeName, nil, nil)
}

// Create a entity on any server
func CreateEntityAnywhere(typeName string) {
	entity.CreateEntityAnywhere(typeName)
}

// Load the specified entity from entity storage
func LoadEntityAnywhere(typeName string, entityID EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

// Get the set of EntityIDs that provides the specified service
func GetServiceProviders(serviceName string) entity.EntityIDSet {
	return entity.GetServiceProviders(serviceName)
}

// Get all saved entity ids in storage, may take long time and block the main routine
//
// returns result in callback
func ListEntityIDs(typeName string, callback storage.ListCallbackFunc) {
	storage.ListEntityIDs(typeName, callback)
}

// Check if entityID exists in entity storage
//
// returns result in callback
func Exists(typeName string, entityID EntityID, callback storage.ExistsCallbackFunc) {
	storage.Exists(typeName, entityID, callback)
}

// Get entity by EntityID
func GetEntity(id EntityID) *entity.Entity {
	return entity.GetEntity(id)
}

// Get the local server ID
//
// server ID is a uint16 number starts from 1, which should be different for each servers
// server ID is also in the game config section name of goworld.ini
func GetServerID() uint16 {
	return server.GetServerID()
}

// Creates a new MapAttr
func MapAttr() *entity.MapAttr {
	return entity.NewMapAttr()
}

// Register the space entity type
//
// All spaces will be created as an instance of this type
func RegisterSpace(spacePtr entity.ISpace) {
	entity.RegisterSpace(spacePtr)
}
