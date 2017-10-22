package entity

import (
	"fmt"
	"reflect"

	"time"

	"github.com/xiaonanln/goTimer"

	"unsafe"

	"github.com/xiaonanln/goworld/components/dispatcher/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
	"github.com/xiaonanln/goworld/engine/storage"
	"github.com/xiaonanln/typeconv"
)

var (
	saveInterval time.Duration
)

// Yaw is the type of entity yaw
type Yaw float32

type entityTimerInfo struct {
	FireTime       time.Time
	RepeatInterval time.Duration
	Method         string
	Args           []interface{}
	Repeat         bool
	rawTimer       *timer.Timer
}

// Entity is the basic execution unit in GoWorld server. Entities can be used to
// represent players, NPCs, monsters. Entities can migrate among spaces.
type Entity struct {
	ID       common.EntityID
	TypeName string
	I        IEntity
	V        reflect.Value

	destroyed bool
	typeDesc  *EntityTypeDesc
	Space     *Space
	aoi       aoi
	yaw       Yaw

	rawTimers   map[*timer.Timer]struct{}
	timers      map[EntityTimerID]*entityTimerInfo
	lastTimerId EntityTimerID

	client            *GameClient
	declaredServices  common.StringSet
	syncingFromClient bool

	Attrs *MapAttr

	enteringSpaceRequest struct {
		SpaceID     common.EntityID
		EnterPos    Vector3
		RequestTime int64
	}

	filterProps map[string]string

	syncInfoFlag syncInfoFlag
}

type syncInfoFlag int

const (
	sifSyncOwnClient syncInfoFlag = 1 << iota
	sifSyncNeighborClients
)

// IEntity declares functions can be override in Entity subclasses
type IEntity interface {
	// Entity Lifetime
	OnInit()    // Called when initializing entity struct, override to initialize entity custom fields
	OnCreated() // Called when entity is just created
	OnDestroy() // Called when entity is destroying (just before destroy)
	// Migration
	OnMigrateOut() // Called just before entity is migrating out
	OnMigrateIn()  // Called just after entity is migrating in
	// Freeze && Restore
	OnFreeze()   // Called when entity is freezing
	OnRestored() // Called when entity is restored
	// Space Operations
	OnEnterSpace()             // Called when entity leaves space
	OnLeaveSpace(space *Space) // Called when entity enters space
	// Client Notifications
	OnClientConnected()    // Called when client is connected to entity (become player)
	OnClientDisconnected() // Called when client disconnected
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

// Destroy destroys the entity
func (e *Entity) Destroy() {
	if e.destroyed {
		return
	}
	gwlog.Debugf("%s.Destroy ...", e)
	e.destroyEntity(false)
	dispatcherclient.GetDispatcherClientForSend().SendNotifyDestroyEntity(e.ID)
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

// IsDestroyed returns if the entity is destroyed
func (e *Entity) IsDestroyed() bool {
	return e.destroyed
}

// Save the entity
func (e *Entity) Save() {
	if !e.IsPersistent() {
		return
	}

	if consts.DEBUG_SAVE_LOAD {
		gwlog.Debugf("SAVING %s ...", e)
	}

	data := e.getPersistentData()

	storage.Save(e.TypeName, e.ID, data, nil)
}

// IsSpaceEntity returns if the entity is actually a space
func (e *Entity) IsSpaceEntity() bool {
	return e.TypeName == _SPACE_ENTITY_TYPE
}

// ToSpace converts entity to space (only works for space entity)
func (e *Entity) ToSpace() *Space {
	if !e.IsSpaceEntity() {
		gwlog.Panicf("%s is not a space", e)
	}

	return (*Space)(unsafe.Pointer(e))
}

func (e *Entity) init(typeName string, entityID common.EntityID, entityInstance reflect.Value) {
	e.ID = entityID
	e.V = entityInstance
	e.I = entityInstance.Interface().(IEntity)
	e.TypeName = typeName

	e.typeDesc = registeredEntityTypes[typeName]

	e.rawTimers = map[*timer.Timer]struct{}{}
	e.timers = map[EntityTimerID]*entityTimerInfo{}
	e.declaredServices = common.StringSet{}
	e.filterProps = map[string]string{}

	attrs := NewMapAttr()
	attrs.owner = e
	e.Attrs = attrs

	initAOI(&e.aoi)
	//gwutils.RunPanicless(e.I.OnInit)
	e.callCompositiveMethod("OnInit")
}

func (e *Entity) callCompositiveMethod(methodName string, args ...interface{}) {
	defer func() {
		if err := recover(); err != nil {
			gwlog.TraceError("%s call compositive method '%s' failed: %v", e, methodName, err)
		}
	}()

	entityPtr := e.V
	entityVal := reflect.Indirect(entityPtr)
	var methodIn []reflect.Value
	if len(args) > 0 {
		methodIn = make([]reflect.Value, len(args), len(args))
		for i := 0; i < len(args); i++ {
			methodIn[i] = reflect.ValueOf(args[i])
		}
	}

	if compIndices, ok := e.typeDesc.compositiveMethodComponentIndices[methodName]; ok {
		for _, ci := range compIndices {
			field := entityVal.Field(ci)
			//gwlog.Infof("Calling method %s on field %d=>%s", methodName, ci, field)
			field.Addr().MethodByName(methodName).Call(methodIn)
		}
	} else {
		// method is not a valid compositive method
		gwlog.Panicf("method %s is not a compositive method", methodName)
	}

	method := entityPtr.MethodByName(methodName)
	if method.IsValid() {
		method.Call(nil)
	}
}

func (e *Entity) setupSaveTimer() {
	e.addRawTimer(saveInterval, e.Save)
}

// SetSaveInterval sets the save interval for entity system
func SetSaveInterval(duration time.Duration) {
	saveInterval = duration
	gwlog.Infof("Save interval set to %s", saveInterval)
}

// Space Operations related to aoi

// Interests and Uninterest among entities
func (e *Entity) interest(other *Entity) {
	e.aoi.interest(other)
	e.client.sendCreateEntity(other, false)
}

func (e *Entity) uninterest(other *Entity) {
	e.aoi.uninterest(other)
	e.client.sendDestroyEntity(other)
}

// Neighbors get all neighbors in an EntitySet
//
// Never modify the return value !
func (e *Entity) Neighbors() EntitySet {
	return e.aoi.neighbors
}

func (e *Entity) IsNeighbor(other *Entity) bool {
	return e.aoi.neighbors.Contains(other)
}

// DistanceTo calculates the distance between two entities
func (e *Entity) DistanceTo(other *Entity) Coord {
	return e.aoi.pos.DistanceTo(other.aoi.pos)
}

// Timer & Callback Management

// EntityTimerID is the type of entity timer ID
type EntityTimerID int

// IsValid returns if the EntityTimerID is still valid (not fired and not cancelled)
func (tid EntityTimerID) IsValid() bool {
	return tid > 0
}

// AddCallback adds a one-time callback for the entity
//
// The callback will be cancelled if entity is destroyed
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
	gwlog.Debugf("%s.AddCallback %s: %d", e, method, tid)
	return tid
}

// AddTimer adds a repeat timer for the entity
//
// The callback will be cancelled if entity is destroyed
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
	gwlog.Debugf("%s.AddTimer %s: %d", e, method, tid)
	return tid
}

// CancelTimer cancels the Callback / Timer
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

func (e *Entity) dumpTimers() []byte {
	if len(e.timers) == 0 {
		return nil
	}

	timers := make([]*entityTimerInfo, 0, len(e.timers))
	for _, t := range e.timers {
		timers = append(timers, t)
	}
	e.timers = nil // no more AddCallback or AddTimer
	data, err := timersPacker.PackMsg(timers, nil)
	if err != nil {
		gwlog.TraceError("%s dump timers failed: %s", e, err)
	}
	//gwlog.Infof("%s dump %d timers: %v", e, len(timers), data)
	return data
}

func (e *Entity) restoreTimers(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var timers []*entityTimerInfo
	if err := timersPacker.UnpackMsg(data, &timers); err != nil {
		return err
	}
	gwlog.Debugf("%s: %d timers restored: %v", e, len(timers), timers)
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
func (e *Entity) Call(id common.EntityID, method string, args ...interface{}) {
	callEntity(id, method, args)
}

// CallService calls a service provider
func (e *Entity) CallService(serviceName string, method string, args ...interface{}) {
	serviceEid := entityManager.chooseServiceProvider(serviceName)
	callEntity(serviceEid, method, args)
}

func (e *Entity) syncPositionYawFromClient(x, y, z Coord, yaw Yaw) {
	//gwlog.Infof("%s.syncPositionYawFromClient: %v,%v,%v, yaw %v", e, x, y, z, yaw)
	if e.syncingFromClient {
		e.setPositionYaw(Vector3{x, y, z}, yaw, true)
	}
}

// SetClientSyncing set if entity infos (position, yaw) is syncing with client
func (e *Entity) SetClientSyncing(syncing bool) {
	e.syncingFromClient = syncing
}

func (e *Entity) onCallFromLocal(methodName string, args []interface{}) {
	defer func() {
		err := recover() // recover from any error during RPC call
		if err != nil {
			gwlog.TraceError("%s.%s paniced: %s", e, methodName, err)
		}
	}()

	rpcDesc := e.typeDesc.rpcDescs[methodName]
	if rpcDesc == nil {
		// rpc not found
		gwlog.Panicf("%s.onCallFromLocal: Method %s is not a valid RPC, args=%v", e, methodName, args)
	}

	// rpc call from server
	if rpcDesc.Flags&rfServer == 0 {
		// can not call from server
		gwlog.Panicf("%s.onCallFromLocal: Method %s can not be called from Server: flags=%v", e, methodName, rpcDesc.Flags)
	}

	if rpcDesc.NumArgs < len(args) {
		gwlog.Panicf("%s.onCallFromLocal: Method %s receives %d arguments, but given %d", e, methodName, rpcDesc.NumArgs, len(args))
	}

	methodType := rpcDesc.MethodType
	in := make([]reflect.Value, rpcDesc.NumArgs+1)
	in[0] = reflect.ValueOf(e.I) // first argument is the bind instance (self)

	for i, arg := range args {
		argType := methodType.In(i + 1)
		in[i+1] = typeconv.Convert(arg, argType)
	}

	for i := len(args); i < rpcDesc.NumArgs; i++ { // use zero value for missing arguments
		argType := methodType.In(i + 1)
		in[i+1] = reflect.Zero(argType)
	}

	rpcDesc.Func.Call(in)
}

func (e *Entity) onCallFromRemote(methodName string, args [][]byte, clientid common.ClientID) {
	defer func() {
		err := recover() // recover from any error during RPC call
		if err != nil {
			gwlog.TraceError("%s.%s paniced: %s", e, methodName, err)
		}
	}()

	rpcDesc := e.typeDesc.rpcDescs[methodName]
	if rpcDesc == nil {
		// rpc not found
		gwlog.Errorf("%s.onCallFromRemote: Method %s is not a valid RPC, args=%v", e, methodName, args)
		return
	}

	methodType := rpcDesc.MethodType
	if clientid == "" {
		// rpc call from server
		if rpcDesc.Flags&rfServer == 0 {
			// can not call from server
			gwlog.Panicf("%s.onCallFromRemote: Method %s can not be called from Server: flags=%v", e, methodName, rpcDesc.Flags)
		}
	} else {
		isFromOwnClient := clientid == e.getClientID()
		if rpcDesc.Flags&rfOwnClient == 0 && isFromOwnClient {
			gwlog.Panicf("%s.onCallFromRemote: Method %s can not be called from OwnClient: flags=%v", e, methodName, rpcDesc.Flags)
		} else if rpcDesc.Flags&rfOtherClient == 0 && !isFromOwnClient {
			gwlog.Panicf("%s.onCallFromRemote: Method %s can not be called from OtherClient: flags=%v, OwnClient=%s, OtherClient=%s", e, methodName, rpcDesc.Flags, e.getClientID(), clientid)
		}
	}

	if rpcDesc.NumArgs < len(args) {
		gwlog.Errorf("%s.onCallFromRemote: Method %s receives %d arguments, but given %d", e, methodName, rpcDesc.NumArgs, len(args))
		return
	}

	in := make([]reflect.Value, rpcDesc.NumArgs+1)
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

	for i := len(args); i < rpcDesc.NumArgs; i++ { // use zero value for missing arguments
		argType := methodType.In(i + 1)
		in[i+1] = reflect.Zero(argType)
	}

	rpcDesc.Func.Call(in)
}

// DeclareService declares global service for service entity
func (e *Entity) DeclareService(serviceName string) {
	e.declaredServices.Add(serviceName)
	dispatcherclient.GetDispatcherClientForSend().SendDeclareService(e.ID, serviceName)
}

// OnInit is called when entity is initializing
//
// Can override this function in custom entity type
func (e *Entity) OnInit() {
	//gwlog.Warnf("%s.OnInit not implemented", e)
}

// OnCreated is called when entity is created
//
// Can override this function in custom entity type
func (e *Entity) OnCreated() {
	//gwlog.Debugf("%s.OnCreated", e)
}

// OnFreeze is called when entity is freezed
//
// Can override this function in custom entity type
func (e *Entity) OnFreeze() {
}

// OnRestored is called when entity is restored
//
// Can override this function in custom entity type
func (e *Entity) OnRestored() {
}

// OnEnterSpace is called when entity enters space
//
// Can override this function in custom entity type
func (e *Entity) OnEnterSpace() {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("%s.OnEnterSpace >>> %s", e, e.Space)
	}
}

// OnLeaveSpace is called when entity leaves space
//
// Can override this function in custom entity type
func (e *Entity) OnLeaveSpace(space *Space) {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("%s.OnLeaveSpace <<< %s", e, space)
	}
}

// OnDestroy is called when entity is destroying
//
// Can override this function in custom entity type
func (e *Entity) OnDestroy() {
}

// Default handlers for persistence

// IsPersistent returns if the entity is persistent
//
// Default implementation check entity for persistent attributes
func (e *Entity) IsPersistent() bool {
	return e.typeDesc.isPersistent
}

// getPersistentData gets the persistent data
//
// Returns persistent attributes by default
func (e *Entity) getPersistentData() map[string]interface{} {
	return e.Attrs.ToMapWithFilter(e.typeDesc.persistentAttrs.Contains)
}

// loadPersistentData loads persistent data
//
// Load persistent data to attributes
func (e *Entity) loadPersistentData(data map[string]interface{}) {
	e.Attrs.AssignMap(data)
}

func (e *Entity) getClientData() map[string]interface{} {
	return e.Attrs.ToMapWithFilter(e.typeDesc.clientAttrs.Contains)
}

func (e *Entity) getAllClientData() map[string]interface{} {
	return e.Attrs.ToMapWithFilter(e.typeDesc.allClientAttrs.Contains)
}

// GetMigrateData gets the migration data
func (e *Entity) GetMigrateData() map[string]interface{} {
	return e.Attrs.ToMap() // all attrs are migrated, without filter
}

// LoadMigrateData loads migrate data
func (e *Entity) LoadMigrateData(data map[string]interface{}) {
	e.Attrs.AssignMap(data)
}

type clientData struct {
	ClientID common.ClientID
	GateID   uint16
}

type enteringSpaceRequestData struct {
	SpaceID  common.EntityID
	EnterPos Vector3
}

type entityFreezeData struct {
	Type      string
	TimerData []byte
	Pos       Vector3
	Attrs     map[string]interface{}
	Yaw       Yaw
	SpaceID   common.EntityID
	Client    *clientData
	ESR       *enteringSpaceRequestData
}

// GetFreezeData gets freezed data
func (e *Entity) GetFreezeData() *entityFreezeData {
	data := &entityFreezeData{
		Type:      e.TypeName,
		TimerData: e.dumpTimers(),
		Attrs:     e.Attrs.ToMap(),
		Pos:       e.aoi.pos,
		Yaw:       e.yaw,
		SpaceID:   e.Space.ID,
	}
	if e.client != nil {
		data.Client = &clientData{
			ClientID: e.client.clientid,
			GateID:   e.client.gateid,
		}
	}

	if !e.enteringSpaceRequest.SpaceID.IsNil() {
		data.ESR = &enteringSpaceRequestData{e.enteringSpaceRequest.SpaceID, e.enteringSpaceRequest.EnterPos}
	}

	return data
}

// Client related utilities

// GetClient returns the client of entity
func (e *Entity) GetClient() *GameClient {
	return e.client
}

func (e *Entity) getClientID() common.ClientID {
	if e.client != nil {
		return e.client.clientid
	}
	return ""
}

// SetClient sets the client of entity
func (e *Entity) SetClient(client *GameClient) {
	oldClient := e.client
	if oldClient == client {
		return
	}

	e.client = client

	if oldClient != nil {
		// send destroy entity to client
		entityManager.onEntityLoseClient(oldClient.clientid)
		dispatcherclient.GetDispatcherClientForSend().SendClearClientFilterProp(oldClient.gateid, oldClient.clientid)

		for neighbor := range e.Neighbors() {
			oldClient.sendDestroyEntity(neighbor)
		}

		oldClient.sendDestroyEntity(e)
	}

	if client != nil {
		// send create entity to new client
		entityManager.onEntityGetClient(e.ID, client.clientid)

		client.sendCreateEntity(e, true)

		for neighbor := range e.Neighbors() {
			client.sendCreateEntity(neighbor, false)
		}

		// set all filter properties to client
		for key, val := range e.filterProps {
			dispatcherclient.GetDispatcherClientForSend().SendSetClientFilterProp(client.gateid, client.clientid, key, val)
		}
	}

	if oldClient == nil && client != nil {
		// got net client
		gwutils.RunPanicless(e.I.OnClientConnected)
	} else if oldClient != nil && client == nil {
		gwutils.RunPanicless(e.I.OnClientDisconnected)
	}
}

// CallClient calls the client entity
func (e *Entity) CallClient(method string, args ...interface{}) {
	e.client.call(e.ID, method, args)
}

// CallAllClients calls the entity method on all clients
func (e *Entity) CallAllClients(method string, args ...interface{}) {
	e.client.call(e.ID, method, args)

	for neighbor := range e.Neighbors() {
		neighbor.client.call(e.ID, method, args)
	}
}

// GiveClientTo gives client to other entity
func (e *Entity) GiveClientTo(other *Entity) {
	if e.client == nil {
		gwlog.Warnf("%s.GiveClientTo(%s): client is nil", e, other)
		return
	}

	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.GiveClientTo(%s): client=%s", e, other, e.client)
	}
	client := e.client
	e.SetClient(nil)

	other.SetClient(client)
}

// ForAllClients visits all clients (own client and clients of neighbors)
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

// OnClientConnected is called when client is connected
//
// Can override this function in custom entity type
func (e *Entity) OnClientConnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.OnClientConnected: %s, %d Neighbors", e, e.client, len(e.Neighbors()))
	}
}

// OnClientDisconnected is called when client is disconnected
//
// Can override this function in custom entity type
func (e *Entity) OnClientDisconnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.OnClientDisconnected: %s", e, e.client)
	}
}

func (e *Entity) getAttrFlag(attrName string) (flag attrFlag) {
	if e.typeDesc.allClientAttrs.Contains(attrName) {
		flag = afAllClient
	} else if e.typeDesc.clientAttrs.Contains(attrName) {
		flag = afClient
	}

	return
}

func (e *Entity) sendMapAttrChangeToClients(ma *MapAttr, key string, val interface{}) {
	var flag attrFlag
	if ma == e.Attrs {
		// this is the root attr
		flag = e.getAttrFlag(key)
	} else {
		flag = ma.flag
	}

	if flag&afAllClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrChange(e.ID, path, key, val)
		for neighbor := range e.aoi.neighbors {
			neighbor.client.sendNotifyMapAttrChange(e.ID, path, key, val)
		}
	} else if flag&afClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrChange(e.ID, path, key, val)
	}
}

func (e *Entity) sendMapAttrDelToClients(ma *MapAttr, key string) {
	var flag attrFlag
	if ma == e.Attrs {
		// this is the root attr
		flag = e.getAttrFlag(key)
	} else {
		flag = ma.flag
	}

	if flag&afAllClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrDel(e.ID, path, key)
		for neighbor := range e.aoi.neighbors {
			neighbor.client.sendNotifyMapAttrDel(e.ID, path, key)
		}
	} else if flag&afClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrDel(e.ID, path, key)
	}
}

func (e *Entity) sendListAttrChangeToClients(la *ListAttr, index int, val interface{}) {
	flag := la.flag

	if flag&afAllClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrChange(e.ID, path, uint32(index), val)
		for neighbor := range e.aoi.neighbors {
			neighbor.client.sendNotifyListAttrChange(e.ID, path, uint32(index), val)
		}
	} else if flag&afClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrChange(e.ID, path, uint32(index), val)
	}
}

func (e *Entity) sendListAttrPopToClients(la *ListAttr) {
	flag := la.flag
	if flag&afAllClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrPop(e.ID, path)
		for neighbor := range e.aoi.neighbors {
			neighbor.client.sendNotifyListAttrPop(e.ID, path)
		}
	} else if flag&afClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrPop(e.ID, path)
	}
}

func (e *Entity) sendListAttrAppendToClients(la *ListAttr, val interface{}) {
	flag := la.flag
	if flag&afAllClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrAppend(e.ID, path, val)
		for neighbor := range e.aoi.neighbors {
			neighbor.client.sendNotifyListAttrAppend(e.ID, path, val)
		}
	} else if flag&afClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrAppend(e.ID, path, val)
	}
}

// Define Attributes Properties

// Fast access to attrs

// GetInt gets an outtermost attribute as int
func (e *Entity) GetInt(key string) int64 {
	return e.Attrs.GetInt(key)
}

// GetStr gets an outtermost attribute as string
func (e *Entity) GetStr(key string) string {
	return e.Attrs.GetStr(key)
}

// GetFloat gets an outtermost attribute as float64
func (e *Entity) GetFloat(key string) float64 {
	return e.Attrs.GetFloat(key)
}

// GetMapAttr gets an outtermost attribute as MapAttr
func (e *Entity) GetMapAttr(key string) *MapAttr {
	return e.Attrs.GetMapAttr(key)
}

// GetListAttr gets an outtermost attribute as ListAttr
func (e *Entity) GetListAttr(key string) *ListAttr {
	return e.Attrs.GetListAttr(key)
}

// Enter Space

// EnterSpace let the entity enters space
func (e *Entity) EnterSpace(spaceID common.EntityID, pos Vector3) {
	if e.isEnteringSpace() {
		gwlog.Errorf("%s is entering space %s, can not enter space %s", e, e.enteringSpaceRequest.SpaceID, spaceID)
		e.I.OnEnterSpace()
		return
	}

	if consts.OPTIMIZE_LOCAL_ENTITIES {
		localSpace := spaceManager.getSpace(spaceID)
		if localSpace != nil { // target space is local, just enter
			e.enterLocalSpace(localSpace, pos)
		} else { // else request migrating to other space
			e.requestMigrateTo(spaceID, pos)
		}
	} else {
		e.requestMigrateTo(spaceID, pos)
	}
}

func (e *Entity) enterLocalSpace(space *Space, pos Vector3) {
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
			gwlog.Warnf("%s: space %s is destroyed, enter space cancelled", e, space.ID)
			return
		}

		//gwlog.Infof("%s.enterLocalSpace ==> %s", e, space)
		e.Space.leave(e)
		space.enter(e, pos, false)
	})
}

func (e *Entity) isEnteringSpace() bool {
	now := time.Now().UnixNano()
	return now < (e.enteringSpaceRequest.RequestTime + int64(consts.ENTER_SPACE_REQUEST_TIMEOUT))
}

// Migrate to the server of space
func (e *Entity) requestMigrateTo(spaceID common.EntityID, pos Vector3) {
	e.enteringSpaceRequest.SpaceID = spaceID
	e.enteringSpaceRequest.EnterPos = pos
	e.enteringSpaceRequest.RequestTime = time.Now().UnixNano()

	dispatcherclient.GetDispatcherClientForSend().SendMigrateRequest(spaceID, e.ID)
}

func (e *Entity) clearEnteringSpaceRequest() {
	e.enteringSpaceRequest.SpaceID = ""
	e.enteringSpaceRequest.EnterPos = Vector3{}
	e.enteringSpaceRequest.RequestTime = 0
}

// OnMigrateRequestAck is called by engine when mgirate request Ack is received
func OnMigrateRequestAck(entityID common.EntityID, spaceID common.EntityID, spaceLoc uint16) {
	entity := entityManager.get(entityID)
	if entity == nil {
		//dispatcher_client.GetDispatcherClientForSend().SendCancelMigrateRequest(entityID)
		gwlog.Errorf("Migrate failed since entity is destroyed: spaceID=%s, entityID=%s", spaceID, entityID)
		return
	}

	if spaceLoc == 0 {
		// target space not found, migrate not started
		gwlog.Errorf("Migrate failed since target space is not found: spaceID=%s, entity=%s", spaceID, entity)
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

func (e *Entity) realMigrateTo(spaceID common.EntityID, pos Vector3, spaceLoc uint16) {
	var clientid common.ClientID
	var clientsrv uint16
	if e.client != nil {
		clientid = e.client.clientid
		clientsrv = e.client.gateid
	}

	e.destroyEntity(true) // disable the entity
	timerData := e.dumpTimers()
	migrateData := e.GetMigrateData()

	dispatcherclient.GetDispatcherClientForSend().SendRealMigrate(e.ID, spaceLoc, spaceID,
		float32(pos.X), float32(pos.Y), float32(pos.Z), e.TypeName, migrateData, timerData, clientid, clientsrv)
}

// OnRealMigrate is used by entity migration
func OnRealMigrate(entityID common.EntityID, spaceID common.EntityID, x, y, z float32, typeName string,
	migrateData map[string]interface{}, timerData []byte,
	clientid common.ClientID, clientsrv uint16) {

	if entityManager.get(entityID) != nil {
		gwlog.Panicf("entity %s already exists", entityID)
	}

	// try to find the target space, but might be nil
	space := spaceManager.getSpace(spaceID)
	var client *GameClient
	if !clientid.IsNil() {
		client = MakeGameClient(clientid, clientsrv)
	}
	pos := Vector3{Coord(x), Coord(y), Coord(z)}
	createEntity(typeName, space, pos, entityID, migrateData, timerData, client, ccMigrate)
}

// OnMigrateOut is called when entity is migrating out
//
// Can override this function in custom entity type
func (e *Entity) OnMigrateOut() {
	if consts.DEBUG_MIGRATE {
		gwlog.Debugf("%s.OnMigrateOut, space=%s, client=%s", e, e.Space, e.client)
	}
}

// OnMigrateIn is called when entity is migrating in
//
// Can override this function in custom entity type
func (e *Entity) OnMigrateIn() {
	if consts.DEBUG_MIGRATE {
		gwlog.Debugf("%s.OnMigrateIn, space=%s, client=%s", e, e.Space, e.client)
	}
}

// SetFilterProp sets a filter property key-value
func (e *Entity) SetFilterProp(key string, val string) {
	if consts.DEBUG_FILTER_PROP {
		gwlog.Debugf("%s.SetFilterProp: %s = %s, client=%s", e, key, val, e.client)
	}

	curval, ok := e.filterProps[key]
	if ok && curval == val {
		return // not changed
	}

	e.filterProps[key] = val
	// send filter property to client
	if e.client != nil {
		dispatcherclient.GetDispatcherClientForSend().SendSetClientFilterProp(e.client.gateid, e.client.clientid, key, val)
	}
}

// CallFitleredClients calls the filtered clients with prop key == value
//
// The message is broadcast to filtered clientproxies directly without going through entities
func (e *Entity) CallFitleredClients(key string, val string, method string, args ...interface{}) {
	dispatcherclient.GetDispatcherClientForSend().SendCallFilterClientProxies(key, val, method, args)
}

// IsUseAOI returns if entity type is using aoi
//
// Entities like Account, Service entities should not be using aoi
func (e *Entity) IsUseAOI() bool {
	return e.typeDesc.useAOI
}

// GetPosition returns the entity position
func (e *Entity) GetPosition() Vector3 {
	return e.aoi.pos
}

// SetPosition sets the entity position
func (e *Entity) SetPosition(pos Vector3) {
	e.setPositionYaw(pos, e.yaw, false)
}

func (e *Entity) setPositionYaw(pos Vector3, yaw Yaw, fromClient bool) {
	space := e.Space
	if space == nil {
		gwlog.Warnf("%s.SetPosition(%s): space is nil", e, pos)
		return
	}

	space.move(e, pos)
	e.yaw = yaw

	// mark the entity as needing sync
	// Real sync packets will be sent before flushing dispatcher client
	e.syncInfoFlag |= sifSyncNeighborClients
	if !fromClient {
		e.syncInfoFlag |= sifSyncOwnClient
	}
}

// CollectEntitySyncInfos is called by game service to collect and broadcast entity sync infos to all clients
func CollectEntitySyncInfos() {
	cfg := config.Get()
	gateCount := len(cfg.Gates)
	entitySyncInfosToGate := make([]*netutil.Packet, gateCount)
	for gateid := 1; gateid <= gateCount; gateid++ {
		packet := netutil.NewPacket()
		packet.AppendUint16(proto.MT_SYNC_POSITION_YAW_ON_CLIENTS)
		packet.AppendUint16(uint16(gateid))
		entitySyncInfosToGate[gateid-1] = packet
	}

	for eid, e := range entityManager.entities {
		syncInfoFlag := e.syncInfoFlag
		if syncInfoFlag == 0 {
			continue
		}

		e.syncInfoFlag = 0
		syncInfo := e.getSyncInfo()
		if syncInfoFlag&sifSyncOwnClient != 0 && e.client != nil {
			gateid := e.client.gateid
			packet := entitySyncInfosToGate[gateid-1]
			packet.AppendClientID(e.client.clientid)
			packet.AppendEntityID(eid)
			packet.AppendFloat32(syncInfo.X)
			packet.AppendFloat32(syncInfo.Y)
			packet.AppendFloat32(syncInfo.Z)
			packet.AppendFloat32(syncInfo.Yaw)
		}
		if syncInfoFlag&sifSyncNeighborClients != 0 {
			for neighbor := range e.aoi.neighbors {
				client := neighbor.client
				if client != nil {
					gateid := client.gateid
					packet := entitySyncInfosToGate[gateid-1]
					packet.AppendClientID(client.clientid)
					packet.AppendEntityID(eid)
					packet.AppendFloat32(syncInfo.X)
					packet.AppendFloat32(syncInfo.Y)
					packet.AppendFloat32(syncInfo.Z)
					packet.AppendFloat32(syncInfo.Yaw)
				}
			}
		}
	}

	// send to dispatcher, one gate by one gate
	for _, packet := range entitySyncInfosToGate {
		//gwlog.Infof("SYNC %d PAYLOAD %d", gateid, packet.GetPayloadLen())

		if packet.GetPayloadLen() > 4 {
			dispatcherclient.GetDispatcherClientForSend().SendPacket(packet)
		}

		packet.Release()
	}
}

func (e *Entity) getSyncInfo() proto.EntitySyncInfo {
	return proto.EntitySyncInfo{
		float32(e.aoi.pos.X),
		float32(e.aoi.pos.Y),
		float32(e.aoi.pos.Z),
		float32(e.yaw),
	}
}

// GetYaw gets entity yaw
func (e *Entity) GetYaw() Yaw {
	return e.yaw
}

// SetYaw sets entity yaw
func (e *Entity) SetYaw(yaw Yaw) {
	e.yaw = yaw
	e.syncInfoFlag |= (sifSyncNeighborClients | sifSyncOwnClient)
	//e.ForAllClients(func(client *GameClient) {
	//	client.updateYawOnClient(e.ID, e.yaw)
	//})
}

// FaceTo let entity face to another entity by setting yaw accordingly
func (e *Entity) FaceTo(other *Entity) {
	e.FaceToPos(other.aoi.pos)
}

// FaceTo let entity face to a specified position, setting yaw accordingly

func (e *Entity) FaceToPos(pos Vector3) {
	dir := pos.Sub(e.aoi.pos)
	dir.Y = 0

	e.SetYaw(dir.DirToYaw())
}

// Some Other Useful Utilities

// PanicOnError panics if err != nil
func (e *Entity) PanicOnError(err error) {
	if err != nil {
		gwlog.Panic(err)
	}
}
