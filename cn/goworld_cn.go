// goworld库为开发者提供大部分的GoWorld服务器引擎接口
// GoWorld是一个分布式的游戏服务器引擎，理论上支持无限横向扩展。
// 一个GoWorld服务器由三种不同的进程注册：dispatcher、gate、game。
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

const (
	// ENTITYID_LENGTH 是EntityID的长度，目前为16
	ENTITYID_LENGTH = common.ENTITYID_LENGTH
)

// GameID 是Game进程的ID。
// GoWorld要求GameID的数值必须是从1~N的连续N个数字，其中N为服务器配置文件中配置的game进程数目。
type GameID = uint16

// GateID 是Gate进程的ID。
// GoWorld要求GateID的数值必须是从1~N的连续N个数字，其中N为服务器配置文件中配置的game进程数目。
type GateID = uint16

// DispatcherID 是Dispatcher进程的ID
// GoWorld要求DispatcherID的数值必须是从1~N的连续N个数字，其中N为服务器配置文件中配置的dispatcher进程数目
type DispatcherID uint16

// EntityID 唯一代表一个Entity。EntityID是一个字符串（string），长度固定（ENTITYID_LENGTH）。
// EntityID是全局唯一的。不同进程上产生的EntityID都是唯一的，不会出现重复。一般来说即使是不用的游戏服务器产生的EntityID也是唯一的。
type EntityID = common.EntityID

// Entity 类型代表游戏服务器中的一个对象。开发者可以使用GoWorld提供的接口进行对象创建、载入。对象载入之后，GoWorld提供定时的对象数据存盘。
// 同一个game进程中的Entity之间可以拿到相互的引用（指针）并直接进行相关的函数调用。不同game进程中的Entity之间可以使用RPC进行相互通信。
type Entity = entity.Entity

// Space 类型代表一个游戏服务器中的一个场景。一个场景中可以包含多个Entity。Space和其中的Entity都存在于一个game进程中。
// Entity可以通过调用EnterSpace函数来切换Space。如果EnterSpace调用所指定的Space在其他game进程上，Entity将被迁移到对应的game进程并添加到Space中。
type Space = entity.Space

// Kind 类型表示Space的种类。开发者在创建Space的时候需要提供Kind参数，从而创建特定Kind的Space。NilSpace的Kind总是为0，并且开发者不能创建Kind=0的Space。
// 开发者可以根据Kind的值来区分不同的场景，具体的区分规则由开发者自己决定。
type Kind = int

// Vector3 是服务端用于存储Entity位置的类型，包含X, Y, Z三个字段。
// GoWorld使用X轴和Z轴坐标进行AOI管理，无视Y轴坐标值。
type Vector3 = entity.Vector3

// Run 开始运行game服务。开发者需要为自己的游戏服务器提供一个main模块和main函数，并在main函数里正确初始化GoWorld服务器并启动服务器。
// 一般来说，开发者需要在main函数中注册相应的Space类型、Service类型、Entity类型，然后调用 goworld.Run() 启动GoWorld服务器即可，可参考：
// https://github.com/xiaonanln/goworld/blob/master/examples/unity_demo/unity_demo.go
func Run() {
	game.Run()
}

// RegisterSpace 注册一个Space对象类型。开发者必须并且只能调用这个接口一次，从而注册特定的Space类型。一个合法的Space类型必须继承goworld.Space类型。
func RegisterSpace(spacePtr entity.ISpace) {
	entity.RegisterSpace(spacePtr)
}

// RegisterEntity 注册一个对象类型到game中。所注册的对象必须是Entity类型的子类（包含一个匿名Entity字段）。
// 使用方法可以参考：https://github.com/xiaonanln/goworld/blob/master/examples/unity_demo/unity_demo.go
func RegisterEntity(typeName string, entityPtr entity.IEntity) *entity.EntityTypeDesc {
	return entity.RegisterEntity(typeName, entityPtr, false)
}

// RegisterService 注册一个Service类型到game中。Service是一种全局唯一的特殊的Entity对象。
// 每个game进程中初始化的时候都应该注册所有的Service。GoWorld服务器会在某一个game进程中自动创建或载入Service对象（取决于Service类型是否是Persistent）。
// 开发者不能手动创建Service对象。
func RegisterService(typeName string, entityPtr entity.IEntity) {
	service.RegisterService(typeName, entityPtr)
}

// CreateSpaceAnywhere 在一个随机选择的game（以后会支持自动负载均衡）上创建一个特定Kind的Space对象。
func CreateSpaceAnywhere(kind Kind) EntityID {
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

// CreateSpaceLocally 在本地game进程上创建一个指定Kind的Space。
func CreateSpaceLocally(kind Kind) *Space {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceLocally(kind)
}

// CreateEntityLocally 在本地game进程上创建一个指定类型的Entity
func CreateEntityLocally(typeName string) *Entity {
	return entity.CreateEntityLocally(typeName, nil)
}

// CreateEntityAnywhere 在随机选择的game进程上创建一个特定类型的Entity
func CreateEntityAnywhere(typeName string) EntityID {
	return entity.CreateEntitySomewhere(0, typeName)
}

func CreateEntityOnGame(gameid uint16, typeName string) EntityID {
	return entity.CreateEntitySomewhere(gameid, typeName)
}

// LoadEntityAnywhere 在随机选择的game进程上载入指定的Entity。
// GoWorld保证每个Entity最多只会存在于一个game进程，即只有一份实例。
// 如果这个Entity当前已经存在，则GoWorld不会做任何操作。
func LoadEntityAnywhere(typeName string, entityID EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

// LoadEntityOnGame 在指定的game进程上载入特定的Entity对象。
// 如果这个Entity当前已经存在，则GoWorld不会做任何操作。因此在调用LoadEntityOnGame之后并不能严格保证Entity必然存在于所指定的game进程中。
func LoadEntityOnGame(typeName string, entityID EntityID, gameid GameID) {
	entity.LoadEntityOnGame(typeName, entityID, gameid)
}

// LoadEntityLocally 在当前的game进程中载入特定的Entity对象
// 如果这个Entity当前已经存在，则GoWorld不会做任何操作。因此在调用LoadEntityOnGame之后并不能严格保证Entity必然存在于当前game进程中。
func LoadEntityLocally(typeName string, entityID EntityID) {
	entity.LoadEntityOnGame(typeName, entityID, GetGameID())
}

// ListEntityIDs 获得某个类型的所有Entity对象的EntityID列表
// （这个接口将被弃用）
func ListEntityIDs(typeName string, callback storage.ListCallbackFunc) {
	storage.ListEntityIDs(typeName, callback)
}

// Exists 检查某个特定的Entity是否存在（已创建存盘）
func Exists(typeName string, entityID EntityID, callback storage.ExistsCallbackFunc) {
	storage.Exists(typeName, entityID, callback)
}

// GetEntity 获得当前game进程中的指定EntityID的Entity对象。不存在则返回nil。
func GetEntity(id EntityID) *Entity {
	return entity.GetEntity(id)
}

// GetSpace 获得当前进程中指定EntityID的Space对象。不存在则返回nil。
func GetSpace(id EntityID) *Space {
	return entity.GetSpace(id)
}

// GetGameID 获得当前game进程的GameID
func GetGameID() GameID {
	return game.GetGameID()
}

// MapAttr 创建一个新的空MapAttr对象
func MapAttr() *entity.MapAttr {
	return entity.NewMapAttr()
}

// ListAttr 创建一个新的空ListAttr对象
func ListAttr() *entity.ListAttr {
	return entity.NewListAttr()
}

// Entities 返回所有的Entity对象（通过EntityMap类型返回）
// 此接口将被弃用
func Entities() entity.EntityMap {
	return entity.Entities()
}

// Call 函数调用指定Entity的指定方法，并传递参数。
// 如果指定的Entity在当前game进程中，则会立刻调用其方法。否则将通过RPC发送函数调用和参数到所对应的game进程中。
func Call(id EntityID, method string, args ...interface{}) {
	entity.Call(id, method, args)
}

// CallService 发起一次Service调用。开发者只需要传入指定的Service名字，不需要指知道Service的EntityID或者当前在哪个game进程。
func CallService(serviceName string, method string, args ...interface{}) {
	service.CallService(serviceName, method, args)
}

// GetServiceEntityID 返回Service对象的EntityID。这个函数可以用来确定Service对象是否已经在某个game进程上成功创建或载入。
func GetServiceEntityID(serviceName string) common.EntityID {
	return service.GetServiceEntityID(serviceName)
}

// CallNilSpaces 向所有game进程中的NilSpace发起RPC调用。
// 由于每个game进程中都有一个唯一的NilSpace，因此这个函数想每个game进程都发起了一次函数调用。
func CallNilSpaces(method string, args ...interface{}) {
	entity.CallNilSpaces(method, args, game.GetGameID())
}

// GetNilSpaceID 返回特定game进程中的NilSpace的EntityID。
// GoWorld为每个game进程中的NilSpace使用了固定的EntityID值，例如目前GoWorld实现中在game1上NilSpace的EntityID总是"AAAAAAAAAAAAAAAx"，每次重启服务器都不会变化。
func GetNilSpaceID(gameid GameID) EntityID {
	return entity.GetNilSpaceID(gameid)
}

// GetNilSpace 返回当前game进程总的NilSpace对象
func GetNilSpace() *Space {
	return entity.GetNilSpace()
}

// GetKVDB 获得KVDB中指定key的值
func GetKVDB(key string, callback kvdb.KVDBGetCallback) {
	kvdb.Get(key, callback)
}

// PutKVDB 将制定的key-value对存入到KVDB中
func PutKVDB(key string, val string, callback kvdb.KVDBPutCallback) {
	kvdb.Put(key, val, callback)
}

// GetOrPutKVDB 读取指定key所对应的value，如果key所对应的值当前为空，则存入key-value键值对
func GetOrPutKVDB(key string, val string, callback kvdb.KVDBGetOrPutCallback) {
	kvdb.GetOrPut(key, val, callback)
}

// ListGameIDs 获得所有的GameID列表
func ListGameIDs() []GameID {
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
