package entity

import (
	"fmt"
	"reflect"

	"time"

	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/typeconv"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity/attrs"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage"
)

type Entity struct {
	ID       EntityID
	TypeName string
	I        IEntity
	IV       reflect.Value

	destroyed        bool
	rpcDescMap       RpcDescMap
	space            *Space
	aoi              AOI
	timers           map[*timer.Timer]struct{}
	client           *GameClient
	declaredServices StringSet

	Attrs *attrs.MapAttr
}

// Functions declared by IEntity can be override in Entity subclasses
type IEntity interface {
	// Entity Lifetime
	OnInit()
	OnCreated()
	OnDestroy()
	// Space Operations
	OnEnterSpace()
	OnLeaveSpace(space *Space)
	// Storage: Save & Load
	IsPersistent() bool
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

	gwlog.Info("%s.Destroy.", e)
	if e.space != nil {
		e.space.leave(e)
	}
	e.clearTimers()
	if e.isCrossServerCallable() {
		dispatcher_client.GetDispatcherClientForSend().SendNotifyDestroyEntity(e.ID)
	}

	defer func() { // make sure always run after OnDestroy
		entityManager.del(e.ID)
		e.destroyed = true
	}()
	e.I.OnDestroy()
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

	data := e.GetPersistentData()

	storage.Save(e.TypeName, e.ID, data)
}

func (e *Entity) init(typeName string, entityID EntityID, entityPtrVal reflect.Value) {
	e.ID = entityID
	e.IV = entityPtrVal
	e.I = entityPtrVal.Interface().(IEntity)
	e.TypeName = typeName

	e.rpcDescMap = entityType2RpcDescMap[typeName]

	e.timers = map[*timer.Timer]struct{}{}
	e.declaredServices = StringSet{}
	e.Attrs = attrs.NewMapAttr()

	initAOI(&e.aoi)
	e.I.OnInit()

}

func (e *Entity) setupSaveTimer() {
	e.AddTimer(consts.SAVE_INTERVAL, e.Save)
}

// Space Operations related to entity

// Interests and Uninterest among entities
func (e *Entity) interest(other *Entity) {
	e.aoi.interest(other)
	e.client.SendCreateEntity(other)
}

func (e *Entity) uninterest(other *Entity) {
	e.aoi.uninterest(other)
	e.client.SendDestroyEntity(other)
}

func (e *Entity) Neighbors() EntitySet {
	return e.aoi.neighbors
}

// Timer & Callback Management
func (e *Entity) AddCallback(d time.Duration, cb timer.CallbackFunc) {
	var t *timer.Timer
	t = timer.AddCallback(d, func() {
		delete(e.timers, t)
		cb()
	})
	e.timers[t] = struct{}{}
}

func (e *Entity) Post(cb timer.CallbackFunc) {
	e.AddCallback(0, cb)
}

func (e *Entity) AddTimer(d time.Duration, cb timer.CallbackFunc) {
	t := timer.AddTimer(d, cb)
	e.timers[t] = struct{}{}
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
		gwlog.Error("%s.onCall: Method %s is not a valid RPC", e, methodName)
		return
	}

	methodType := rpcDesc.MethodType
	if clientid == "" {
		// rpc call from server
		if rpcDesc.Flags&RF_SERVER == 0 {
			// can not call from server
			gwlog.Error("%s.onCall: Method %s can not be called from Server: flags=%v", e, methodName, rpcDesc.Flags)
			return
		}
	} else {
		isFromOwnClient := clientid == e.getClientID()
		if rpcDesc.Flags&RF_OWN_CLIENT == 0 && isFromOwnClient {
			gwlog.Error("%s.onCall: Method %s can not be called from OwnClient: flags=%v", e, methodName, rpcDesc.Flags)
		} else if rpcDesc.Flags&RF_OTHER_CLIENT == 0 && !isFromOwnClient {
			gwlog.Error("%s.onCall: Method %s can not be called from OtherClient: flags=%v", e, methodName, rpcDesc.Flags)
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
	gwlog.Warn("%s.OnInit not implemented", e)
}

func (e *Entity) OnCreated() {
	gwlog.Debug("%s.OnCreated", e)
}

// Space Utilities
func (e *Entity) GetSpace() *Space {
	return e.space
}

func (e *Entity) OnEnterSpace() {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnEnterSpace >>> %s", e, e.space)
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
	return e.Attrs.ToMap()
}

func (e *Entity) isCrossServerCallable() bool {
	return e.IsPersistent() || len(e.declaredServices) > 0
}

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
		oldClient.SendDestroyEntity(e)
	}

	if client != nil {
		// send create entity to new client
		client.SendCreateEntity(e)
	}

	if oldClient == nil && client != nil {
		// got net client
		e.I.OnClientConnected()
	} else if oldClient != nil && client == nil {
		e.I.OnClientDisconnected()
	}
}

func (e *Entity) CallClient(method string, args ...interface{}) {
	e.client.Call(method, args...)
}

func (e *Entity) GiveClientTo(other *Entity) {
	if e.client == nil {
		return
	}

	client := e.client
	e.SetClient(nil)

	if other.client != nil {
		other.SetClient(nil)
	}

	other.SetClient(client)
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
