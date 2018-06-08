package goworld

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/kvdb"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/storage"
)

// Export useful types
type Vector3 = entity.Vector3
type Entity = entity.Entity
type Space = entity.Space
type EntityID = common.EntityID

// Run runs the server endless loop
//
// This is the main routine for the server and all entity logic,
// and this function never quit
func Run() {
	game.Run()
}

// RegisterEntity registers the entity type so that entities can be created or loaded
//
// returns the entity type description object which can be used to define more properties
// of entity type
func RegisterEntity(typeName string, entityPtr entity.IEntity) *entity.EntityTypeDesc {
	return entity.RegisterEntity(typeName, entityPtr)
}

// CreateSpaceAnywhere creates a space with specified kind in any game server
func CreateSpaceAnywhere(kind int) EntityID {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceAnywhere(kind)
}

// CreateSpaceLocally creates a space with specified kind in the local game server
//
// returns the space EntityID
func CreateSpaceLocally(kind int) EntityID {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceLocally(kind)
}

// CreateEntityLocally creates a entity on the local server
//
// returns EntityID
func CreateEntityLocally(typeName string) EntityID {
	return entity.CreateEntityLocally(typeName, nil, nil)
}

// CreateEntityAnywhere creates a entity on any server
func CreateEntityAnywhere(typeName string) EntityID {
	return entity.CreateEntityAnywhere(typeName)
}

// LoadEntityAnywhere loads the specified entity from entity storage
func LoadEntityAnywhere(typeName string, entityID EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

// GetServiceProviders get the set of EntityIDs that provides the specified service
func GetServiceProviders(serviceName string) entity.EntityIDSet {
	return entity.GetServiceProviders(serviceName)
}

// ListEntityIDs gets all saved entity ids in storage, may take long time and block the main routine
//
// returns result in callback
func ListEntityIDs(typeName string, callback storage.ListCallbackFunc) {
	storage.ListEntityIDs(typeName, callback)
}

// Exists checks if entityID exists in entity storage
//
// returns result in callback
func Exists(typeName string, entityID EntityID, callback storage.ExistsCallbackFunc) {
	storage.Exists(typeName, entityID, callback)
}

// GetEntity gets the entity by EntityID
func GetEntity(id EntityID) *Entity {
	return entity.GetEntity(id)
}

// GetSpace gets the space by ID
func GetSpace(id EntityID) *Space {
	return entity.GetSpace(id)
}

// GetGameID gets the local server ID
//
// server ID is a uint16 number starts from 1, which should be different for each servers
// server ID is also in the game config section name of goworld.ini
func GetGameID() uint16 {
	return game.GetGameID()
}

// MapAttr creates a new MapAttr
func MapAttr() *entity.MapAttr {
	return entity.NewMapAttr()
}

// ListAttr creates a new ListAttr
func ListAttr() *entity.ListAttr {
	return entity.NewListAttr()
}

// RegisterSpace registers the space entity type.
//
// All spaces will be created as an instance of this type
func RegisterSpace(spacePtr entity.ISpace) {
	entity.RegisterSpace(spacePtr)
}

// Entities gets all entities as an EntityMap (do not modify it!)
func Entities() entity.EntityMap {
	return entity.Entities()
}

// Call other entities
func Call(id EntityID, method string, args ...interface{}) {
	entity.Call(id, method, args)
}

// CallService calls a service provider
func CallService(serviceName string, method string, args ...interface{}) {
	entity.CallService(serviceName, method, args)
}

// CallNilSpaces calls methods of all nil spaces on all games
func CallNilSpaces(method string, args ...interface{}) {
	entity.CallNilSpaces(method, args, game.GetGameID())
}

// GetNilSpaceID returns the Entity ID of nil space on the specified game
func GetNilSpaceID(gameid uint16) EntityID {
	return entity.GetNilSpaceID(gameid)
}

// GetKVDB gets value of key from KVDB
func GetKVDB(key string, callback kvdb.KVDBGetCallback) {
	kvdb.Get(key, callback)
}

// PutKVDB puts key-value to KVDB
func PutKVDB(key string, val string, callback kvdb.KVDBPutCallback) {
	kvdb.Put(key, val, callback)
}

// GetOrPut gets value of key from KVDB, if val not exists or is "", put key-value to KVDB.
func GetOrPutKVDB(key string, val string, callback kvdb.KVDBGetOrPutCallback) {
	kvdb.GetOrPut(key, val, callback)
}

// ListGameIDs returns all game IDs
func ListGameIDs() []uint16 {
	return config.GetGameIDs()
}

// AddTimer adds a timer to be executed after specified duration
func AddCallback(d time.Duration, callback func()) {
	timer.AddCallback(d, callback)
}

// AddTimer adds a repeat timer to be executed every specified duration
func AddTimer(d time.Duration, callback func()) {
	timer.AddTimer(d, callback)
}

// Post posts a callback to be executed
// It is almost same as AddCallback(0, callback)
func Post(callback post.PostCallback) {
	post.Post(callback)
}
