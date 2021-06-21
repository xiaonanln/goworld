package entity

import (
	"fmt"
	"reflect"

	"time"

	"unsafe"

	"github.com/pkg/errors"
	"github.com/xiaonanln/go-aoi"
	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
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

// Yaw is the type of entity Yaw
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
	ID                   common.EntityID
	TypeName             string
	I                    IEntity
	V                    reflect.Value
	destroyed            bool
	typeDesc             *EntityTypeDesc
	Space                *Space
	Position             Vector3
	InterestedIn         EntitySet
	InterestedBy         EntitySet
	aoi                  aoi.AOI
	yaw                  Yaw
	rawTimers            map[*timer.Timer]struct{}
	timers               map[EntityTimerID]*entityTimerInfo
	lastTimerId          EntityTimerID
	client               *GameClient
	syncingFromClient    bool
	Attrs                *MapAttr
	syncInfoFlag         syncInfoFlag
	enteringSpaceRequest struct {
		SpaceID              common.EntityID
		EnterPos             Vector3
		RequestTime          int64
		migrateRequestIsSent bool
	}
}

type clientData struct {
	ClientID common.ClientID
	GateID   uint16
}

// entity info that should be migrated
type entityMigrateData struct {
	Type              string                 `msgpack:"T"`
	Attrs             map[string]interface{} `msgpack:"A"`
	Client            *clientData            `msgpack:"C,omitempty"`
	Pos               Vector3                `msgpack:"Pos"`
	Yaw               Yaw                    `msgpack:"Yaw"`
	SpaceID           common.EntityID        `msgpack:"SP"`
	TimerData         []byte                 `msgpack:"TD,omitempty"`
	FilterProps       map[string]string      `msgpack:"FP"`
	SyncingFromClient bool                   `msgpack:"SFC"`
	SyncInfoFlag      syncInfoFlag           `msgpack:"SIF"`
}

type syncInfoFlag int

const (
	sifSyncOwnClient syncInfoFlag = 1 << iota
	sifSyncNeighborClients
)

// IEntity declares functions that is defined in Entity
// These functions are mostly component functions
type IEntity interface {
	// Entity Lifetime
	OnInit()       // Called when initializing entity struct, override to initialize entity custom fields
	OnAttrsReady() // Called when entity attributes are ready.
	OnCreated()    // Called when entity is just created
	OnDestroy()    // Called when entity is destroying (just before destroy)
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
	OnClientConnected()    // Called when Client is connected to entity (become player)
	OnClientDisconnected() // Called when Client disconnected

	DescribeEntityType(desc *EntityTypeDesc) // Define entity attributes in this function
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
	dispatchercluster.SendNotifyDestroyEntity(e.ID)
}

func (e *Entity) destroyEntity(isMigrate bool) {
	e.Space.leave(e)

	if !isMigrate {
		e.I.OnDestroy()
	} else {
		e.I.OnMigrateOut()
	}

	e.clearRawTimers()
	e.rawTimers = nil // prohibit further use

	if !isMigrate {
		e.SetClient(nil) // always set Client to nil before destroy
		e.Save()
	} else {
		e.assignClient(nil)
	}

	e.destroyed = true
	entityManager.del(e)
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

// AsSpace converts entity to space (only works for space entity)
func (e *Entity) AsSpace() *Space {
	if !e.IsSpaceEntity() {
		gwlog.Panicf("%s is not a space", e)
	}

	return (*Space)(unsafe.Pointer(e))
}

func (e *Entity) init(typeName string, entityid common.EntityID, entityInstance reflect.Value) {
	e.ID = entityid
	e.V = entityInstance
	e.I = entityInstance.Interface().(IEntity)
	e.TypeName = typeName

	e.typeDesc = registeredEntityTypes[typeName]

	e.rawTimers = map[*timer.Timer]struct{}{}
	e.timers = map[EntityTimerID]*entityTimerInfo{}

	attrs := NewMapAttr()
	attrs.owner = e
	e.Attrs = attrs

	e.InterestedIn = EntitySet{}
	e.InterestedBy = EntitySet{}
	aoi.InitAOI(&e.aoi, aoi.Coord(e.typeDesc.aoiDistance), e, e)

	e.I.OnInit()
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

func (e *Entity) OnEnterAOI(otherAoi *aoi.AOI) {
	e.interest(otherAoi.Data.(*Entity))
}

func (e *Entity) OnLeaveAOI(otherAoi *aoi.AOI) {
	e.uninterest(otherAoi.Data.(*Entity))
}

// Interests and Uninterest among entities
func (e *Entity) interest(other *Entity) {
	e.InterestedIn.Add(other)
	other.InterestedBy.Add(e)
	e.client.sendCreateEntity(other, false)
}

func (e *Entity) uninterest(other *Entity) {
	e.InterestedIn.Del(other)
	other.InterestedBy.Del(e)
	e.client.sendDestroyEntity(other)
}

// IsInterestedIn checks if other entity is interested by this entity
func (e *Entity) IsInterestedIn(other *Entity) bool {
	return e.InterestedIn.Contains(other)
}

// DistanceTo calculates the distance between two entities
func (e *Entity) DistanceTo(other *Entity) Coord {
	return e.Position.DistanceTo(other.Position)
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
	Call(id, method, args)
}

func (e *Entity) syncPositionYawFromClient(x, y, z Coord, yaw Yaw) {
	//gwlog.Infof("%s.syncPositionYawFromClient: %v,%v,%v, Yaw %v, syncing %v", e, x, y, z, Yaw, e.SyncingFromClient)
	if e.syncingFromClient {
		e.setPositionYaw(Vector3{x, y, z}, yaw, true)
	}
}

// SetClientSyncing set if entity infos (position, Yaw) is syncing with Client
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
	in[0] = e.V // first argument is the bind instance (self)

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
	in[0] = e.V // first argument is the bind instance (self)

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

// OnInit is called when entity is initializing
//
// Can override this function in custom entity type
func (e *Entity) OnInit() {
	//gwlog.Warnf("%s.OnInit not implemented", e)
}

// OnAttrsReady is called when entity's attribute is ready
//
// Can override this function in custom entity type
func (e *Entity) OnAttrsReady() {

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
	return e.typeDesc.IsPersistent
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
func (e *Entity) GetMigrateData(spaceid common.EntityID, pos Vector3) *entityMigrateData {
	md := &entityMigrateData{
		Type:              e.TypeName,
		Attrs:             e.Attrs.ToMap(), // all Attrs are migrated, without filter
		Pos:               pos,
		Yaw:               e.yaw,
		TimerData:         e.dumpTimers(),
		SpaceID:           spaceid,
		SyncingFromClient: e.syncingFromClient,
		SyncInfoFlag:      e.syncInfoFlag,
	}

	if e.client != nil {
		md.Client = &clientData{
			ClientID: e.client.clientid,
			GateID:   e.client.gateid,
		}
	}

	return md
}

// loadMigrateData loads migrate data
func (e *Entity) loadMigrateData(data map[string]interface{}) {
	e.Attrs.AssignMap(data)
}

// getFreezeData gets freezed data
func (e *Entity) getFreezeData() *entityMigrateData {
	return e.GetMigrateData(e.Space.ID, e.Position)
}

// Client related utilities

// GetClient returns the Client of entity
func (e *Entity) GetClient() *GameClient {
	return e.client
}

func (e *Entity) getClientID() common.ClientID {
	if e.client != nil {
		return e.client.clientid
	}
	return ""
}

// SetClient sets the Client of entity
func (e *Entity) SetClient(client *GameClient) {
	oldClient := e.client
	if oldClient == client {
		return
	}

	if oldClient != nil {
		// send destroy entity to Client
		dispatchercluster.SelectByEntityID(e.ID).SendClearClientFilterProp(oldClient.gateid, oldClient.clientid)

		for neighbor := range e.InterestedBy {
			oldClient.sendDestroyEntity(neighbor)
		}

		if !e.Space.IsNil() {
			oldClient.sendDestroyEntity(&e.Space.Entity)
		}

		oldClient.sendDestroyEntity(e)
	}

	e.assignClient(client) // remove old client, assign new client

	if client != nil {
		// send create entity to new client
		dispatchercluster.SelectByEntityID(e.ID).SendClearClientFilterProp(client.gateid, client.clientid)
		client.sendCreateEntity(e, true)

		if !e.Space.IsNil() {
			client.sendCreateEntity(&e.Space.Entity, false)
		}

		for neighbor := range e.InterestedBy {
			client.sendCreateEntity(neighbor, false)
		}
	}

	if oldClient != nil && client == nil {
		gwutils.RunPanicless(func() {
			e.I.OnClientDisconnected()
		})
	} else if client != nil {
		gwutils.RunPanicless(func() {
			e.I.OnClientConnected()
		})
	}
}

func (e *Entity) assignClient(client *GameClient) {
	if e.client != nil {
		e.client.ownerid = ""
	}

	e.client = client
	if client != nil {
		client.ownerid = e.ID
	}
}

// CallClient calls the Client entity
func (e *Entity) CallClient(method string, args ...interface{}) {
	e.client.call(e.ID, method, args)
}

// CallAllClients calls the entity method on all clients
func (e *Entity) CallAllClients(method string, args ...interface{}) {
	e.client.call(e.ID, method, args)

	for neighbor := range e.InterestedBy {
		neighbor.client.call(e.ID, method, args)
	}
}

// GiveClientTo gives Client to other entity
func (e *Entity) GiveClientTo(other *Entity) {
	if e.client == nil {
		gwlog.Warnf("%s.GiveClientTo(%s): Client is nil", e, other)
		return
	}

	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.GiveClientTo(%s): Client=%s", e, other, e.client)
	}
	client := e.client
	client.ownerid = other.ID // hack ownerid so that destroy entity messages will be synced with create entity messages
	e.SetClient(nil)
	other.SetClient(client)
}

// ForAllClients visits all clients (own Client and clients of neighbors)
func (e *Entity) ForAllClients(f func(client *GameClient)) {
	if e.client != nil {
		f(e.client)
	}

	for neighbor := range e.InterestedBy {
		if neighbor.client != nil {
			f(neighbor.client)
		}
	}
}

func (e *Entity) notifyClientDisconnected() {
	// called when Client disconnected
	e.assignClient(nil)
	e.I.OnClientDisconnected()
}

// OnClientConnected is called when Client is connected
//
// Can override this function in custom entity type
func (e *Entity) OnClientConnected() {
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.OnClientConnected: %s, %d Neighbors", e, e.client, len(e.InterestedIn))
	}
}

// OnClientDisconnected is called when Client is disconnected
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
		for neighbor := range e.InterestedBy {
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
		for neighbor := range e.InterestedBy {
			neighbor.client.sendNotifyMapAttrDel(e.ID, path, key)
		}
	} else if flag&afClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrDel(e.ID, path, key)
	}
}

func (e *Entity) sendMapAttrClearToClients(ma *MapAttr) {
	if ma == e.Attrs {
		// this is the root attr
		gwlog.Panicf("outmost e.Attrs can not be cleared")
	}
	flag := ma.flag

	if flag&afAllClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrClear(e.ID, path)
		for neighbor := range e.InterestedBy {
			neighbor.client.sendNotifyMapAttrClear(e.ID, path)
		}
	} else if flag&afClient != 0 {
		path := ma.getPathFromOwner()
		e.client.sendNotifyMapAttrClear(e.ID, path)
	}
}

func (e *Entity) sendListAttrChangeToClients(la *ListAttr, index int, val interface{}) {
	flag := la.flag

	if flag&afAllClient != 0 {
		// TODO: only pack 1 packet, do not marshal multiple times
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrChange(e.ID, path, uint32(index), val)
		for neighbor := range e.InterestedBy {
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
		for neighbor := range e.InterestedBy {
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
		for neighbor := range e.InterestedBy {
			neighbor.client.sendNotifyListAttrAppend(e.ID, path, val)
		}
	} else if flag&afClient != 0 {
		path := la.getPathFromOwner()
		e.client.sendNotifyListAttrAppend(e.ID, path, val)
	}
}

// Define Attributes Properties

// Fast access to Attrs

// GetInt gets an outtermost attribute as int
func (e *Entity) GetInt(key string) int64 {
	return e.Attrs.GetInt(key)
}

// GetBool gets an outtermost attribute as int
func (e *Entity) GetBool(key string) bool {
	return e.Attrs.GetBool(key)
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
func (e *Entity) EnterSpace(spaceid common.EntityID, pos Vector3) {
	if e.isEnteringSpace() {
		gwlog.Errorf("%s is entering space %s, can not enter space %s", e, e.enteringSpaceRequest.SpaceID, spaceid)
		e.I.OnEnterSpace()
		return
	}

	if consts.OPTIMIZE_LOCAL_ENTITY_CALL {
		localSpace := spaceManager.getSpace(spaceid)
		if localSpace != nil { // target space is local, just enter
			e.enterLocalSpace(localSpace, pos)
		} else { // else request migrating to other space
			e.requestMigrateTo(spaceid, pos)
		}
	} else {
		e.requestMigrateTo(spaceid, pos)
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
		e.cancelEnterSpace()

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
func (e *Entity) requestMigrateTo(spaceid common.EntityID, pos Vector3) {
	e.enteringSpaceRequest.SpaceID = spaceid
	e.enteringSpaceRequest.EnterPos = pos
	e.enteringSpaceRequest.RequestTime = time.Now().UnixNano()

	dispatchercluster.SelectByEntityID(spaceid).SendQuerySpaceGameIDForMigrate(spaceid, e.ID)
}

func (e *Entity) cancelEnterSpace() {
	e.enteringSpaceRequest.SpaceID = ""
	e.enteringSpaceRequest.EnterPos = Vector3{}
	e.enteringSpaceRequest.RequestTime = 0
	if e.enteringSpaceRequest.migrateRequestIsSent {
		// if migrate request is already sent, the entity is already blocked in dispatcher, so we should tell the dispatcher to unblock this entity
		e.enteringSpaceRequest.migrateRequestIsSent = false
		dispatchercluster.SelectByEntityID(e.ID).SendCancelMigrate(e.ID)
	}
}

// OnQuerySpaceGameIDForMigrateAck is called by engine when query entity gameid ACK is received
func OnQuerySpaceGameIDForMigrateAck(entityid common.EntityID, spaceid common.EntityID, spaceGameID uint16) {

	//gwlog.Infof("OnQuerySpaceGameIDForMigrateAck: entityid=%s, spaceid=%s, spaceGameID=%v", entityid, spaceid, spaceGameID)

	entity := entityManager.get(entityid)
	if entity == nil {
		//dispatcher_client.GetDispatcherClientForSend().SendCancelMigrateRequest(entityid)
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since entity is destroyed: entityid=%s, spaceid=%s", entityid, spaceid)
		return
	}

	if !entity.isEnteringSpace() {
		// replay from dispatcher is too late ?
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since entity is not migrating: entity=%s, spaceid=%s", entity, spaceid)
		return
	}

	if entity.enteringSpaceRequest.SpaceID != spaceid {
		// not entering this space ?
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since entity is enter other space: entity=%s, spaceid=%s, other space=%s", entity, spaceid, entity.enteringSpaceRequest.SpaceID)
		return
	}

	if spaceGameID == 0 {
		// target space not found, migrate not started
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since target space is not found: entity=%s, spaceid=%s", entity, spaceid)
		entity.cancelEnterSpace()
		return
	}

	entity.enteringSpaceRequest.migrateRequestIsSent = true
	dispatchercluster.SendMigrateRequest(entityid, spaceid, spaceGameID)
}

// OnMigrateRequestAck is called by engine when mgirate request Ack is received
func OnMigrateRequestAck(entityid common.EntityID, spaceid common.EntityID, spaceGameID uint16) {
	//gwlog.Infof("OnMigrateRequestAck: entityid=%s, spaceid=%s, spaceGameID=%v", entityid, spaceid, spaceGameID)
	entity := entityManager.get(entityid)
	if entity == nil {
		//dispatcher_client.GetDispatcherClientForSend().SendCancelMigrateRequest(entityid)
		gwlog.Errorf("Migrate failed since entity is destroyed: spaceid=%s, entityid=%s", spaceid, entityid)
		return
	}

	if !entity.isEnteringSpace() {
		// replay from dispatcher is too late ?
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since entity is not migrating: entity=%s, spaceid=%s", entity, spaceid)
		return
	}

	if entity.enteringSpaceRequest.SpaceID != spaceid {
		// not entering this space ?
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since entity is enter other space: entity=%s, spaceid=%s, other space=%s", entity, spaceid, entity.enteringSpaceRequest.SpaceID)
		return
	}

	if spaceGameID == 0 {
		// target space not found, migrate not started
		gwlog.Errorf("entity.OnQuerySpaceGameIDForMigrateAck: migrate failed since target space is not found: entity=%s, spaceid=%s", entity, spaceid)
		entity.cancelEnterSpace()
		return
	}

	entity.realMigrateTo(spaceid, entity.enteringSpaceRequest.EnterPos, spaceGameID)
}

func (e *Entity) realMigrateTo(spaceid common.EntityID, pos Vector3, spaceGameID uint16) {
	migrateData := e.GetMigrateData(spaceid, pos)
	data, err := netutil.MSG_PACKER.PackMsg(migrateData, nil)
	if err != nil {
		gwlog.Panicf("%s is migrating to space %s, but pack migrate data failed: %s", e, spaceid, err)
	}

	e.destroyEntity(true) // disable the entity
	dispatchercluster.SendRealMigrate(e.ID, spaceGameID, data)
}

// OnRealMigrate is used by entity migration
func OnRealMigrate(entityid common.EntityID, data []byte) {
	if entityManager.get(entityid) != nil {
		gwlog.Panicf("entity %s already exists", entityid)
	}

	var md entityMigrateData
	if err := netutil.MSG_PACKER.UnpackMsg(data, &md); err != nil {
		gwlog.Panic(errors.Wrap(err, "unpack migrate data failed"))
	}

	restoreEntity(entityid, &md, false)
}

// OnMigrateOut is called when entity is migrating out
//
// Can override this function in custom entity type
func (e *Entity) OnMigrateOut() {
	if consts.DEBUG_MIGRATE {
		gwlog.Debugf("%s.OnMigrateOut, space=%s, Client=%s", e, e.Space, e.client)
	}
}

// OnMigrateIn is called when entity is migrating in
//
// Can override this function in custom entity type
func (e *Entity) OnMigrateIn() {
	if consts.DEBUG_MIGRATE {
		gwlog.Debugf("%s.OnMigrateIn, space=%s, Client=%s", e, e.Space, e.client)
	}
}

// SetClientFilterProp sets a filter property key-value
func (e *Entity) SetClientFilterProp(key string, val string) {
	if key == "" {
		gwlog.Panicf("%s SetClientFilterProp: key must not be empty", e)
	}

	// send filter property to Client
	e.client.sendSetClientFilterProp(key, val)
}

// CallFilteredClients calls the filtered clients with prop key == value
// supported op includes "=", "!=", "<", "<=", ">", ">="
// if key = "", all clients are called despite the value of op and val
//
// The message is broadcast to filtered clientproxies directly without going through entities, and therefore more efficient
func (e *Entity) CallFilteredClients(key, op, val string, method string, args ...interface{}) {
	// parse op from string to FilterClientsOpType
	var realop proto.FilterClientsOpType
	if op == "=" {
		realop = proto.FILTER_CLIENTS_OP_EQ
	} else if op == "!=" {
		realop = proto.FILTER_CLIENTS_OP_NE
	} else if op == ">" {
		realop = proto.FILTER_CLIENTS_OP_GT
	} else if op == "<" {
		realop = proto.FILTER_CLIENTS_OP_LT
	} else if op == ">=" {
		realop = proto.FILTER_CLIENTS_OP_GTE
	} else if op == "<=" {
		realop = proto.FILTER_CLIENTS_OP_LTE
	} else {
		gwlog.Panicf("%s.CallFilteredClients: unsupported op: calling method %s on clients filtered by %s %s %s", e, method, key, op, val)
	}

	dispatchercluster.SendCallFilterClientProxies(realop, key, val, method, args)
}

// IsUseAOI returns if entity type is using aoi
//
// Entities like Account, Service entities should not be using aoi
func (e *Entity) IsUseAOI() bool {
	return e.typeDesc.useAOI
}

// GetPosition returns the entity position
func (e *Entity) GetPosition() Vector3 {
	return e.Position
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
	// Real sync packets will be sent before flushing dispatcher Client
	e.syncInfoFlag |= sifSyncNeighborClients
	if !fromClient {
		e.syncInfoFlag |= sifSyncOwnClient
	}
}

// CollectEntitySyncInfos is called by game service to collect and broadcast entity sync infos to all clients
var entitySyncInfosToGate = map[uint16]*netutil.Packet{}

func getEntitySyncInfosPacket(gateid uint16) *netutil.Packet {
	pkt := entitySyncInfosToGate[gateid]
	if pkt == nil {
		pkt = netutil.NewPacket()
		pkt.AppendUint16(proto.MT_SYNC_POSITION_YAW_ON_CLIENTS)
		pkt.AppendUint16(gateid)
		entitySyncInfosToGate[gateid] = pkt
	}
	return pkt
}

func CollectEntitySyncInfos() {
	for eid, e := range entityManager.entities {
		syncInfoFlag := e.syncInfoFlag
		if syncInfoFlag == 0 {
			continue
		}

		e.syncInfoFlag = 0
		syncInfo := e.getSyncInfo()
		if syncInfoFlag&sifSyncOwnClient != 0 && e.client != nil {
			gateid := e.client.gateid
			packet := getEntitySyncInfosPacket(gateid)
			packet.AppendClientID(e.client.clientid)
			packet.AppendEntityID(eid)
			packet.AppendFloat32(syncInfo.X)
			packet.AppendFloat32(syncInfo.Y)
			packet.AppendFloat32(syncInfo.Z)
			packet.AppendFloat32(syncInfo.Yaw)
		}
		if syncInfoFlag&sifSyncNeighborClients != 0 {
			for neighbor := range e.InterestedBy {
				client := neighbor.client
				if client != nil {
					gateid := client.gateid
					packet := getEntitySyncInfosPacket(gateid)
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
	if len(entitySyncInfosToGate) > 0 {
		for gateid, packet := range entitySyncInfosToGate {
			//gwlog.Infof("SYNC %d PAYLOAD %d", gateid, packet.GetPayloadLen())
			dispatchercluster.SelectByGateID(gateid).SendPacket(packet)
			packet.Release()
		}

		entitySyncInfosToGate = map[uint16]*netutil.Packet{} // clear all packets
	}
}

func (e *Entity) getSyncInfo() proto.EntitySyncInfo {
	return proto.EntitySyncInfo{
		float32(e.Position.X),
		float32(e.Position.Y),
		float32(e.Position.Z),
		float32(e.yaw),
	}
}

// GetYaw gets entity Yaw
func (e *Entity) GetYaw() Yaw {
	return e.yaw
}

// SetYaw sets entity Yaw
func (e *Entity) SetYaw(yaw Yaw) {
	e.yaw = yaw
	e.syncInfoFlag |= sifSyncNeighborClients | sifSyncOwnClient
	//e.ForAllClients(func(Client *GameClient) {
	//	Client.updateYawOnClient(e.ID, e.Yaw)
	//})
}

// FaceTo let entity face to another entity by setting Yaw accordingly
func (e *Entity) FaceTo(other *Entity) {
	e.FaceToPos(other.Position)
}

// FaceTo let entity face to a specified position, setting Yaw accordingly

func (e *Entity) FaceToPos(pos Vector3) {
	dir := pos.Sub(e.Position)
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
