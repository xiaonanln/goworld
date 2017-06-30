package entity

import (
	"fmt"
	"reflect"

	"time"

	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/typeconv"

	"unsafe"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
	"github.com/xiaonanln/goworld/storage"
)

var (
	saveInterval time.Duration
)

type Entity struct {
	ID       EntityID
	TypeName string
	I        IEntity
	IV       reflect.Value

	destroyed        bool
	rpcDescMap       RpcDescMap
	Space            *Space
	aoi              AOI
	timers           map[*timer.Timer]struct{}
	client           *GameClient
	declaredServices StringSet
	becamePlayer     bool

	Attrs *MapAttr

	enterSpaceRequest struct {
		SpaceID     EntityID
		RequestTime int64
	}
	filterProps map[string]string
}

// Functions declared by IEntity can be override in Entity subclasses
type IEntity interface {
	// Entity Lifetime
	OnInit()
	OnCreated()
	OnDestroy()
	// Migration
	OnMigrateOut()
	OnMigrateIn()
	// Space Operations
	OnEnterSpace()
	OnLeaveSpace(space *Space)
	// Storage: Save & Load
	IsPersistent() bool
	GetPersistentData() map[string]interface{}
	LoadPersistentData(data map[string]interface{})
	// Client Notifications
	OnClientConnected()
	OnClientDisconnected()
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *Entity) Destroy() {
	if e.destroyed {
		return
	}
	gwlog.Info("%s.Destroy ...", e)
	//isCrossServerCallable := e.isCrossServerCallable()
	e.destroyEntity(false)
	//if isCrossServerCallable {
	dispatcher_client.GetDispatcherClientForSend().SendNotifyDestroyEntity(e.ID)
	//}
}

func (e *Entity) destroyEntity(isMigrate bool) {
	e.Space.leave(e)
	e.clearTimers()
	e.timers = nil // prohibit further use

	if !isMigrate {
		gwutils.RunPanicless(e.I.OnDestroy)
		e.SetClient(nil) // always set client to nil before destroy
		e.Save()
	} else {
		gwutils.RunPanicless(e.I.OnMigrateOut)
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

	e.rpcDescMap = entityType2RpcDescMap[typeName]

	e.timers = map[*timer.Timer]struct{}{}
	e.declaredServices = StringSet{}
	e.filterProps = map[string]string{}

	attrs := NewMapAttr()
	attrs.owner = e
	e.Attrs = attrs

	initAOI(&e.aoi)
	gwutils.RunPanicless(e.I.OnInit)
}

func (e *Entity) setupSaveTimer() {
	e.AddTimer(saveInterval, e.Save)
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

// Timer & Callback Management TODO: better management, maybe use Integer to indicate timer ?
func (e *Entity) AddCallback(d time.Duration, cb timer.CallbackFunc) *timer.Timer {
	var t *timer.Timer
	t = timer.AddCallback(d, func() {
		delete(e.timers, t)
		cb()
	})
	e.timers[t] = struct{}{}
	return t
}

func (e *Entity) Post(cb timer.CallbackFunc) {
	e.AddCallback(0, cb)
}

func (e *Entity) AddTimer(d time.Duration, cb timer.CallbackFunc) *timer.Timer {
	t := timer.AddTimer(d, cb)
	e.timers[t] = struct{}{}
	return t
}

func (e *Entity) CancelTimer(t *timer.Timer) {
	delete(e.timers, t)
	t.Cancel()
}

func (e *Entity) clearTimers() {
	for t := range e.timers {
		t.Cancel()
	}
	e.timers = map[*timer.Timer]struct{}{}
}

// Call other entities
func (e *Entity) Call(id EntityID, method string, args ...interface{}) {
	callRemote(id, method, args)
}

func (e *Entity) CallService(serviceName string, method string, args ...interface{}) {
	serviceEid := entityManager.chooseServiceProvider(serviceName)
	callRemote(serviceEid, method, args)
}

func (e *Entity) onCall(methodName string, args []interface{}, clientid ClientID) {
	defer func() {
		err := recover() // recover from any error during RPC call
		if err != nil {
			gwlog.TraceError("%s.%s%v paniced: %s", e, methodName, args, err)
		}
	}()

	rpcDesc := e.rpcDescMap[methodName]
	if rpcDesc == nil {
		// rpc not found
		gwlog.Error("%s.onCall: Method %s is not a valid RPC, args=%v", e, methodName, args)
		return
	}

	methodType := rpcDesc.MethodType
	if clientid == "" {
		// rpc call from server
		if rpcDesc.Flags&RF_SERVER == 0 {
			// can not call from server
			gwlog.Panicf("%s.onCall: Method %s can not be called from Server: flags=%v", e, methodName, rpcDesc.Flags)
		}
	} else {
		isFromOwnClient := clientid == e.getClientID()
		if rpcDesc.Flags&RF_OWN_CLIENT == 0 && isFromOwnClient {
			gwlog.Panicf("%s.onCall: Method %s can not be called from OwnClient: flags=%v", e, methodName, rpcDesc.Flags)
		} else if rpcDesc.Flags&RF_OTHER_CLIENT == 0 && !isFromOwnClient {
			gwlog.Panicf("%s.onCall: Method %s can not be called from OtherClient: flags=%v, OwnClient=%s, OtherClient=%s", e, methodName, rpcDesc.Flags, e.getClientID(), clientid)
		}
	}

	if rpcDesc.NumArgs != len(args) {
		gwlog.Error("%s.onCall: Method %s receives %d arguments, but given %d: %v", e, methodName, rpcDesc.NumArgs, len(args), args)
		return
	}

	in := make([]reflect.Value, len(args)+1)
	in[0] = reflect.ValueOf(e.I)

	for i, arg := range args {
		argType := methodType.In(i + 1)
		in[i+1] = typeconv.Convert(arg, argType)
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

func (e *Entity) IsPersistent() bool {
	return false
}

func (e *Entity) GetPersistentData() map[string]interface{} {
	return e.Attrs.ToMap()
}

func (e *Entity) LoadPersistentData(data map[string]interface{}) {
	e.Attrs.AssignMap(data)
}

func (e *Entity) getClientData() map[string]interface{} {
	return e.Attrs.ToMap() // TODO: only returns client data
}

func (e *Entity) getMigrateData() map[string]interface{} {
	return e.Attrs.ToMap() // TODO: return all data (client, all_client, server, etc)
}

//func (e *Entity) isCrossServerCallable() bool {
//	// find a better rule to determine if entity is cross server callable
//	if e.IsPersistent() {
//		return true
//	}
//	if e.IsSpaceEntity() && !e.ToSpace().IsNil() {
//		return true
//	}
//	if len(e.declaredServices) > 0 {
//		return true
//	}
//	return false
//}

// Client related utilities
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
		dispatcher_client.GetDispatcherClientForSend().SendClearClientFilterProp(oldClient.serverid, oldClient.clientid)

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
			dispatcher_client.GetDispatcherClientForSend().SendSetClientFilterProp(client.serverid, client.clientid, key, val)
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

func (e *Entity) notifyClientDisconnected() {
	if e == nil {
		// FIXME: might happen due to a bug
		return
	}

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
func (e *Entity) EnterSpace(spaceID EntityID) {
	//space := spaceManager.getSpace(spaceID)
	//if space != nil {
	//	// space on the same server
	//	e.Space.leave(e)
	//	space.enter(e)
	//	return
	//}

	if e.isEnteringSpace() {
		gwlog.Error("%s is entering space %s, can not enter space %s", e, e.enterSpaceRequest.SpaceID, spaceID)
		return
	}

	e.requestMigrateTo(spaceID)
}

func (e *Entity) isEnteringSpace() bool {
	now := time.Now().UnixNano()
	return now < (e.enterSpaceRequest.RequestTime + int64(consts.ENTER_SPACE_REQUEST_TIMEOUT))
}

func (e *Entity) clearMigrateRequest() {
	e.enterSpaceRequest.SpaceID = ""
	e.enterSpaceRequest.RequestTime = 0
}

// Migrate to the server of space
func (e *Entity) requestMigrateTo(spaceID EntityID) {
	e.enterSpaceRequest.SpaceID = spaceID
	e.enterSpaceRequest.RequestTime = time.Now().UnixNano()
	dispatcher_client.GetDispatcherClientForSend().SendMigrateRequest(spaceID, e.ID)
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
		entity.clearMigrateRequest()
		return
	}

	if !entity.isEnteringSpace() {
		// replay from dispatcher is too late ?
		return
	}

	entity.realMigrateTo(spaceID, spaceLoc)
}

func (e *Entity) realMigrateTo(spaceID EntityID, spaceLoc uint16) {
	var clientid ClientID
	var clientsrv uint16
	if e.client != nil {
		clientid = e.client.clientid
		clientsrv = e.client.serverid
	}

	e.destroyEntity(true) // disable the entity
	migrateData := e.getMigrateData()

	dispatcher_client.GetDispatcherClientForSend().SendRealMigrate(e.ID, spaceLoc, spaceID, e.TypeName, migrateData, clientid, clientsrv)
}

func OnRealMigrate(entityID EntityID, spaceID EntityID, typeName string, migrateData map[string]interface{}, clientid ClientID, clientsrv uint16) {
	if entityManager.get(entityID) != nil {
		gwlog.Panicf("entity %s already exists", entityID)
	}

	// try to find the target space, but might be nil
	space := spaceManager.getSpace(spaceID)
	var client *GameClient
	if !clientid.IsNil() {
		client = MakeGameClient(clientid, clientsrv)
	}
	createEntity(typeName, space, entityID, migrateData, client, true)
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
		dispatcher_client.GetDispatcherClientForSend().SendSetClientFilterProp(e.client.serverid, e.client.clientid, key, val)
	}
}

func (e *Entity) CallFitleredClients(key string, value interface{}, method string, args ...interface{}) {

}

func (e *Entity) CallFilteredEntities(key string, value interface{}, method string, args ...interface{}) {

}

// Some Other Useful Utilities

func (e *Entity) PanicOnError(err error) {
	if err != nil {
		gwlog.Panic(err)
	}
}
