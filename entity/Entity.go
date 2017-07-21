package entity

import (
	"fmt"
	"reflect"

	"time"

	timer "github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"

	"unsafe"

	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/post"
	"github.com/xiaonanln/goworld/storage"
	"github.com/xiaonanln/typeconv"
)

var (
	saveInterval time.Duration
)

type Yaw float32

type entityTimerInfo struct {
	FireTime       time.Time
	RepeatInterval time.Duration
	Method         string
	Args           []interface{}
	Repeat         bool
	rawTimer       *timer.Timer
}

type Entity struct {
	ID       EntityID
	TypeName string
	I        IEntity
	IV       reflect.Value

	destroyed bool
	typeDesc  *EntityTypeDesc
	Space     *Space
	aoi       AOI
	yaw       Yaw

	rawTimers   map[*timer.Timer]struct{}
	timers      map[EntityTimerID]*entityTimerInfo
	lastTimerId EntityTimerID

	client           *GameClient
	declaredServices StringSet
	becamePlayer     bool

	Attrs *MapAttr

	enteringSpaceRequest struct {
		SpaceID     EntityID
		EnterPos    Position
		RequestTime int64
	}
	filterProps map[string]string
}

// Functions declared by IEntity can be override in Entity subclasses
type IEntity interface {
	// Entity Lifetime
	OnInit()    // Called when initializing entity struct, override to initialize entity custom fields
	OnCreated() // Called when entity is just created
	OnDestroy() // Called when entity is destroying (just before destroy)
	// Migration
	OnMigrateOut() // Called just before entity is migrating out
	OnMigrateIn()  // Called just after entity is migrating in
	// Space Operations
	OnEnterSpace()             // Called when entity leaves space
	OnLeaveSpace(space *Space) // Called when entity enters space
	// Storage: Save & Load
	IsPersistent() bool                             // Return whether entity is persistent, override to return true for persistent entity
	GetPersistentData() map[string]interface{}      // Convert persistent entity attributes to persistent data for storage, can override to customize entity saving
	LoadPersistentData(data map[string]interface{}) // Initialize entity attributes with persistetn data, can override to customize entity loading
	GetMigrateData() map[string]interface{}         // Convert entity attributes for migrating to other servers, can override to customize data migrating
	LoadMigrateData(data map[string]interface{})    // Initialize attributes with migrating data, can override to customize data migrating
	// Client Notifications
	OnClientConnected()    // Called when client is connected to entity (become player)
	OnClientDisconnected() // Called when client disconnected
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *Entity) Destroy() {
	if e.destroyed {
		return
	}
	gwlog.Debug("%s.Destroy ...", e)
	e.destroyEntity(false)
	dispatcher_client.GetDispatcherClientForSend().SendNotifyDestroyEntity(e.ID)
}

func (e *Entity) destroyEntity(isMigrate bool) {
	e.Space.leave(e)

	if !isMigrate {
		gwutils.RunPanicless(e.I.OnDestroy)
	} else {
		gwutils.RunPanicless(e.I.OnMigrateOut)
	}

	e.clearRawTimers()
	e.rawTimers = nil // prohibit further use

	if !isMigrate {
		e.SetClient(nil) // always set client to nil before destroy
		e.Save()
	} else {
		if e.client != nil {
			entityManager.onEntityLoseClient(e.client.clientid)
			e.client = nil
		}
	}

	entityManager.del(e.ID)
	e.destroyed = true
}

func (e *Entity) IsDestroyed() bool {
	return e.destroyed
}

func (e *Entity) Save() {
	if !e.I.IsPersistent() {
		return
	}

	if consts.DEBUG_SAVE_LOAD {
		gwlog.Debug("SAVING %s ...", e)
	}

	data := e.I.GetPersistentData()

	storage.Save(e.TypeName, e.ID, data, nil)
}

func (e *Entity) IsSpaceEntity() bool {
	return e.TypeName == SPACE_ENTITY_TYPE
}

// Convert entity to space (only works for space entity)
func (e *Entity) ToSpace() *Space {
	if !e.IsSpaceEntity() {
		gwlog.Panicf("%s is not a space", e)
	}

	return (*Space)(unsafe.Pointer(e))
}

func (e *Entity) init(typeName string, entityID EntityID, entityInstance reflect.Value) {
	e.ID = entityID
	e.IV = entityInstance
	e.I = entityInstance.Interface().(IEntity)
	e.TypeName = typeName

	e.typeDesc = registeredEntityTypes[typeName]

	e.rawTimers = map[*timer.Timer]struct{}{}
	e.timers = map[EntityTimerID]*entityTimerInfo{}
	e.declaredServices = StringSet{}
	e.filterProps = map[string]string{}

	attrs := NewMapAttr()
	attrs.owner = e
	e.Attrs = attrs

	initAOI(&e.aoi)
	gwutils.RunPanicless(e.I.OnInit)
}

func (e *Entity) setupSaveTimer() {
	e.addRawTimer(saveInterval, e.Save)
}

func SetSaveInterval(duration time.Duration) {
	saveInterval = duration
	gwlog.Info("Save interval set to %s", saveInterval)
}

// Space Operations related to e

// Interests and Uninterest among entities
func (e *Entity) interest(other *Entity) {
	e.aoi.interest(other)
	e.client.SendCreateEntity(other, false)
}

func (e *Entity) uninterest(other *Entity) {
	e.aoi.uninterest(other)
	e.client.SendDestroyEntity(other)
}

func (e *Entity) Neighbors() EntitySet {
	return e.aoi.neighbors
}

// Timer & Callback Management
type EntityTimerID int

func (tid EntityTimerID) IsValid() bool {
	return tid > 0
}

func (e *Entity) AddCallback(d time.Duration, method string, args ...interface{}) EntityTimerID {
	tid := e.genTimerId()
	now := time.Now()
	info := &entityTimerInfo{
		FireTime: now.Add(d),
		Method:   method,
		Args:     args,
		Repeat:   false,
	}
	e.timers[tid] = info
	info.rawTimer = e.addRawCallback(d, func() {
		e.triggerTimer(tid, false)
	})
	gwlog.Debug("%s.AddCallback %s: %d", e, method, tid)
	return tid
}

func (e *Entity) AddTimer(d time.Duration, method string, args ...interface{}) EntityTimerID {
	if d < time.Millisecond*10 { // minimal interval for repeat timer
		d = time.Millisecond * 10
	}

	tid := e.genTimerId()
	now := time.Now()
	info := &entityTimerInfo{
		FireTime:       now.Add(d),
		RepeatInterval: d,
		Method:         method,
		Args:           args,
		Repeat:         true,
	}
	e.timers[tid] = info
	info.rawTimer = e.addRawTimer(d, func() {
		e.triggerTimer(tid, true)
	})
	gwlog.Debug("%s.AddTimer %s: %d", e, method, tid)
	return tid
}

func (e *Entity) CancelTimer(tid EntityTimerID) {
	timerInfo := e.timers[tid]
	if timerInfo == nil {
		return // timer already fired or cancelled
	}
	delete(e.timers, tid)
	e.cancelRawTimer(timerInfo.rawTimer)
}

func (e *Entity) triggerTimer(tid EntityTimerID, isRepeat bool) {
	timerInfo := e.timers[tid] // should never be nil
	gwlog.Debug("%s trigger timer %d: %v", e, tid, timerInfo)
	if !timerInfo.Repeat {
		delete(e.timers, tid)
	} else {
		if !isRepeat {
			timerInfo.rawTimer = e.addRawTimer(timerInfo.RepeatInterval, func() {
				e.triggerTimer(tid, true)
			})
		}

		now := time.Now()
		timerInfo.FireTime = now.Add(timerInfo.RepeatInterval)
	}

	e.onCallFromLocal(timerInfo.Method, timerInfo.Args)
}

func (e *Entity) genTimerId() EntityTimerID {
	e.lastTimerId += 1
	tid := e.lastTimerId
	return tid
}

var timersPacker = netutil.MessagePackMsgPacker{}

func (e *Entity) dumpTimers() ([]byte, error) {
	timers := make([]*entityTimerInfo, 0, len(e.timers))
	for _, t := range e.timers {
		timers = append(timers, t)
	}
	e.timers = nil // no more AddCallback or AddTimer
	data, err := timersPacker.PackMsg(timers, nil)
	//gwlog.Info("%s dump %d timers: %v", e, len(timers), data)
	return data, err
}

func (e *Entity) restoreTimers(data []byte) error {
	var timers []*entityTimerInfo
	if err := timersPacker.UnpackMsg(data, &timers); err != nil {
		return err
	}
	gwlog.Debug("%s: %d timers restored: %v", e, len(timers), timers)
	now := time.Now()
	for _, timer := range timers {
		//if timer.rawTimer != nil {
		//	gwlog.Panicf("raw timer should be nil")
		//}

		tid := e.genTimerId()
		e.timers[tid] = timer

		timer.rawTimer = e.addRawCallback(timer.FireTime.Sub(now), func() {
			e.triggerTimer(tid, false)
		})
	}
	return nil
}

func (e *Entity) addRawCallback(d time.Duration, cb timer.CallbackFunc) *timer.Timer {
	var t *timer.Timer
	t = timer.AddCallback(d, func() {
		delete(e.rawTimers, t)
		cb()
	})
	e.rawTimers[t] = struct{}{}
	return t
}

func (e *Entity) addRawTimer(d time.Duration, cb timer.CallbackFunc) *timer.Timer {
	t := timer.AddTimer(d, cb)
	e.rawTimers[t] = struct{}{}
	return t
}

func (e *Entity) cancelRawTimer(t *timer.Timer) {
	delete(e.rawTimers, t)
	t.Cancel()
}

func (e *Entity) clearRawTimers() {
	for t := range e.rawTimers {
		t.Cancel()
	}
	e.rawTimers = map[*timer.Timer]struct{}{}
}

// Post a function which will be executed immediately but not in the current stack frames
func (e *Entity) Post(cb func()) {
	post.Post(cb)
}

// Call other entities
func (e *Entity) Call(id EntityID, method string, args ...interface{}) {
	callEntity(id, method, args)
}

func (e *Entity) CallService(serviceName string, method string, args ...interface{}) {
	serviceEid := entityManager.chooseServiceProvider(serviceName)
	callEntity(serviceEid, method, args)
}

func (e *Entity) onCallFromLocal(methodName string, args []interface{}) {
	defer func() {
		err := recover() // recover from any error during RPC call
		if err != nil {
			gwlog.TraceError("%s.%s%v paniced: %s", e, methodName, args, err)
		}
	}()

	rpcDesc := e.typeDesc.rpcDescs[methodName]
	if rpcDesc == nil {
		// rpc not found
		gwlog.Panicf("%s.onCallFromLocal: Method %s is not a valid RPC, args=%v", e, methodName, args)
	}

	// rpc call from server
	if rpcDesc.Flags&RF_SERVER == 0 {
		// can not call from server
		gwlog.Panicf("%s.onCallFromLocal: Method %s can not be called from Server: flags=%v", e, methodName, rpcDesc.Flags)
	}

	if rpcDesc.NumArgs != len(args) {
		gwlog.Panicf("%s.onCallFromLocal: Method %s receives %d arguments, but given %d: %v", e, methodName, rpcDesc.NumArgs, len(args), args)
	}

	methodType := rpcDesc.MethodType
	in := make([]reflect.Value, len(args)+1)
	in[0] = reflect.ValueOf(e.I) // first argument is the bind instance (self)

	for i, arg := range args {
		argType := methodType.In(i + 1)
		in[i+1] = typeconv.Convert(arg, argType)
	}

	rpcDesc.Func.Call(in)
}

func (e *Entity) onCallFromRemote(methodName string, args [][]byte, clientid ClientID) {
	defer func() {
		err := recover() // recover from any error during RPC call
		if err != nil {
			gwlog.TraceError("%s.%s%v paniced: %s", e, methodName, args, err)
		}
	}()

	rpcDesc := e.typeDesc.rpcDescs[methodName]
	if rpcDesc == nil {
		// rpc not found
		gwlog.Error("%s.onCallFromRemote: Method %s is not a valid RPC, args=%v", e, methodName, args)
		return
	}

	methodType := rpcDesc.MethodType
	if clientid == "" {
		// rpc call from server
		if rpcDesc.Flags&RF_SERVER == 0 {
			// can not call from server
			gwlog.Panicf("%s.onCallFromRemote: Method %s can not be called from Server: flags=%v", e, methodName, rpcDesc.Flags)
		}
	} else {
		isFromOwnClient := clientid == e.getClientID()
		if rpcDesc.Flags&RF_OWN_CLIENT == 0 && isFromOwnClient {
			gwlog.Panicf("%s.onCallFromRemote: Method %s can not be called from OwnClient: flags=%v", e, methodName, rpcDesc.Flags)
		} else if rpcDesc.Flags&RF_OTHER_CLIENT == 0 && !isFromOwnClient {
			gwlog.Panicf("%s.onCallFromRemote: Method %s can not be called from OtherClient: flags=%v, OwnClient=%s, OtherClient=%s", e, methodName, rpcDesc.Flags, e.getClientID(), clientid)
		}
	}

	if rpcDesc.NumArgs != len(args) {
		gwlog.Error("%s.onCallFromRemote: Method %s receives %d arguments, but given %d: %v", e, methodName, rpcDesc.NumArgs, len(args), args)
		return
	}

	in := make([]reflect.Value, len(args)+1)
	in[0] = reflect.ValueOf(e.I) // first argument is the bind instance (self)

	for i, arg := range args {
		argType := methodType.In(i + 1)
		argValPtr := reflect.New(argType)

		err := netutil.MSG_PACKER.UnpackMsg(arg, argValPtr.Interface())
		if err != nil {
			gwlog.Panicf("Convert argument %d failed: type=%s", i+1, argType.Name())
		}

		in[i+1] = reflect.Indirect(argValPtr)
	}

	rpcDesc.Func.Call(in)
}

// Register for global service
func (e *Entity) DeclareService(serviceName string) {
	e.declaredServices.Add(serviceName)
	dispatcher_client.GetDispatcherClientForSend().SendDeclareService(e.ID, serviceName)
}

// Default Handlers
func (e *Entity) OnInit() {
	//gwlog.Warn("%s.OnInit not implemented", e)
}

func (e *Entity) OnCreated() {
	//gwlog.Debug("%s.OnCreated", e)
}

// Space Utilities
func (e *Entity) OnEnterSpace() {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnEnterSpace >>> %s", e, e.Space)
	}
}

func (e *Entity) OnLeaveSpace(space *Space) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnLeaveSpace <<< %s", e, space)
	}
}

func (e *Entity) OnDestroy() {
}

// Default handlers for persistence

// Return if the entity is persistent
//
// Default implementation check entity for persistent attributes
func (e *Entity) IsPersistent() bool {
	return len(e.typeDesc.persistentAttrs) > 0
}

// Get the persistent data
//
// Returns persistent attributes by default
func (e *Entity) GetPersistentData() map[string]interface{} {
	return e.Attrs.ToMapWithFilter(e.typeDesc.persistentAttrs.Contains)
}

// Load persistent data
//
// Load persistent data to attributes
func (e *Entity) LoadPersistentData(data map[string]interface{}) {
	e.Attrs.AssignMap(data)
}

func (e *Entity) getClientData() map[string]interface{} {
	return e.Attrs.ToMapWithFilter(e.typeDesc.clientAttrs.Contains)
}

func (e *Entity) getAllClientData() map[string]interface{} {
	return e.Attrs.ToMapWithFilter(e.typeDesc.allClientAttrs.Contains)
}

func (e *Entity) GetMigrateData() map[string]interface{} {
	return e.Attrs.ToMap() // all attrs are migrated, without filter
}

func (e *Entity) LoadMigrateData(data map[string]interface{}) {
	e.Attrs.AssignMap(data)
}

// Client related utilities

// Get client
func (e *Entity) GetClient() *GameClient {
	return e.client
}

func (e *Entity) getClientID() ClientID {
	if e.client != nil {
		return e.client.clientid
	} else {
		return ""
	}
}

func (e *Entity) SetClient(client *GameClient) {
	oldClient := e.client
	if oldClient == client {
		return
	}

	e.client = client

	if oldClient != nil {
		// send destroy entity to client
		entityManager.onEntityLoseClient(oldClient.clientid)
		dispatcher_client.GetDispatcherClientForSend().SendClearClientFilterProp(oldClient.gateid, oldClient.clientid)

		for neighbor := range e.Neighbors() {
			oldClient.SendDestroyEntity(neighbor)
		}

		oldClient.SendDestroyEntity(e)
	}

	if client != nil {
		// send create entity to new client
		entityManager.onEntityGetClient(e.ID, client.clientid)

		client.SendCreateEntity(e, true)

		for neighbor := range e.Neighbors() {
			client.SendCreateEntity(neighbor, false)
		}

		// set all filter properties to client
		for key, val := range e.filterProps {
			dispatcher_client.GetDispatcherClientForSend().SendSetClientFilterProp(client.gateid, client.clientid, key, val)
		}
	}

	if oldClient == nil && client != nil {
		// got net client
		gwutils.RunPanicless(e.I.OnClientConnected)
	} else if oldClient != nil && client == nil {
		gwutils.RunPanicless(e.I.OnClientDisconnected)
	}
}

func (e *Entity) CallClient(method string, args ...interface{}) {
	e.client.call(e.ID, method, args...)
}

func (e *Entity) GiveClientTo(other *Entity) {
	if e.client == nil {
		gwlog.Warn("%s.GiveClientTo(%s): client is nil", e, other)
		return
	}

	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.GiveClientTo(%s): client=%s", e, other, e.client)
	}
	client := e.client
	e.SetClient(nil)

	other.SetClient(client)
}

func (e *Entity) ForAllClients(f func(client *GameClient)) {
	if e.client != nil {
		f(e.client)
	}

	for neighbor := range e.Neighbors() {
		if neighbor.client != nil {
			f(neighbor.client)
		}
	}
}

func (e *Entity) notifyClientDisconnected() {
	// called when client disconnected
	if e.client == nil {
		gwlog.Panic(e.client)
	}
	e.client = nil
	gwutils.RunPanicless(e.I.OnClientDisconnected)
}

func (e *Entity) OnClientConnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.OnClientConnected: %s, %d Neighbors", e, e.client, len(e.Neighbors()))
	}
}

func (e *Entity) OnClientDisconnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.OnClientDisconnected: %s", e, e.client)
	}
}

func (e *Entity) OnBecomePlayer() {
	gwlog.Info("%s.OnBecomePlayer: client=%s", e, e.client)
}

func (e *Entity) sendAttrChangeToClients(ma *MapAttr, key string, val interface{}) {
	path := ma.getPathFromOwner()
	e.client.SendNotifyAttrChange(e.ID, path, key, val)
}

func (e *Entity) sendAttrDelToClients(ma *MapAttr, key string) {
	path := ma.getPathFromOwner()
	e.client.SendNotifyAttrDel(e.ID, path, key)
}

// Define Attributes Properties

// Fast access to attrs
func (e *Entity) GetInt(key string) int {
	return e.Attrs.GetInt(key)
}

func (e *Entity) GetStr(key string) string {
	return e.Attrs.GetStr(key)
}

func (e *Entity) GetFloat(key string) float64 {
	return e.Attrs.GetFloat(key)
}

// Enter Space

// Enter target space
func (e *Entity) EnterSpace(spaceID EntityID, pos Position) {
	if e.isEnteringSpace() {
		gwlog.Error("%s is entering space %s, can not enter space %s", e, e.enteringSpaceRequest.SpaceID, spaceID)
		return
	}

	//e.requestMigrateTo(spaceID, pos)
	localSpace := spaceManager.getSpace(spaceID)
	if localSpace != nil { // target space is local, just enter
		e.enterLocalSpace(localSpace, pos)
	} else { // else request migrating to other space
		e.requestMigrateTo(spaceID, pos)
	}
}
func (e *Entity) enterLocalSpace(space *Space, pos Position) {
	if space == e.Space {
		// space not changed
		gwlog.TraceError("%s.enterLocalSpace: already in space %s", e, space)
		return
	}

	e.enteringSpaceRequest.SpaceID = space.ID
	e.enteringSpaceRequest.EnterPos = pos
	e.enteringSpaceRequest.RequestTime = time.Now().UnixNano()

	e.Post(func() {
		e.clearEnteringSpaceRequest()

		if space.IsDestroyed() {
			gwlog.Warn("%s: space %s is destroyed, enter space cancelled", e, space.ID)
			return
		}

		//gwlog.Info("%s.enterLocalSpace ==> %s", e, space)
		e.Space.leave(e)
		space.enter(e, pos)
	})
}

func (e *Entity) isEnteringSpace() bool {
	now := time.Now().UnixNano()
	return now < (e.enteringSpaceRequest.RequestTime + int64(consts.ENTER_SPACE_REQUEST_TIMEOUT))
}

// Migrate to the server of space
func (e *Entity) requestMigrateTo(spaceID EntityID, pos Position) {
	e.enteringSpaceRequest.SpaceID = spaceID
	e.enteringSpaceRequest.EnterPos = pos
	e.enteringSpaceRequest.RequestTime = time.Now().UnixNano()

	dispatcher_client.GetDispatcherClientForSend().SendMigrateRequest(spaceID, e.ID)
}

func (e *Entity) clearEnteringSpaceRequest() {
	e.enteringSpaceRequest.SpaceID = ""
	e.enteringSpaceRequest.EnterPos = Position{}
	e.enteringSpaceRequest.RequestTime = 0
}

func OnMigrateRequestAck(entityID EntityID, spaceID EntityID, spaceLoc uint16) {
	entity := entityManager.get(entityID)
	if entity == nil {
		//dispatcher_client.GetDispatcherClientForSend().SendCancelMigrateRequest(entityID)
		gwlog.Error("Migrate failed since entity is destroyed: spaceID=%s, entityID=%s", spaceID, entityID)
		return
	}

	if spaceLoc == 0 {
		// target space not found, migrate not started
		gwlog.Error("Migrate failed since target space is not found: spaceID=%s, entity=%s", spaceID, entity)
		entity.clearEnteringSpaceRequest()
		return
	}

	if !entity.isEnteringSpace() {
		// replay from dispatcher is too late ?
		return
	}

	if entity.enteringSpaceRequest.SpaceID != spaceID {
		// not entering this space ?
		return
	}

	entity.realMigrateTo(spaceID, entity.enteringSpaceRequest.EnterPos, spaceLoc)
}

func (e *Entity) realMigrateTo(spaceID EntityID, pos Position, spaceLoc uint16) {
	var clientid ClientID
	var clientsrv uint16
	if e.client != nil {
		clientid = e.client.clientid
		clientsrv = e.client.gateid
	}

	e.destroyEntity(true) // disable the entity
	timerData, err := e.dumpTimers()
	if err != nil { // dump timer fail ? should not happen
		gwlog.Error("%s.realMigrateTo: dump timers failed: %v", e, err)
	}

	migrateData := e.I.GetMigrateData()

	dispatcher_client.GetDispatcherClientForSend().SendRealMigrate(e.ID, spaceLoc, spaceID,
		float32(pos.X), float32(pos.Y), float32(pos.Z), e.TypeName, migrateData, timerData, clientid, clientsrv)
}

func OnRealMigrate(entityID EntityID, spaceID EntityID, x, y, z float32, typeName string,
	migrateData map[string]interface{}, timerData []byte,
	clientid ClientID, clientsrv uint16) {

	if entityManager.get(entityID) != nil {
		gwlog.Panicf("entity %s already exists", entityID)
	}

	// try to find the target space, but might be nil
	space := spaceManager.getSpace(spaceID)
	var client *GameClient
	if !clientid.IsNil() {
		client = MakeGameClient(clientid, clientsrv)
	}
	pos := Position{Coord(x), Coord(y), Coord(z)}
	createEntity(typeName, space, pos, entityID, migrateData, timerData, client, true)
}

func (e *Entity) OnMigrateOut() {
	if consts.DEBUG_MIGRATE {
		gwlog.Debug("%s.OnMigrateOut, space=%s, client=%s", e, e.Space, e.client)
	}
}

func (e *Entity) OnMigrateIn() {
	if consts.DEBUG_MIGRATE {
		gwlog.Debug("%s.OnMigrateIn, space=%s, client=%s", e, e.Space, e.client)
	}
}

//
func (e *Entity) SetFilterProp(key string, val string) {
	if consts.DEBUG_FILTER_PROP {
		gwlog.Debug("%s.SetFilterProp: %s = %s, client=%s", e, key, val, e.client)
	}

	curval, ok := e.filterProps[key]
	if ok && curval == val {
		return // not changed
	}

	e.filterProps[key] = val
	// send filter property to client
	if e.client != nil {
		dispatcher_client.GetDispatcherClientForSend().SendSetClientFilterProp(e.client.gateid, e.client.clientid, key, val)
	}
}

// Call the filtered clients with prop key = value
// The message is broadcast to filtered clientproxies directly without going through entities
func (e *Entity) CallFitleredClients(key string, val string, method string, args ...interface{}) {
	dispatcher_client.GetDispatcherClientForSend().SendCallFilterClientProxies(key, val, method, args)
}

// Move in Space

func (e *Entity) GetPosition() Position {
	return e.aoi.pos
}

func (e *Entity) SetPosition(pos Position) {
	space := e.Space
	if space == nil {
		gwlog.Warn("%s.SetPosition(%s): space is nil", e, pos)
		return
	}

	space.move(e, pos)
	pos = e.aoi.pos
	e.ForAllClients(func(client *GameClient) {
		client.UpdatePositionOnClient(e.ID, pos)
	})
}

func (e *Entity) GetYaw() Yaw {
	return e.yaw
}

func (e *Entity) SetYaw(yaw Yaw) {
	e.yaw = yaw
	e.ForAllClients(func(client *GameClient) {
		client.UpdateYawOnClient(e.ID, e.yaw)
	})
}

// Some Other Useful Utilities
func (e *Entity) PanicOnError(err error) {
	if err != nil {
		gwlog.Panic(err)
	}
}
