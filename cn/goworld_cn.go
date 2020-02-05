/*
GoWorld是一个分布式的游戏服务器引擎，理论上支持无限横向扩展。
一个GoWorld服务器由三种不同的进程注册：dispatcher、gate、game。
gate负责接受客户端连接并对通信数据进行压缩和加密。
game负责所有的游戏逻辑。
dispatcher作为game和gate之间的数据转发中心，负责将数据发送到正确的game或gate。

gate和diapatcher是直接编译运行的，开发者不需要自己编写任何代码。
game本质上是一个库，开发者需要自己提供main函数并调用合适的goworld模块方法启动game。

game采用一种场景（space）和（entity）的方式进行逻辑开发和通信。
当客户端登录到goworld服务器之后，就会在任意一个game上创建一个Account对象。这个Account对象就负责处理所有的客户端请求，即成为了一个ClientOwner。
一般来说，Account对象负责玩家的登录逻辑。当玩家登录成功的时候，Account就创建一个Player对象并将客户端移交（GiveClientTo）Player。
开发者应该根据游戏逻辑创建相应的space并让Player进入这些space。
space提供一种房间，场景的逻辑抽象。space被创建就永远常驻在某个game上直到被销毁，space无法迁移。
entity（上述Account，Player都是entity）则可以在space之间进行迁移。entity可以通过EnterSpace调用进入场景，如果这个场景在其他game上，goworld就会将entity的所有属性数据都打包并发送到目标game，然后在目标game上重建这个entity。这个过程对开发者来说是无缝透明的。
同一个space里的所有entity都在同一个game，因此可以直接相互调用。不同space中的entity很可能在不同的game上，因此只能通过rpc相互调用。

goworld在逻辑开发的时候使用一直单线程事件触发的方式进行开发。game只在主线程（单个goroutine）运行游戏逻辑。
因此任何游戏逻辑都不能调用任何堵塞的系统调用（例如time.Sleep）。单线程的逻辑开发可以大幅度简化逻辑代码的复杂度，因为任何逻辑和数据结构都不需要考虑并发和加锁。


*goworld*库为开发者提供大部分的GoWorld服务器引擎接口。例如goworld模块提供注册entity，创建space，创建entity等核心功能。
开发者可以参考现有的服务器例子代码来学习如何初始化并启动game。
*/
package goworld

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/kvdb"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/service"
	"github.com/xiaonanln/goworld/engine/storage"
)

const (
	// ENTITYID_LENGTH 是EntityID的长度，目前固定为16字节。
	ENTITYID_LENGTH = common.ENTITYID_LENGTH
)

// EntityID 唯一代表一个Entity。EntityID是一个字符串（string），长度固定为ENTITYID_LENGTH。
// EntityID是全局唯一的。不同进程上产生的EntityID都是唯一的，不会出现重复。因此即使是不同的游戏服务器产生的EntityID也是唯一的。
type EntityID = common.EntityID

// GameID 是Game进程的ID（数字，uint16）
// GoWorld要求GameID的数值必须是从1~N的连续N个数字，其中N为服务器配置文件中配置的game进程数目。
type GameID = uint16

// GateID 是Gate进程的ID（数字，uint16）
// GoWorld要求GateID的数值必须是从1~N的连续N个数字，其中N为服务器配置文件中配置的game进程数目。
type GateID = uint16

// DispatcherID 是Dispatcher进程的ID（数字，uint16）
// GoWorld要求DispatcherID的数值必须是从1~N的连续N个数字，其中N为服务器配置文件中配置的dispatcher进程数目
type DispatcherID uint16

// Entity 类型代表游戏服务器中的一个对象。开发者可以使用GoWorld提供的接口进行对象创建、载入。
// 同一个game进程中的Entity之间可以拿到相互的引用（指针）并直接进行相关的函数调用。不同game进程中的Entity之间可以通过EntityID使用RPC进行相互通信。
// Entity通过属性机制进行定时存盘和客户端数据同步。一般来说，开发者为Entity定义各种不同类型不同特性的属性，PERSISTENT的属性将会被定时存入到数据库中
// Client和AllClients属性能够被自动同步到客户端。当entity在game之间迁移（切换场景）的时候，所有的属性将会被导出并发送到目标game从而在目标game上重建entity对象。
// 游戏中的所有对象都必须继承Entity对象（添加一个goworld.Entity类型的匿名成员）
type Entity = entity.Entity

// Space 类型代表一个游戏服务器中的一个场景。一个场景中可以包含多个Entity。Space和其中的Entity都存在于一个game进程中。
// Entity可以通过调用EnterSpace函数来切换Space。如果EnterSpace调用所指定的Space在其他game进程上，Entity将被迁移到对应的game进程并添加到Space中。
// 游戏中的场景对象必须继承Space对象（添加一个goworld.Space类型的匿名成员）
type Space = entity.Space

// SpaceKind 类型表示Space的种类。开发者在创建Space的时候需要提供kind参数，从而创建特定SpaceKind的Space。
// NilSpace的Kind总是为0，并且开发者不能创建Kind=0的Space。
// 开发者可以根据Kind的值来区分不同的场景，具体的区分规则由开发者自己决定。
type SpaceKind = int

// Vector3 是服务端用于存储Entity位置的类型，包含X, Y, Z三个字段。
// GoWorld使用X轴和Z轴坐标进行AOI管理，无视Y轴坐标值。
type Vector3 = entity.Vector3

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
func RegisterService(typeName string, entityPtr entity.IEntity, shardCount int) {
	service.RegisterService(typeName, entityPtr, shardCount)
}

// Run 开始运行game服务。开发者需要为自己的游戏服务器提供一个main模块和main函数，并在main函数里正确初始化GoWorld服务器并启动服务器。
// 一般来说，开发者需要在main函数中注册相应的Space类型、Service类型、Entity类型，然后调用 goworld.Run() 启动GoWorld服务器即可，可参考：
// https://github.com/xiaonanln/goworld/blob/master/examples/unity_demo/unity_demo.go
func Run() {
	game.Run()
}

// CreateSpaceAnywhere 在一个随机选择的game（以后会支持自动负载均衡）上创建一个特定SpaceKind的Space对象。
// 返回所创建space的EntityID（space本质上也是entity），entity可以调用EnterSpace进入这个space。
func CreateSpaceAnywhere(kind SpaceKind) EntityID {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceSomewhere(0, kind)
}

// CreateSpaceAnywhere 做制定的game上创建一个特定SpaceKind的Space对象。
// 返回所创建space的EntityID（space本质上也是entity），entity可以调用EnterSpace进入这个space。
func CreateSpaceOnGame(gameid uint16, kind int) EntityID {
	return entity.CreateSpaceSomewhere(gameid, kind)
}

// CreateSpaceLocally 在本地game进程上创建一个指定Kind的Space。
// 返回对应的Space对象。
func CreateSpaceLocally(kind SpaceKind) *Space {
	if kind == 0 {
		gwlog.Panicf("Can not create nil space with kind=0. Game will create 1 nil space automatically.")
	}
	return entity.CreateSpaceLocally(kind)
}

// CreateEntityLocally 在本地game进程上创建一个指定类型的Entity
// 返回创建的entity对象
func CreateEntityLocally(typeName string) *Entity {
	return entity.CreateEntityLocally(typeName, nil)
}

// CreateEntityAnywhere 在随机选择的game进程上创建一个特定类型的Entity
// 返回创建对象的EntityID，可以使用这个EntityID向entity进行RPC调用
func CreateEntityAnywhere(typeName string) EntityID {
	return entity.CreateEntitySomewhere(0, typeName)
}

// CreateEntityOnGame 在指定的game进程上创建一个特定类型的Entity
// 返回创建对象的EntityID，可以使用这个EntityID向entity进行RPC调用
func CreateEntityOnGame(gameid uint16, typeName string) EntityID {
	return entity.CreateEntitySomewhere(gameid, typeName)
}

// LoadEntityAnywhere 在随机选择的game进程上载入指定的Entity。
// 如果这个Entity当前已经在任意一个game上存在，则不会重复创建。
// GoWorld保证每个Entity最多只会存在于一个game进程，即只有一份实例，在根本上规避了重复创建玩家对象可能导致的各种回档等严重问题。
func LoadEntityAnywhere(typeName string, entityID EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

// LoadEntityOnGame 在指定的game进程上载入特定的Entity对象。
// 如果这个Entity当前已经做任意一个game上存在，则GoWorld不会做任何操作。
// 因此在调用LoadEntityOnGame之后并不能100%保证Entity必然存在于所指定的game进程中！
func LoadEntityOnGame(typeName string, entityID EntityID, gameid GameID) {
	entity.LoadEntityOnGame(typeName, entityID, gameid)
}

// LoadEntityLocally 在当前的game进程中载入特定的Entity对象
// 如果这个Entity当前已经存在，则GoWorld不会做任何操作。
// 因此在调用LoadEntityOnGame之后并不能严格保证Entity必然存在于当前game进程中。
// 由于载入Entity的操作是异步的，做调用本函数之后立刻调用GetEntity不能立刻找到Entity。
func LoadEntityLocally(typeName string, entityID EntityID) {
	entity.LoadEntityOnGame(typeName, entityID, GetGameID())
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

// MapAttr 创建一个新的空MapAttr属性
//
// MapAttr允许将多个属性按照Key-Value的形式组合成一个MapAttr。Key的类型总是字符串，Value的类型可以是普通的Int, Float, Bool, Str也可以是嵌套的MapAttr和ListAttr。
func MapAttr() *entity.MapAttr {
	return entity.NewMapAttr()
}

// ListAttr 创建一个新的空ListAttr属性
//
// ListAttr允许将多个属性按照数组/列表的形式组合成一个ListAttr。ListAttr中的元素类型可以是普通的Int, Float, Bool, Str也可以是嵌套的MapAttr和ListAttr。
func ListAttr() *entity.ListAttr {
	return entity.NewListAttr()
}

// Entities 返回所有的Entity对象（通过EntityMap类型返回）
//
// 返回的EntityMap只能读取，不能做任何修改
func Entities() entity.EntityMap {
	return entity.Entities()
}

// Call 函数调用指定Entity的指定方法，并传递参数。
func Call(id EntityID, method string, args ...interface{}) {
	entity.Call(id, method, args)
}

// CallServiceAny 发起一次Service调用。开发者只需要传入指定的Service名字，不需要指知道Service的EntityID或者当前在哪个game进程。
func CallServiceAny(serviceName string, method string, args ...interface{}) {
	service.CallServiceAny(serviceName, method, args)
}

// CallServiceAll 向所有Service对象发起方法调用
func CallServiceAll(serviceName string, method string, args ...interface{}) {
	service.CallServiceAll(serviceName, method, args)
}

// CallServiceShardIndex 向shard index所指定的service entity发起方法调用
func CallServiceShardIndex(serviceName string, shardIndex int, method string, args ...interface{}) {
	service.CallServiceShardIndex(serviceName, shardIndex, method, args)
}

// CallServiceShardKey 向shard key所指定的service entity发起方法调用
func CallServiceShardKey(serviceName string, shardKey string, method string, args ...interface{}) {
	service.CallServiceShardKey(serviceName, shardKey, method, args)
}

// GetServiceEntityID 返回Service对象的EntityID。这个函数可以用来确定Service对象是否已经在某个game进程上成功创建或载入。
func GetServiceEntityID(serviceName string, shardIndex int) common.EntityID {
	return service.GetServiceEntityID(serviceName, shardIndex)
}

// GetServiceShardCount 返回Service的分片数目
func GetServiceShardCount(serviceName string) int {
	return service.GetServiceShardCount(serviceName)
}

// CheckServiceEntitiesReady 返回Service的所有Entity是否创建完毕
func CheckServiceEntitiesReady(serviceName string) bool {
	return service.CheckServiceEntitiesReady(serviceName)
}

// CallNilSpaces 向所有game进程中的NilSpace发起RPC调用。
//
// 每个game在启动之后都会在本地创建一个唯一的NilSpace。
// NilSpace和普通的space的区别在于只能由系统创建，并且kind等于0。所有Entity在创建出来还没有EnterSpace之前，总是属于NilSpace。
// 每个game上的NilSpace的EntityID都是确定的，可以根据gameid直接计算出来。因此我们可以使用NilSpace的EntityID来调用EnterSpace实现Entity到目标game的迁移。
// CallNilSpaces将会在所有game上执行一次目标函数。
func CallNilSpaces(method string, args ...interface{}) {
	entity.CallNilSpaces(method, args, game.GetGameID())
}

// GetNilSpaceID 返回特定game进程中的NilSpace的EntityID。
//
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

// AddCallback 添加一个定时回调。回调将在指定时间之后触发。回调函数（callback）总是在主线程（逻辑coroutine）中运行。
func AddCallback(d time.Duration, callback func()) {
	timer.AddCallback(d, callback)
}

// AddTimer 添加一个定时触发的回调函数。在制定时间间隔之后触发第一次，以后每过指定时间触发一次。所有触发函数总是在主线程（逻辑coroutine）中执行。
func AddTimer(d time.Duration, callback func()) {
	timer.AddTimer(d, callback)
}

func Post(callback post.PostCallback) {
	post.Post(callback)
}
