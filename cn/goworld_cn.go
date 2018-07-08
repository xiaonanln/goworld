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
	"github.com/xiaonanln/goworld/engine/service"
	"github.com/xiaonanln/goworld/engine/storage"
)

// GoWorld是一个分布式的游戏服务器，提供

type GameID = uint16
type GateID = uint16
type DispatcherID uint16

// Export useful types
type Vector3 = entity.Vector3

// Entity 类型代表游戏服务器中的一个对象
type Entity = entity.Entity

// Space 类型代表一个游戏服务器中的一个场景。一个场景中可以包含多个Entity。Space和其中的Entity都存在于一个game进程中。
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
	return entity.RegisterEntity(typeName, entityPtr, false)
}

// RegisterService registeres an service type
// After registeration, the service entity will be created automatically on some game
func RegisterService(typeName string, entityPtr entity.IEntity) {
	service.RegisterService(typeName, entityPtr)
}

// CreateSpaceAnywhere creates a space with specified kind in any game server
func CreateSpaceAnywhere(kind int) EntityID {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceSomewhere(0, kind)
}

// CreateSpaceOnGame creates a space with specified kind on the specified game
//
// returns the space EntityID
func CreateSpaceOnGame(gameid uint16, kind int) EntityID {
	return entity.CreateSpaceSomewhere(gameid, kind)
}

// CreateSpaceLocally creates a space with specified kind in the local game server
//
// returns the space EntityID
func CreateSpaceLocally(kind int) *Space {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceLocally(kind)
}

// CreateEntityLocally creates a entity on the local server
//
// returns EntityID
func CreateEntityLocally(typeName string) *Entity {
	return entity.CreateEntityLocally(typeName, nil)
}

// CreateEntitySomewhere creates a entity on any server
func CreateEntityAnywhere(typeName string) EntityID {
	return entity.CreateEntitySomewhere(0, typeName)
}

func CreateEntityOnGame(gameid uint16, typeName string) EntityID {
	return entity.CreateEntitySomewhere(gameid, typeName)
}

// LoadEntityAnywhere loads the specified entity from entity storage
func LoadEntityAnywhere(typeName string, entityID EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

// LoadEntityOnGame loads entity in the specified game
// If the entity already exists on any server, LoadEntityOnGame will do nothing
func LoadEntityOnGame(typeName string, entityID EntityID, gameid GameID) {
	entity.LoadEntityOnGame(typeName, entityID, gameid)
}

// LoadEntityLocally load entity in the local game
// If the entity already exists on any server, LoadEntityLocally will do nothing
func LoadEntityLocally(typeName string, entityID EntityID) {
	entity.LoadEntityOnGame(typeName, entityID, GetGameID())
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

// CallService calls a service entity
func CallService(serviceName string, method string, args ...interface{}) {
	service.CallService(serviceName, method, args)
}

// GetServiceEntityID returns the entityid of the service
func GetServiceEntityID(serviceName string) common.EntityID {
	return service.GetServiceEntityID(serviceName)
}

// CallNilSpaces calls methods of all nil spaces on all games
func CallNilSpaces(method string, args ...interface{}) {
	entity.CallNilSpaces(method, args, game.GetGameID())
}

// GetNilSpaceID returns the Entity ID of nil space on the specified game
func GetNilSpaceID(gameid GameID) EntityID {
	return entity.GetNilSpaceID(gameid)
}

// GetNilSpace returns the nil space on this game
// Nil space is a special space with Kind = 0. Nil space is the default space for all created entities.
// Each game has one nil space with fixed EntityID for each game, which can be acquired by calling `GetNilSpaceID`
//
// Since nil game exists on each game with fixed EntityID, an entity can migrate to target game by calling `e.EnterSpace(GetNilSpaceID(gameid), Vector3{})`
func GetNilSpace() *Space {
	return entity.GetNilSpace()
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
